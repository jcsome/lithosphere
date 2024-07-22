package processors

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/mantlenetworkio/lithosphere/common/bigint"
	"github.com/mantlenetworkio/lithosphere/common/tasks"
	"github.com/mantlenetworkio/lithosphere/config"
	"github.com/mantlenetworkio/lithosphere/database"
	common2 "github.com/mantlenetworkio/lithosphere/database/common"
	"github.com/mantlenetworkio/lithosphere/event/processors/bridge"
	mantle_da "github.com/mantlenetworkio/lithosphere/event/processors/mantle-da"
	"github.com/mantlenetworkio/lithosphere/event/processors/stateroot"
	"github.com/mantlenetworkio/lithosphere/synchronizer"
)

var blocksLimit = 10_000

type EventProcessor struct {
	log                         log.Logger
	db                          *database.DB
	metrics                     bridge.Metricer
	resourceCtx                 context.Context
	resourceCancel              context.CancelFunc
	tasks                       tasks.Group
	l1Sync                      *synchronizer.L1Sync
	l2Sync                      *synchronizer.L2Sync
	chainConfig                 config.ChainConfig
	LatestL1L2InitL1Header      *common2.L1BlockHeader
	LatestL1L2L2Header          *common2.L2BlockHeader
	LatestL1L2FinalizedL2Header *common2.L2BlockHeader
	LatestStateRootL1Header     *common2.L1BlockHeader
	LatestMantleDAL1Header      *common2.L1BlockHeader
	LatestL2L1InitL2Header      *common2.L2BlockHeader
	LatestProvenL1Header        *common2.L1BlockHeader
	LatestFinalizedL1Header     *common2.L1BlockHeader
}

func NewBridgeProcessor(log log.Logger, db *database.DB, metrics bridge.Metricer, l1Sync *synchronizer.L1Sync, l2Sync *synchronizer.L2Sync,
	chainConfig config.ChainConfig, shutdown context.CancelCauseFunc) (*EventProcessor, error) {
	log = log.New("processor", "bridge")
	latestL1L2InitL1Header, err := db.L1ToL2.L1L2LatestL1BlockHeader()
	if err != nil {
		return nil, err
	}
	latestL1L2L2Header, err := db.L1ToL2.L1L2LatestL2BlockHeader()
	if err != nil {
		return nil, err
	}
	latestL1L2FinalizedL2Header, err := db.L1ToL2.L1L2LatestFinalizedL2BlockHeader()
	if err != nil {
		return nil, err
	}
	latestStateRootL1Header, err := db.StateRoots.StateRootL1BlockHeader()
	if err != nil {
		return nil, err
	}
	latestMantleDAL1Header, err := db.DataStore.DataStoreL1BlockHeader()
	if err != nil {
		return nil, err
	}
	latestL2L1InitL2Header, err := db.L2ToL1.L2L1LatestBlockL2Header()
	if err != nil {
		return nil, err
	}
	latestProvenL1Header, err := db.WithdrawProven.WithdrawProvenL1BlockHeader()
	if err != nil {
		return nil, err
	}
	latestFinalizedL1Header, err := db.WithdrawFinalized.WithdrawFinalizedL1BlockHeader()
	if err != nil {
		return nil, err
	}
	resCtx, resCancel := context.WithCancel(context.Background())
	return &EventProcessor{
		log:            log,
		db:             db,
		metrics:        metrics,
		l1Sync:         l1Sync,
		l2Sync:         l2Sync,
		resourceCtx:    resCtx,
		resourceCancel: resCancel,
		chainConfig:    chainConfig,
		tasks: tasks.Group{HandleCrit: func(err error) {
			shutdown(fmt.Errorf("critical error in bridge processor: %w", err))
		}},
		LatestL1L2InitL1Header:      latestL1L2InitL1Header,
		LatestL1L2L2Header:          latestL1L2L2Header,
		LatestL1L2FinalizedL2Header: latestL1L2FinalizedL2Header,
		LatestStateRootL1Header:     latestStateRootL1Header,
		LatestMantleDAL1Header:      latestMantleDAL1Header,
		LatestL2L1InitL2Header:      latestL2L1InitL2Header,
		LatestProvenL1Header:        latestProvenL1Header,
		LatestFinalizedL1Header:     latestFinalizedL1Header,
	}, nil
}

func (ep *EventProcessor) Start() error {
	ep.log.Info("starting bridge processor...")
	tickerL1Worker := time.NewTicker(time.Second * 5)
	ep.tasks.Go(func() error {
		for range tickerL1Worker.C {
			done := ep.metrics.RecordL1Interval()
			done(ep.onL1Data())
		}
		return nil
	})

	tickerL2Worker := time.NewTicker(time.Second * 5)
	ep.tasks.Go(func() error {
		for range tickerL2Worker.C {
			done := ep.metrics.RecordL2Interval()
			done(ep.onL2Data())
		}
		return nil
	})
	return nil
}

func (ep *EventProcessor) Close() error {
	ep.resourceCancel()
	return ep.tasks.Wait()
}

func (ep *EventProcessor) onL1Data() error {
	ep.log.Info("start on l1 data")
	var errs error
	if err := ep.processInitiatedL1Events(); err != nil {
		ep.log.Error("failed to process initiated L1 events", "err", err)
		errs = errors.Join(errs, err)
	}

	if err := ep.processFinalizedL2Events(); err != nil {
		ep.log.Error("failed to process finalized L2 events", "err", err)
		errs = errors.Join(errs, err)
	}

	if err := ep.processRollupMantleDA(); err != nil {
		ep.log.Error("failed to process rollup events", "err", err)
		errs = errors.Join(errs, err)
	}

	if err := ep.processRollupStateRoot(); err != nil {
		ep.log.Error("failed to process rollup events", "err", err)
		errs = errors.Join(errs, err)
	}
	return errs
}

func (ep *EventProcessor) onL2Data() error {
	ep.log.Info("start on l2 data")

	var errs error
	if err := ep.processInitiatedL2Events(); err != nil {
		ep.log.Error("failed to process initiated L2 events", "err", err)
		errs = errors.Join(errs, err)
	}

	if err := ep.processProvenL1Events(); err != nil {
		ep.log.Error("failed to process proven L1 events", "err", err)
		errs = errors.Join(errs, err)
	}

	if err := ep.processFinalizedL1Events(); err != nil {
		ep.log.Error("failed to process finalized L1 events", "err", err)
		errs = errors.Join(errs, err)
	}
	return errs
}

func (ep *EventProcessor) processInitiatedL1Events() error {
	l1BridgeLog := ep.log.New("bridge", "l1", "kind", "initiated")
	lastL1BlockNumber := big.NewInt(int64(ep.chainConfig.L1StartingHeight))
	if ep.LatestL1L2InitL1Header != nil {
		lastL1BlockNumber = ep.LatestL1L2InitL1Header.Number
	}
	l1BridgeLog.Info("Process init l1 event", "lastL1BlockNumber", lastL1BlockNumber)
	latestL1HeaderScope := func(db *gorm.DB) *gorm.DB {
		newQuery := db.Session(&gorm.Session{NewDB: true})
		headers := newQuery.Model(common2.L1BlockHeader{}).Where("number > ?", lastL1BlockNumber)
		return db.Where("number = (?)", newQuery.Table("(?) as block_numbers", headers.Order("number ASC").Limit(blocksLimit)).Select("MAX(number)"))
	}
	latestL1Header, err := ep.db.Blocks.L1BlockHeaderWithScope(latestL1HeaderScope)
	if err != nil {
		return fmt.Errorf("failed to query new L1 state: %w", err)
	} else if latestL1Header == nil {
		l1BridgeLog.Debug("no new L1 state found")
		return nil
	}
	fromL1Height, toL1Height := new(big.Int).Add(lastL1BlockNumber, bigint.One), latestL1Header.Number
	if err := ep.db.Transaction(func(tx *database.DB) error {
		l1BedrockStartingHeight := big.NewInt(int64(ep.chainConfig.L1BedrockStartingHeight))
		if l1BedrockStartingHeight.Cmp(fromL1Height) > 0 {
			legacyFromL1Height, legacyToL1Height := fromL1Height, toL1Height
			if l1BedrockStartingHeight.Cmp(toL1Height) <= 0 {
				legacyToL1Height = new(big.Int).Sub(l1BedrockStartingHeight, bigint.One)
			}

			legacyBridgeLog := l1BridgeLog.New("mode", "legacy", "from_block_number", legacyFromL1Height, "to_block_number", legacyToL1Height)
			legacyBridgeLog.Info("scanning for initiated bridge events")
			if err := bridge.LegacyL1ProcessInitiatedBridgeEvents(legacyBridgeLog, tx, ep.metrics, ep.chainConfig.L1Contracts, legacyFromL1Height, legacyToL1Height); err != nil {
				return err
			} else if legacyToL1Height.Cmp(toL1Height) == 0 {
				return nil
			}
			legacyBridgeLog.Info("detected switch to bedrock", "bedrock_block_number", l1BedrockStartingHeight)
			fromL1Height = l1BedrockStartingHeight
		}
		l1BridgeLog = l1BridgeLog.New("from_block_number", fromL1Height, "to_block_number", toL1Height)
		l1BridgeLog.Info("scanning for initiated bridge events")
		return bridge.L1ProcessInitiatedBridgeEvents(l1BridgeLog, tx, ep.metrics, ep.chainConfig.L1Contracts, fromL1Height, toL1Height)
	}); err != nil {
		return err
	}
	ep.LatestL1L2InitL1Header = latestL1Header
	ep.metrics.RecordL1LatestHeight(latestL1Header.Number)
	return nil
}

func (ep *EventProcessor) processInitiatedL2Events() error {
	l2BridgeLog := ep.log.New("bridge", "l2", "kind", "initiated")
	lastL2BlockNumber := big.NewInt(int64(ep.chainConfig.L2StartingHeight))
	if ep.LatestL2L1InitL2Header != nil {
		lastL2BlockNumber = ep.LatestL2L1InitL2Header.Number
	}
	latestL2HeaderScope := func(db *gorm.DB) *gorm.DB {
		newQuery := db.Session(&gorm.Session{NewDB: true})
		headers := newQuery.Model(common2.L2BlockHeader{}).Where("number > ?", lastL2BlockNumber)
		return db.Where("number = (?)", newQuery.Table("(?) as block_numbers", headers.Order("number ASC").Limit(blocksLimit)).Select("MAX(number)"))
	}
	latestL2Header, err := ep.db.Blocks.L2BlockHeaderWithScope(latestL2HeaderScope)
	if err != nil {
		return fmt.Errorf("failed to query new L2 state: %w", err)
	} else if latestL2Header == nil {
		l2BridgeLog.Warn("no new L2 state found")
		return nil
	}
	fromL2Height, toL2Height := new(big.Int).Add(lastL2BlockNumber, bigint.One), latestL2Header.Number
	if err := ep.db.Transaction(func(tx *database.DB) error {
		l2BedrockStartingHeight := big.NewInt(int64(ep.chainConfig.L2BedrockStartingHeight))
		if l2BedrockStartingHeight.Cmp(fromL2Height) > 0 { // OP Mainnet & OP Goerli Only
			legacyFromL2Height, legacyToL2Height := fromL2Height, toL2Height
			if l2BedrockStartingHeight.Cmp(toL2Height) <= 0 {
				legacyToL2Height = new(big.Int).Sub(l2BedrockStartingHeight, bigint.One)
			}
			legacyBridgeLog := l2BridgeLog.New("mode", "legacy", "from_block_number", legacyFromL2Height, "to_block_number", legacyToL2Height)
			legacyBridgeLog.Info("scanning for initiated bridge events")
			if err := bridge.LegacyL2ProcessInitiatedBridgeEvents(legacyBridgeLog, tx, ep.metrics, ep.chainConfig.L2Contracts, legacyFromL2Height, legacyToL2Height); err != nil {
				return err
			} else if legacyToL2Height.Cmp(toL2Height) == 0 {
				return nil
			}
			legacyBridgeLog.Info("detected switch to bedrock")
			fromL2Height = l2BedrockStartingHeight
		}
		l2BridgeLog = l2BridgeLog.New("from_block_number", fromL2Height, "to_block_number", toL2Height)
		l2BridgeLog.Info("scanning for initiated bridge events")
		return bridge.L2ProcessInitiatedBridgeEvents(l2BridgeLog, tx, ep.metrics, ep.chainConfig.L2Contracts, fromL2Height, toL2Height)
	}); err != nil {
		return err
	}
	ep.LatestL2L1InitL2Header = latestL2Header
	ep.metrics.RecordL2LatestHeight(latestL2Header.Number)
	return nil
}

func (ep *EventProcessor) processProvenL1Events() error {
	l1BridgeLog := ep.log.New("bridge", "l1", "kind", "proven")
	lastProvenL1BlockNumber := big.NewInt(int64(ep.chainConfig.L1StartingHeight))
	if ep.LatestProvenL1Header != nil {
		lastProvenL1BlockNumber = ep.LatestProvenL1Header.Number
	}
	l1BridgeLog.Info("Process proven l1 event latest block number", "lastProvenL1BlockNumber", lastProvenL1BlockNumber)
	latestProvenL1HeaderScope := func(db *gorm.DB) *gorm.DB {
		newQuery := db.Session(&gorm.Session{NewDB: true})
		headers := newQuery.Model(common2.L1BlockHeader{}).Where("number > ?", lastProvenL1BlockNumber)
		return db.Where("number = (?)", newQuery.Table("(?) as block_numbers", headers.Order("number ASC").Limit(blocksLimit)).Select("MAX(number)"))
	}
	if latestProvenL1HeaderScope == nil {
		return nil
	}
	latestL1Header, err := ep.db.Blocks.L1BlockHeaderWithScope(latestProvenL1HeaderScope)
	if err != nil {
		return fmt.Errorf("failed to query for latest unfinalized L1 state: %w", err)
	} else if latestL1Header == nil {
		l1BridgeLog.Debug("no new l1 state to proven")
		return nil
	}
	fromL1Height, toL1Height := new(big.Int).Add(lastProvenL1BlockNumber, bigint.One), latestL1Header.Number
	if err := ep.db.Transaction(func(tx *database.DB) error {
		l1BridgeLog = l1BridgeLog.New("from_block_number", fromL1Height, "to_block_number", toL1Height)
		l1BridgeLog.Info("scanning for withdraw proven events")
		if err := bridge.L1ProcessProvenBridgeEvents(l1BridgeLog, tx, ep.metrics, ep.chainConfig.L1Contracts, fromL1Height, toL1Height); err != nil {
			ep.log.Error("failed to index withdraw proven events", "err", err)
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	ep.LatestProvenL1Header = latestL1Header
	ep.metrics.RecordL1LatestProvenHeight(lastProvenL1BlockNumber)
	return nil
}

func (ep *EventProcessor) processFinalizedL1Events() error {
	l1BridgeLog := ep.log.New("bridge", "l1", "kind", "finalization")
	lastFinalizedL1BlockNumber := big.NewInt(int64(ep.chainConfig.L1StartingHeight))
	if ep.LatestFinalizedL1Header != nil {
		lastFinalizedL1BlockNumber = ep.LatestFinalizedL1Header.Number
	}
	l1BridgeLog.Info("Process finalized l1 event l1 header scope", "lastFinalizedL1BlockNumber", lastFinalizedL1BlockNumber)
	latestFinalizedL1HeaderScope := func(db *gorm.DB) *gorm.DB {
		newQuery := db.Session(&gorm.Session{NewDB: true})
		headers := newQuery.Model(common2.L1BlockHeader{}).Where("number > ?", lastFinalizedL1BlockNumber)
		return db.Where("number = (?)", newQuery.Table("(?) as block_numbers", headers.Order("number ASC").Limit(blocksLimit)).Select("MAX(number)"))
	}
	if latestFinalizedL1HeaderScope == nil {
		return nil
	}
	latestL1Header, err := ep.db.Blocks.L1BlockHeaderWithScope(latestFinalizedL1HeaderScope)
	if err != nil {
		return fmt.Errorf("failed to query for latest unfinalized L1 state: %w", err)
	} else if latestL1Header == nil {
		l1BridgeLog.Debug("no new l1 state to finalize", "last_finalized_block_number", lastFinalizedL1BlockNumber)
		return nil
	}
	fromL1Height, toL1Height := new(big.Int).Add(lastFinalizedL1BlockNumber, bigint.One), latestL1Header.Number
	if err := ep.db.Transaction(func(tx *database.DB) error {
		l1BedrockStartingHeight := big.NewInt(int64(ep.chainConfig.L1BedrockStartingHeight))
		if l1BedrockStartingHeight.Cmp(fromL1Height) > 0 {
			legacyFromL1Height, legacyToL1Height := fromL1Height, toL1Height
			if l1BedrockStartingHeight.Cmp(toL1Height) <= 0 {
				legacyToL1Height = new(big.Int).Sub(l1BedrockStartingHeight, bigint.One)
			}
			legacyBridgeLog := l1BridgeLog.New("mode", "legacy", "from_block_number", legacyFromL1Height, "to_block_number", legacyToL1Height)
			legacyBridgeLog.Info("scanning for finalized bridge events")
			if err := bridge.LegacyL1ProcessFinalizedBridgeEvents(legacyBridgeLog, tx, ep.metrics, ep.l1Sync.EthClient, ep.chainConfig.L1Contracts, legacyFromL1Height, legacyToL1Height); err != nil {
				return err
			} else if legacyToL1Height.Cmp(toL1Height) == 0 {
				return nil
			}
			legacyBridgeLog.Info("detected switch to bedrock")
			fromL1Height = l1BedrockStartingHeight
		}
		l1BridgeLog = l1BridgeLog.New("from_block_number", fromL1Height, "to_block_number", toL1Height)
		l1BridgeLog.Info("scanning for finalized bridge events")
		return bridge.L1ProcessFinalizedBridgeEvents(l1BridgeLog, tx, ep.metrics, ep.chainConfig.L1Contracts, fromL1Height, toL1Height)
	}); err != nil {
		return err
	}
	ep.LatestFinalizedL1Header = latestL1Header
	ep.metrics.RecordL1LatestFinalizedHeight(lastFinalizedL1BlockNumber)
	return nil
}

func (ep *EventProcessor) processFinalizedL2Events() error {
	l2BridgeLog := ep.log.New("bridge", "l2", "kind", "finalization")
	lastFinalizedL2BlockNumber := big.NewInt(int64(ep.chainConfig.L2StartingHeight))
	latestL1L2BlockNumber := big.NewInt(int64(ep.chainConfig.L2StartingHeight))
	if ep.LatestL1L2FinalizedL2Header != nil {
		lastFinalizedL2BlockNumber = ep.LatestL1L2FinalizedL2Header.Number
	}
	l2BlockHeader, err := ep.db.L1ToL2.L1L2LatestL2BlockHeader()
	if err != nil {
		ep.log.Error("get latest l2 block header fail", "err", err)
		return err
	}
	if l2BlockHeader != nil {
		ep.LatestL1L2L2Header = l2BlockHeader
		latestL1L2BlockNumber = l2BlockHeader.Number
	}
	l2BridgeLog.Info("Process finalized l2 event l2 header scope", "start", lastFinalizedL2BlockNumber, "end", latestL1L2BlockNumber)
	latestL2HeaderScope := func(db *gorm.DB) *gorm.DB {
		newQuery := db.Session(&gorm.Session{NewDB: true})
		headers := newQuery.Model(common2.L2BlockHeader{}).Where("number > ? AND number <= ?", lastFinalizedL2BlockNumber, latestL1L2BlockNumber)
		return db.Where("number = (?)", newQuery.Table("(?) as block_numbers", headers.Order("number ASC").Limit(blocksLimit)).Select("MAX(number)"))
	}
	latestL2Header, err := ep.db.Blocks.L2BlockHeaderWithScope(latestL2HeaderScope)
	if err != nil {
		return fmt.Errorf("failed to query for latest unfinalized L2 state: %w", err)
	} else if latestL2Header == nil {
		l2BridgeLog.Debug("no new l2 state to finalize", "last_finalized_block_number", lastFinalizedL2BlockNumber)
		ep.LatestL1L2FinalizedL2Header = latestL2Header
		return nil
	}
	log.Info("latest l2 header", "latestL2Header", latestL2Header.Number)
	fromL2Height, toL2Height := new(big.Int).Add(lastFinalizedL2BlockNumber, bigint.One), latestL2Header.Number
	if err := ep.db.Transaction(func(tx *database.DB) error {
		l2BedrockStartingHeight := big.NewInt(int64(ep.chainConfig.L2BedrockStartingHeight))
		if l2BedrockStartingHeight.Cmp(fromL2Height) > 0 {
			legacyFromL2Height, legacyToL2Height := fromL2Height, toL2Height
			if l2BedrockStartingHeight.Cmp(toL2Height) <= 0 {
				legacyToL2Height = new(big.Int).Sub(l2BedrockStartingHeight, bigint.One)
			}
			legacyBridgeLog := l2BridgeLog.New("mode", "legacy", "from_block_number", legacyFromL2Height, "to_block_number", legacyToL2Height)
			legacyBridgeLog.Info("scanning for finalized bridge events")
			if err := bridge.LegacyL2ProcessFinalizedBridgeEvents(legacyBridgeLog, tx, ep.metrics, ep.chainConfig.L2Contracts, legacyFromL2Height, legacyToL2Height); err != nil {
				return err
			} else if legacyToL2Height.Cmp(toL2Height) == 0 {
				return nil
			}
			legacyBridgeLog.Info("detected switch to bedrock", "bedrock_block_number", l2BedrockStartingHeight)
			fromL2Height = l2BedrockStartingHeight
		}

		l2BridgeLog = l2BridgeLog.New("from_block_number", fromL2Height, "to_block_number", toL2Height)
		l2BridgeLog.Info("scanning for finalized bridge events")
		return bridge.L2ProcessFinalizedBridgeEvents(l2BridgeLog, tx, ep.metrics, ep.chainConfig.L2Contracts, fromL2Height, toL2Height)
	}); err != nil {
		return err
	}
	ep.LatestL1L2FinalizedL2Header = latestL2Header
	ep.metrics.RecordL2LatestFinalizedHeight(lastFinalizedL2BlockNumber)
	return nil
}

func (ep *EventProcessor) processRollupStateRoot() error {
	rollupStateRootLog := ep.log.New("rollup", "l1", "kind", "state root")
	lastStateRootL1BlockNumber := big.NewInt(int64(ep.chainConfig.L1StartingHeight))
	if ep.LatestStateRootL1Header != nil {
		lastStateRootL1BlockNumber = ep.LatestStateRootL1Header.Number
	}
	latestRollupL1HeaderScope := func(db *gorm.DB) *gorm.DB {
		newQuery := db.Session(&gorm.Session{NewDB: true})
		headers := newQuery.Model(common2.L1BlockHeader{}).Where("number > ?", lastStateRootL1BlockNumber)
		return db.Where("number = (?)", newQuery.Table("(?) as block_numbers", headers.Order("number ASC").Limit(blocksLimit)).Select("MAX(number)"))
	}
	latestL1StateRootHeader, err := ep.db.Blocks.L1BlockHeaderWithScope(latestRollupL1HeaderScope)
	if err != nil {
		return fmt.Errorf("failed to query new L1 state: %w", err)
	} else if latestL1StateRootHeader == nil {
		rollupStateRootLog.Debug("no new L1 state found for process rollup")
		return nil
	}
	fromL1Height, toL1Height := new(big.Int).Add(lastStateRootL1BlockNumber, bigint.One), latestL1StateRootHeader.Number
	if err := ep.db.Transaction(func(tx *database.DB) error {
		rollupStateRootLog = rollupStateRootLog.New("from_block_number", fromL1Height, "to_block_number", toL1Height)
		rollupStateRootLog.Info("scanning for state root events")
		if err := stateroot.L2OutputEvent(rollupStateRootLog, tx, ep.metrics, ep.chainConfig.L1Contracts, fromL1Height, toL1Height); err != nil {
			ep.log.Error("failed to index l1 l2output proposed events", "err", err)
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	ep.LatestStateRootL1Header = latestL1StateRootHeader
	ep.metrics.RecordL1LatestRollupSateRootHeight(lastStateRootL1BlockNumber)
	return nil
}

func (ep *EventProcessor) processRollupMantleDA() error {
	rollupMantleDaLog := ep.log.New("rollup", "l1", "kind", "mantleDa")
	lastRollupMantleDaL1BlockNumber := big.NewInt(int64(ep.chainConfig.L1StartingHeight))
	if ep.LatestMantleDAL1Header != nil {
		lastRollupMantleDaL1BlockNumber = ep.LatestMantleDAL1Header.Number
	}

	latestRollupL1HeaderScope := func(db *gorm.DB) *gorm.DB {
		newQuery := db.Session(&gorm.Session{NewDB: true})
		headers := newQuery.Model(common2.L1BlockHeader{}).Where("number > ?", lastRollupMantleDaL1BlockNumber)
		return db.Where("number = (?)", newQuery.Table("(?) as block_numbers", headers.Order("number ASC").Limit(blocksLimit)).Select("MAX(number)"))
	}
	latestL1RollupMantleDaHeader, err := ep.db.Blocks.L1BlockHeaderWithScope(latestRollupL1HeaderScope)
	if err != nil {
		return fmt.Errorf("failed to query new L1 state: %w", err)
	} else if latestL1RollupMantleDaHeader == nil {
		rollupMantleDaLog.Debug("no new L1 state found for process rollup")
		return nil
	}
	fromL1Height, toL1Height := new(big.Int).Add(lastRollupMantleDaL1BlockNumber, bigint.One), latestL1RollupMantleDaHeader.Number
	if err := ep.db.Transaction(func(tx *database.DB) error {
		rollupMantleDaLog = rollupMantleDaLog.New("from_block_number", fromL1Height, "to_block_number", toL1Height)
		rollupMantleDaLog.Info("scanning for mantle da events")
		if err := mantle_da.L1ProcessMantleDAEvents(rollupMantleDaLog, tx, ep.chainConfig.L1Contracts, fromL1Height, toL1Height); err != nil {
			ep.log.Error("failed to index l1 mantle da events", "err", err)
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	ep.LatestMantleDAL1Header = latestL1RollupMantleDaHeader
	ep.metrics.RecordL1LatestRollupMantleDaHeight(lastRollupMantleDaL1BlockNumber)
	return nil
}
