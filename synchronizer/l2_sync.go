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
	common1 "github.com/mantlenetworkio/lithosphere/database/common"
	"github.com/mantlenetworkio/lithosphere/database/event"
	"github.com/mantlenetworkio/lithosphere/metrics"
	"github.com/mantlenetworkio/lithosphere/synchronizer/node"
	"github.com/mantlenetworkio/lithosphere/synchronizer/retry"
)

type L2Sync struct {
	Synchronizer
	LatestHeader   *types.Header
	resourceCtx    context.Context
	resourceCancel context.CancelFunc
	tasks          tasks.Group
	db             *database.DB
}

func NewL2Sync(cfg Config, log log.Logger, db *database.DB, metrics metrics.Metricer, client node.EthClient,
	contracts config.L2Contracts, shutdown context.CancelCauseFunc, transferBigValueInMantle string) (*L2Sync, error) {
	log = log.New("syncer", "l2")
	zeroAddr := common.Address{}
	l2Contracts := []common.Address{}
	if err := contracts.ForEach(func(name string, addr common.Address) error {
		if addr == zeroAddr {
			log.Error("address not configured", "name", name)
			return errors.New("all L2Contracts must be configured")
		}
		log.Info("configured l2 contract", "name", name, "addr", addr)
		l2Contracts = append(l2Contracts, addr)
		return nil
	}); err != nil {
		return nil, err
	}
	transferBigValueAddresses := strings.Split(transferBigValueInMantle, " ")
	for _, bigValueAddress := range transferBigValueAddresses {
		address := common.HexToAddress(bigValueAddress)
		l2Contracts = append(l2Contracts, address)
	}

	latestHeader, err := db.Blocks.L2LatestBlockHeader()
	if err != nil {
		return nil, err
	}

	var fromHeader *types.Header
	if latestHeader != nil {
		log.Info("l2 sync detected last indexed block", "number", latestHeader.Number, "hash", latestHeader.Hash)
		fromHeader = latestHeader.RLPHeader.Header()
	} else if cfg.StartHeight.BitLen() > 0 {
		log.Info("no l2 sync indexed state starting from supplied L2 height", "height", cfg.StartHeight.String())
		header, err := client.BlockHeaderByNumber(cfg.StartHeight)
		if err != nil {
			return nil, fmt.Errorf("could not fetch starting block header: %w", err)
		}
		fromHeader = header
	} else {
		log.Info("no l2 indexed state")
	}

	syncerBatches := make(chan *SynchronizerBatch)
	syncer := Synchronizer{
		loopInterval:     time.Duration(cfg.LoopIntervalMsec) * time.Millisecond,
		headerBufferSize: uint64(cfg.HeaderBufferSize),
		log:              log,
		metrics:          metrics,
		headerTraversal:  node.NewHeaderTraversal(client, fromHeader, cfg.ConfirmationDepth),
		contracts:        l2Contracts,
		syncerBatches:    syncerBatches,
		EthClient:        client,
	}

	resCtx, resCancel := context.WithCancel(context.Background())
	return &L2Sync{
		Synchronizer:   syncer,
		LatestHeader:   fromHeader,
		resourceCtx:    resCtx,
		resourceCancel: resCancel,
		db:             db,
		tasks: tasks.Group{HandleCrit: func(err error) {
			shutdown(fmt.Errorf("critical error in L2 Synchronizer: %w", err))
		}},
	}, nil
}

func (l2Sync *L2Sync) Close() error {
	var result error
	if err := l2Sync.Synchronizer.Close(); err != nil {
		result = errors.Join(result, fmt.Errorf("failed to close internal l2 Synchronizer: %w", err))
	}
	l2Sync.resourceCancel()
	if err := l2Sync.tasks.Wait(); err != nil {
		result = errors.Join(result, fmt.Errorf("failed to await l2 batch handler completion: %w", err))
	}
	return result
}

func (l2Sync *L2Sync) Start() error {
	l2Sync.log.Info("starting l2 synchronizer...")
	if err := l2Sync.Synchronizer.Start(); err != nil {
		return fmt.Errorf("failed to start internal l2 Synchronizer: %w", err)
	}
	l2Sync.tasks.Go(func() error {
		for batch := range l2Sync.syncerBatches {
			if err := l2Sync.handleBatch(batch); err != nil {
				return fmt.Errorf("failed to handle batch, stopping L2 Synchronizer: %w", err)
			}
		}
		return nil
	})
	return nil
}

func (l2Sync *L2Sync) handleBatch(batch *SynchronizerBatch) error {
	l2BlockHeaders := make([]common1.L2BlockHeader, len(batch.Headers))
	var txList []common1.Transactions
	for i := range batch.Headers {
		l2BlockHeaders[i] = common1.L2BlockHeader{BlockHeader: common1.BlockHeaderFromHeader(&batch.Headers[i])}
		tansactionList, err := l2Sync.EthClient.TxsByHash(l2BlockHeaders[i].Hash)
		if err != nil {
			return err
		}
		batch.Logger.Info("start l2 handle batch transaction", "transactions len", tansactionList.Len())
		for j := range tansactionList {
			tx, err := l2Sync.EthClient.TxDetailByHash(tansactionList[j].Hash())
			if err != nil {
				return err
			}
			txReceipt, err := l2Sync.EthClient.TxReceiptDetailByHash(tansactionList[j].Hash())
			if err != nil {
				return err
			}
			transaction, err := l2Sync.db.Transactions.BuildTransactions(tx, txReceipt)
			if err != nil {
				return err
			}
			txList = append(txList, transaction)
		}
	}
	l2ContractEvents := make([]event.L2ContractEvent, len(batch.Logs))
	for i := range batch.Logs {
		timestamp := batch.HeaderMap[batch.Logs[i].BlockHash].Time
		l2ContractEvents[i] = event.L2ContractEvent{ContractEvent: event.ContractEventFromLog(&batch.Logs[i], timestamp)}
		l2Sync.Synchronizer.metrics.RecordBatchLog(batch.Logs[i].Address)
	}
	retryStrategy := &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
	if _, err := retry.Do[interface{}](l2Sync.resourceCtx, 10, retryStrategy, func() (interface{}, error) {
		if err := l2Sync.db.Transaction(func(tx *database.DB) error {
			if err := tx.Blocks.StoreL2BlockHeaders(l2BlockHeaders); err != nil {
				return err
			}
			if len(l2ContractEvents) > 0 {
				if err := tx.ContractEvents.StoreL2ContractEvents(l2ContractEvents); err != nil {
					return err
				}
			}
			if len(txList) > 0 {
				if err := tx.Transactions.StoreTransactions(txList); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			batch.Logger.Error("unable to persist l2 batch", "err", err)
			return nil, err
		}
		l2Sync.Synchronizer.metrics.RecordIndexedHeaders(len(l2BlockHeaders))
		l2Sync.Synchronizer.metrics.RecordIndexedLatestHeight(l2BlockHeaders[len(l2BlockHeaders)-1].Number)
		return nil, nil
	}); err != nil {
		return err
	}
	batch.Logger.Info("indexed l2 batch")
	return nil
}
