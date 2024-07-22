package synchronizer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mantlenetworkio/lithosphere/common/tasks"
	"github.com/mantlenetworkio/lithosphere/config"
	"github.com/mantlenetworkio/lithosphere/database"
	common2 "github.com/mantlenetworkio/lithosphere/database/common"
	"github.com/mantlenetworkio/lithosphere/database/event"
	"github.com/mantlenetworkio/lithosphere/metrics"
	"github.com/mantlenetworkio/lithosphere/synchronizer/node"
	"github.com/mantlenetworkio/lithosphere/synchronizer/retry"
)

type L1Sync struct {
	Synchronizer
	LatestHeader   *types.Header
	resourceCtx    context.Context
	resourceCancel context.CancelFunc
	tasks          tasks.Group
	db             *database.DB
}

func NewL1Sync(cfg Config, log log.Logger, db *database.DB, metrics metrics.Metricer, client node.EthClient,
	contracts config.L1Contracts, shutdown context.CancelCauseFunc, transferBigValueInEthereum string) (*L1Sync, error) {
	log = log.New("synchronizer", "l1")
	zeroAddr := common.Address{}
	l1Contracts := []common.Address{}
	if err := contracts.ForEach(func(name string, addr common.Address) error {
		if addr == zeroAddr && !strings.HasPrefix(name, "Legacy") {
			log.Error("address not configured", "name", name)
			return errors.New("all L1Contracts must be configured")
		}
		log.Info("configured contract", "name", name, "addr", addr)
		l1Contracts = append(l1Contracts, addr)
		return nil
	}); err != nil {
		return nil, err
	}
	transferBigValueAddresses := strings.Split(transferBigValueInEthereum, " ")
	for _, bigValueAddress := range transferBigValueAddresses {
		address := common.HexToAddress(bigValueAddress)
		l1Contracts = append(l1Contracts, address)
	}

	latestHeader, err := db.Blocks.L1LatestBlockHeader()
	if err != nil {
		return nil, err
	}
	var fromHeader *types.Header
	if latestHeader != nil {
		log.Info("l1 sync detected last indexed block", "number", latestHeader.Number, "hash", latestHeader.Hash)
		fromHeader = latestHeader.RLPHeader.Header()
	} else if cfg.StartHeight.BitLen() > 0 {
		log.Info("no l2 sync indexed state starting from supplied L1 height", "height", cfg.StartHeight.String())
		header, err := client.BlockHeaderByNumber(cfg.StartHeight)
		if err != nil {
			return nil, fmt.Errorf("could not fetch starting block header: %w", err)
		}
		fromHeader = header
	} else {
		log.Info("no l2 sync indexed state, starting from genesis")
	}
	synchronizerBatches := make(chan *SynchronizerBatch)
	synchronizer := Synchronizer{
		loopInterval:     time.Duration(cfg.LoopIntervalMsec) * time.Millisecond,
		headerBufferSize: uint64(cfg.HeaderBufferSize),
		log:              log,
		metrics:          metrics,
		headerTraversal:  node.NewHeaderTraversal(client, fromHeader, cfg.ConfirmationDepth),
		contracts:        l1Contracts,
		syncerBatches:    synchronizerBatches,
		EthClient:        client,
	}

	resCtx, resCancel := context.WithCancel(context.Background())
	return &L1Sync{
		Synchronizer:   synchronizer,
		LatestHeader:   fromHeader,
		db:             db,
		resourceCtx:    resCtx,
		resourceCancel: resCancel,
		tasks: tasks.Group{HandleCrit: func(err error) {
			shutdown(fmt.Errorf("critical error in L1 Synchronizer: %w", err))
		}},
	}, nil
}

func (l1Sync *L1Sync) Close() error {
	var result error
	if err := l1Sync.Synchronizer.Close(); err != nil {
		result = errors.Join(result, fmt.Errorf("failed to close internal Synchronizer: %w", err))
	}
	l1Sync.resourceCancel()
	if err := l1Sync.tasks.Wait(); err != nil {
		result = errors.Join(result, fmt.Errorf("failed to await batch handler completion: %w", err))
	}
	return result
}

func (l1Sync *L1Sync) Start() error {
	l1Sync.log.Info("starting l1 synchronizer...")
	if err := l1Sync.Synchronizer.Start(); err != nil {
		return fmt.Errorf("failed to start internal Synchronizer: %w", err)
	}
	l1Sync.tasks.Go(func() error {
		for batch := range l1Sync.syncerBatches {
			if err := l1Sync.handleBatch(batch); err != nil {
				return fmt.Errorf("failed to handle batch, stopping L2 Synchronizer: %w", err)
			}
		}
		return nil
	})
	return nil
}

func (l1Sync *L1Sync) handleBatch(batch *SynchronizerBatch) error {
	l1BlockHeaders := make([]common2.L1BlockHeader, 0, len(batch.Headers))
	for i := range batch.Headers {
		if _, ok := batch.HeadersWithLog[batch.Headers[i].Hash()]; ok {
			l1BlockHeaders = append(l1BlockHeaders, common2.L1BlockHeader{BlockHeader: common2.BlockHeaderFromHeader(&batch.Headers[i])})
		}
	}

	if len(l1BlockHeaders) == 0 {
		batch.Logger.Info("no l1 blocks with logs in batch")
		return nil
	}

	l1ContractEvents := make([]event.L1ContractEvent, len(batch.Logs))
	for i := range batch.Logs {
		timestamp := batch.HeaderMap[batch.Logs[i].BlockHash].Time
		l1ContractEvents[i] = event.L1ContractEvent{ContractEvent: event.ContractEventFromLog(&batch.Logs[i], timestamp)}
		l1Sync.Synchronizer.metrics.RecordBatchLog(batch.Logs[i].Address)
	}

	retryStrategy := &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
	if _, err := retry.Do[interface{}](l1Sync.resourceCtx, 10, retryStrategy, func() (interface{}, error) {
		if err := l1Sync.db.Transaction(func(tx *database.DB) error {
			if err := tx.Blocks.StoreL1BlockHeaders(l1BlockHeaders); err != nil {
				return err
			}
			if err := tx.ContractEvents.StoreL1ContractEvents(l1ContractEvents); err != nil {
				return err
			}
			return nil
		}); err != nil {
			batch.Logger.Error("unable to persist batch", "err", err)
			return nil, fmt.Errorf("unable to persist batch: %w", err)
		}
		l1Sync.Synchronizer.metrics.RecordIndexedHeaders(len(l1BlockHeaders))
		l1Sync.Synchronizer.metrics.RecordIndexedLatestHeight(l1BlockHeaders[len(l1BlockHeaders)-1].Number)
		return nil, nil
	}); err != nil {
		return err
	}
	batch.Logger.Info("indexed l1 batch")
	return nil
}
