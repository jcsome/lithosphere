package business

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mantlenetworkio/lithosphere/business/mantle_da"
	"github.com/mantlenetworkio/lithosphere/common/tasks"
	"github.com/mantlenetworkio/lithosphere/config"
	"github.com/mantlenetworkio/lithosphere/database"
	"github.com/mantlenetworkio/lithosphere/database/business"
	"github.com/mantlenetworkio/lithosphere/database/event"
	"github.com/mantlenetworkio/lithosphere/database/exporter"
	"github.com/mantlenetworkio/lithosphere/synchronizer/node"
)

type BusinessProcessor struct {
	log                      log.Logger
	db                       *database.DB
	resourceCtx              context.Context
	resourceCancel           context.CancelFunc
	tasks                    tasks.Group
	l1Client                 node.EthClient
	l2Client                 node.EthClient
	mantleDA                 *mantle_da.MantleDataStore
	startDataStoreId         uint32
	fraudProofWindows        uint64
	L1AccountCheckingAddress string
	L2AccountCheckingAddress string
	L1StandardBridge         common.Address
	tokenListUrl             string
}

func NewBusinessProcessor(logger log.Logger, db *database.DB, l1Client node.EthClient, l2Client node.EthClient, da *mantle_da.MantleDataStore, cfg config.Config, shutdown context.CancelCauseFunc) *BusinessProcessor {

	resCtx, resCancel := context.WithCancel(context.Background())
	businessProcessor := BusinessProcessor{
		log:                      logger,
		db:                       db,
		resourceCtx:              resCtx,
		resourceCancel:           resCancel,
		l1Client:                 l1Client,
		l2Client:                 l2Client,
		mantleDA:                 da,
		fraudProofWindows:        cfg.FraudProofWindows,
		startDataStoreId:         cfg.StartDataStoreId,
		L1AccountCheckingAddress: cfg.CheckingAddress.L1AccountCheckingAddress,
		L2AccountCheckingAddress: cfg.CheckingAddress.L2AccountCheckingAddress,
		L1StandardBridge:         cfg.Chain.L1Contracts.L1StandardBridgeProxy,
		tokenListUrl:             cfg.TokenListUrl,
		tasks: tasks.Group{HandleCrit: func(err error) {
			shutdown(fmt.Errorf("critical error in business processor: %w", err))
		}},
	}
	return &businessProcessor
}

func (bp *BusinessProcessor) Start() error {
	bp.log.Info("starting business processor...")
	tickerRun := time.NewTicker(time.Second * 5)
	bp.tasks.Go(func() error {
		for range tickerRun.C {
			err := bp.onRollup()
			if err != nil {
				bp.log.Error("business processor onRollup", "error", err)
			}
		}
		return nil
	})

	ticker := time.NewTicker(time.Second)
	bp.tasks.Go(func() error {
		for range ticker.C {
			if err := bp.db.L2ToL1.UpdateTimeLeft(); err != nil {
				bp.log.Error("business processor UpdateTimeLeft", "error", err)
			}
		}
		return nil
	})

	tickerL1ToL2 := time.NewTicker(time.Second * 5)
	bp.tasks.Go(func() error {
		for range tickerL1ToL2.C {
			if err := bp.onDepositTxStatus(); err != nil {
				bp.log.Error("business processor onDepositTxStatus", "error", err)
			}
		}
		return nil
	})

	tickerL2ToL1 := time.NewTicker(time.Second * 5)
	bp.tasks.Go(func() error {
		for range tickerL2ToL1.C {
			if err := bp.onWithdrawTxStatus(); err != nil {
				bp.log.Error("business processor onWithdrawTxStatus", "error", err)
			}
		}
		return nil
	})

	tickerBridgeCheck := time.NewTicker(time.Hour * 6)
	bp.tasks.Go(func() error {
		for range tickerBridgeCheck.C {
			if err := bp.syncTokenBalance(); err != nil {
				bp.log.Error("business processor syncTokenBalance", "error", err)
			}
		}
		return nil
	})

	tokenListTicker := time.NewTicker(time.Minute * 1)
	bp.tasks.Go(func() error {
		for range tokenListTicker.C {
			if err := bp.syncTokenList(); err != nil {
				bp.log.Error(err.Error())
			}
		}
		return nil
	})

	return nil
}

func (bp *BusinessProcessor) Close() error {
	bp.resourceCancel()
	return bp.tasks.Wait()
}

func (bp *BusinessProcessor) onDepositTxStatus() error {
	if err := bp.markedL1ToL2Finalized(); err != nil {
		bp.log.Error("marked l2 to l1 finalized fail", "err", err)
		return err
	}
	return nil
}

func (bp *BusinessProcessor) onWithdrawTxStatus() error {
	if err := bp.syncL2ToL1StateRoot(); err != nil {
		bp.log.Error("sync l2 to l1 state root fail", "err", err)
		return err
	}
	if err := bp.markedL2ToL1Proven(); err != nil {
		bp.log.Error("marked l2 to l1 prove fail", "err", err)
		return err
	}
	if err := bp.markedL2ToL1Finalized(); err != nil {
		bp.log.Error("marked l2 to l1 finalized fail", "err", err)
		return err
	}
	return nil
}

func (bp *BusinessProcessor) onRollup() error {
	if err := bp.syncMantleDaData(); err != nil {
		bp.log.Error("sync mantle da data fail", "err", err)
		return err
	}
	if err := bp.syncStateRootStatus(); err != nil {
		bp.log.Error("sync state root status fail", "err", err)
		return err
	}
	return nil
}

func (bp *BusinessProcessor) syncMantleDaData() error {
	if err := bp.db.Transaction(func(tx *database.DB) error {
		maxEpochRange := uint64(2000)
		dataStoreId := tx.DataStore.LatestDataStoreId()
		if dataStoreId != 0 {
			bp.startDataStoreId = uint32(dataStoreId)
		}
		dataStoreList, err := tx.DataStoreEvent.DataStoreEventListByRange(uint64(bp.startDataStoreId), uint64(bp.startDataStoreId)+maxEpochRange)
		if err != nil {
			bp.log.Error("Get data store event fail", "err", err)
			return err
		}
		dataStores, dataStoreBlocks, latestDataStoreId, err := mantle_da.DataFromMantleDa(dataStoreList, bp.mantleDA, bp.log)
		if err != nil {
			bp.log.Error("Get dataStore from mantle da error", "err", err)
			return err
		}
		bp.startDataStoreId = latestDataStoreId
		bp.log.Info("latest data store id", "startDataStoreId", bp.startDataStoreId)
		if len(dataStores) != 0 {
			if err := tx.DataStore.StoreBatchDataStores(dataStores); err != nil {
				return err
			}
		}
		if len(dataStoreBlocks) != 0 {
			if err := tx.DataStore.StoreBatchDataStoreBlocks(dataStoreBlocks); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (bp *BusinessProcessor) syncL2ToL1StateRoot() error {
	blockNumber, err := bp.db.StateRoots.GetLatestStateRootL2BlockNumber()
	if err != nil {
		bp.log.Error(err.Error())
		return err
	}
	if blockNumber == 0 {
		return nil
	}
	bp.log.Info("get state root l2 block number success", "l2BlockNumber", blockNumber, "fraudProofWindows", bp.fraudProofWindows)
	err = bp.db.L2ToL1.UpdateReadyForProvedStatus(blockNumber, bp.fraudProofWindows)
	if err != nil {
		bp.log.Error(err.Error())
		return err
	}
	bp.log.Info("update ready for proven status success")
	return nil
}

func (bp *BusinessProcessor) syncStateRootStatus() error {
	latestSafeBlockHeader, err := bp.l1Client.LatestSafeBlockHeader()
	if err != nil {
		bp.log.Error(err.Error())
		return err
	}
	latestFinalizedBlockHeader, err := bp.l1Client.LatestFinalizedBlockHeader()
	if err != nil {
		bp.log.Error(err.Error())
		return err
	}
	err = bp.db.StateRoots.UpdateSafeStatus(latestSafeBlockHeader.Number)
	if err != nil {
		bp.log.Error(err.Error())
		return err
	}
	err = bp.db.StateRoots.UpdateFinalizedStatus(latestFinalizedBlockHeader.Number)
	if err != nil {
		bp.log.Error(err.Error())
		return err
	}
	bp.log.Info("update state root status success")
	return nil
}

func (bp *BusinessProcessor) markedL1ToL2Finalized() error {
	bp.log.Info("start marked l1 to l2 finalized")
	finalizedList, err := bp.db.RelayMessage.RelayMessageUnRelatedList()
	if err != nil {
		return err
	}
	var depositL2ToL1List []business.L1ToL2
	var needMarkDepositList []event.RelayMessage
	for i := range finalizedList {
		finalized := finalizedList[i]
		l1l2Tx := business.L1ToL2{
			L2TransactionHash: finalized.RelayTransactionHash,
			L1BlockNumber:     finalized.BlockNumber,
		}
		withdrawTx, _ := bp.db.L1ToL2.L1ToL2TransactionDeposit(finalized.MessageHash)
		if withdrawTx != nil {
			depositL2ToL1List = append(depositL2ToL1List, l1l2Tx)
			needMarkDepositList = append(needMarkDepositList, finalized)
		}
	}
	if err := bp.db.Transaction(func(tx *database.DB) error {
		if len(depositL2ToL1List) > 0 {
			if err := bp.db.L1ToL2.MarkL1ToL2TransactionDepositFinalized(depositL2ToL1List); err != nil {
				bp.log.Error("Marked l2 to l1 transaction withdraw proven fail", "err", err)
				return err
			}
			if err := bp.db.RelayMessage.MarkedRelayMessageRelated(needMarkDepositList); err != nil {
				bp.log.Error("Marked withdraw proven related fail", "err", err)
				return err
			}
			bp.log.Info("marked deposit transaction success", "deposit size", len(depositL2ToL1List), "marked size", len(needMarkDepositList))
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (bp *BusinessProcessor) markedL2ToL1Proven() error {
	bp.log.Info("start marked l2 to l1 proven")
	provenList, err := bp.db.WithdrawProven.WithdrawProvenUnRelatedList()
	if err != nil {
		return err
	}
	var withdrawL2ToL1List []business.L2ToL1
	var withdrawL2ToL1ListV0 []business.L2ToL1
	var needMarkWithdrawList []event.WithdrawProven
	var needMarkWithdrawListV0 []event.WithdrawProven
	for i := range provenList {
		provenTxn := provenList[i]
		l2l1Tx := business.L2ToL1{
			WithdrawTransactionHash: provenTxn.WithdrawHash,
			L1ProveTxHash:           provenTxn.ProvenTransactionHash,
			L1BlockNumber:           provenTxn.BlockNumber,
		}
		withdrawTx, _ := bp.db.L2ToL1.L2ToL1TransactionWithdrawal(provenTxn.WithdrawHash)
		if withdrawTx != nil {
			if withdrawTx.Version != 0 {
				withdrawL2ToL1List = append(withdrawL2ToL1List, l2l1Tx)
				needMarkWithdrawList = append(needMarkWithdrawList, provenTxn)
			} else {
				withdrawL2ToL1ListV0 = append(withdrawL2ToL1ListV0, l2l1Tx)
				needMarkWithdrawListV0 = append(needMarkWithdrawListV0, provenTxn)
			}
		}
	}
	if err := bp.db.Transaction(func(tx *database.DB) error {
		if len(withdrawL2ToL1List) > 0 {
			if err := bp.db.L2ToL1.MarkL2ToL1TransactionWithdrawalProven(withdrawL2ToL1List); err != nil {
				bp.log.Error("Marked l2 to l1 transaction withdraw proven fail", "err", err)
				return err
			}
			if err := bp.db.WithdrawProven.MarkedWithdrawProvenRelated(needMarkWithdrawList); err != nil {
				bp.log.Error("Marked withdraw proven related fail", "err", err)
				return err
			}
			bp.log.Info("marked proven transaction success", "withdraw size", len(provenList), "marked size", len(needMarkWithdrawList))
		}
		if len(withdrawL2ToL1ListV0) > 0 {
			if err := bp.db.L2ToL1.MarkL2ToL1TransactionWithdrawalProven(withdrawL2ToL1ListV0); err != nil {
				bp.log.Error("Marked l2 to l1 transaction withdraw proven fail", "err", err)
				return err
			}
			if err := bp.db.WithdrawProven.MarkedWithdrawProvenRelated(needMarkWithdrawListV0); err != nil {
				bp.log.Error("Marked withdraw proven related fail", "err", err)
				return err
			}
			bp.log.Info("marked proven v0 transaction success", "withdraw size", len(provenList), "marked size", len(needMarkWithdrawList))
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (bp *BusinessProcessor) markedL2ToL1Finalized() error {
	bp.log.Info("start marked l2 to l1 finalized")
	withdrawList, err := bp.db.WithdrawFinalized.WithdrawFinalizedUnRelatedList()
	if err != nil {
		bp.log.Error("fetch withdraw finalized un-related list fail", "err", err)
		return err
	}
	var withdrawL2ToL1List []business.L2ToL1
	var withdrawL2ToL1ListV0 []business.L2ToL1
	var needMarkWithdrawList []event.WithdrawFinalized
	var needMarkWithdrawListV0 []event.WithdrawFinalized
	for i := range withdrawList {
		finalizedTxn := withdrawList[i]
		l2l1Tx := business.L2ToL1{
			WithdrawTransactionHash: finalizedTxn.WithdrawHash,
			L1FinalizeTxHash:        finalizedTxn.FinalizedTransactionHash,
			L1BlockNumber:           finalizedTxn.BlockNumber,
		}
		withdrawTx, _ := bp.db.L2ToL1.L2ToL1TransactionWithdrawal(finalizedTxn.WithdrawHash)
		if withdrawTx != nil {
			if withdrawTx != nil {
				if withdrawTx.Version != 0 {
					withdrawL2ToL1List = append(withdrawL2ToL1List, l2l1Tx)
					needMarkWithdrawList = append(needMarkWithdrawList, finalizedTxn)
				} else {
					withdrawL2ToL1ListV0 = append(withdrawL2ToL1ListV0, l2l1Tx)
					needMarkWithdrawListV0 = append(needMarkWithdrawListV0, finalizedTxn)
				}
			}
		}
	}
	if err := bp.db.Transaction(func(tx *database.DB) error {
		if len(withdrawL2ToL1List) > 0 {
			if err := bp.db.L2ToL1.MarkL2ToL1TransactionWithdrawalFinalized(withdrawL2ToL1List); err != nil {
				bp.log.Error("Marked l2 to l1 transaction withdraw finalized fail", "err", err)
				return err
			}
			if err := bp.db.WithdrawFinalized.MarkedWithdrawFinalizedRelated(needMarkWithdrawList); err != nil {
				bp.log.Error("Marked withdraw finalized related fail", "err", err)
				return err
			}
			bp.log.Info("marked finalized transaction success", "withdraw size", len(withdrawList), "marked size", len(needMarkWithdrawList))
		}
		if len(withdrawL2ToL1ListV0) > 0 {
			if err := bp.db.L2ToL1.MarkL2ToL1TransactionWithdrawalFinalizedV0(withdrawL2ToL1ListV0); err != nil {
				bp.log.Error("Marked l2 to l1 transaction withdraw proven fail", "err", err)
				return err
			}
			if err := bp.db.WithdrawFinalized.MarkedWithdrawFinalizedRelated(needMarkWithdrawListV0); err != nil {
				bp.log.Error("Marked withdraw proven related fail", "err", err)
				return err
			}
			bp.log.Info("marked proven v0 transaction success", "withdraw size", len(withdrawL2ToL1ListV0), "marked size", len(needMarkWithdrawList))
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (bp *BusinessProcessor) syncTokenBalance() error {
	l1AddressList := bp.L1AccountCheckingAddress
	l1Addresses := strings.Split(l1AddressList, " ")
	l2AddressList := bp.L2AccountCheckingAddress
	l2Addresses := strings.Split(l2AddressList, " ")
	checkpoints := bp.db.CheckPoint.GetLatestBridgeCheckpoint()

	for i, l1Address := range l1Addresses {

		l1BlockHeader, err := bp.db.Blocks.L1LatestBlockHeader()
		if err != nil {
			bp.log.Error(err.Error())
			return err
		}

		l1LatestBlockNumber := new(big.Int)
		l1LatestBlockNumber = l1BlockHeader.Number

		l2BlockHeader, err := bp.db.Blocks.L2LatestBlockHeader()
		if err != nil {
			bp.log.Error(err.Error())
			return err
		}

		l2LatestBlockNumber := new(big.Int)
		l2LatestBlockNumber = l2BlockHeader.Number

		latestBalance := new(big.Int)
		if strings.Compare(l1Address, "0x0000000000000000000000000000000000000000") == 0 {

			latestBalance, err = bp.l1Client.GetBalanceByBlockNumber(bp.L1StandardBridge.String(), l1LatestBlockNumber)
			if err != nil {
				bp.log.Error(err.Error())
				return err
			}
		} else {

			latestBalance, err = bp.l1Client.GetERC20Balance(common.HexToAddress(l1Address), bp.L1StandardBridge, l1LatestBlockNumber)
			if err != nil {
				bp.log.Error(err.Error())
				return err
			}
		}

		latestTotalSupply, err := bp.l2Client.GetERC20TotalSupply(l2Addresses[i], l2LatestBlockNumber)
		if err != nil {
			bp.log.Error(err.Error())
			return err
		}

		// Create new checkpoint data
		newCheckpoint := exporter.BridgeCheckpoint{
			SnapshotTime:    time.Now(),
			L1Number:        l1LatestBlockNumber.Uint64(),
			L1TokenAddress:  l1Address,
			L2Number:        l2LatestBlockNumber.Uint64(),
			L2TokenAddress:  l2Addresses[i],
			L1BridgeBalance: latestBalance.String(),
			TotalSupply:     latestTotalSupply.String(),
			Checked:         false,
			Status:          1,
		}

		if len(checkpoints) == 0 {
			newCheckpoint.Checked = true
			err := bp.db.Transaction(func(tx *database.DB) error {
				result := tx.CheckPoint.StoreBridgeCheckpoint(newCheckpoint)
				if result != nil {
					bp.log.Error(fmt.Errorf("failed to save newCheckpoint to bridge_checkpoint, err:%s", result.Error()).Error())
					return result
				}
				return nil
			})
			bp.log.Info("first checkpoint has ready!")
			if err != nil {
				return err
			}
		} else {

			latestDifferenceValue := new(big.Int).Sub(latestBalance, latestTotalSupply)

			for _, checkpoint := range checkpoints {

				if strings.Compare(checkpoint.L1TokenAddress, l1Address) == 0 {

					l1BridgeBalance, err := new(big.Int).SetString(checkpoint.L1BridgeBalance, 10)
					if !err {
						bp.log.Error("Failed to convert l1BridgeBalance to big.Int")
						return nil
					}

					totalSupply, err := new(big.Int).SetString(checkpoint.TotalSupply, 10)
					if !err {
						bp.log.Error("Failed to convert TotalSupply to big.Int")
						return nil
					}
					differenceValue := new(big.Int).Sub(l1BridgeBalance, totalSupply)

					if latestDifferenceValue.Cmp(differenceValue) == 0 {
						newCheckpoint.Checked = true
						err := bp.db.CheckPoint.StoreBridgeCheckpoint(newCheckpoint)
						if err != nil {
							bp.log.Error(err.Error())
							return err
						}

						bp.log.Info(fmt.Sprintf("Layer 1 and layer 2 token:%s amounts aligned. Latest bridge balance: %s, totalSupply: %s.\nCheckpoint bridge balance: %s, checkpoint totalsupply: %s.",
							checkpoint.L2TokenAddress,
							latestBalance.String(),
							latestTotalSupply.String(),
							checkpoint.L1BridgeBalance,
							checkpoint.TotalSupply))

					} else {
						deposits := new(business.L1ToL2s)
						l2NotRelayedDeposits := new(business.L1ToL2s)
						withdraws := new(business.L2ToL1s)
						claimedWithdraws := new(business.L2ToL1s)

						// 1层deposit，2层暂未到账的情况
						l2NotRelayedDeposits = bp.db.CheckPoint.GetL1DepositUnrelay(checkpoint.L1Number, l1LatestBlockNumber.Uint64(), checkpoint.L1TokenAddress, common.Hash{}.String())

						// 2层withdraw，1层暂未claimed的情况
						withdraws = bp.db.CheckPoint.GetL2WithdrawUnclaimed(checkpoint.L2Number, l2LatestBlockNumber.Uint64(), checkpoint.L2TokenAddress, common.Hash{}.String())

						// 1层在checkpoint之前充值，2层确认在checkpoint之后
						deposits = bp.db.CheckPoint.GetL1DepositRelayed(checkpoint.L1Number, checkpoint.L2Number, checkpoint.L1TokenAddress, common.Hash{}.String())

						// 2层withdraw，1层claimed的情况
						claimedWithdraws = bp.db.CheckPoint.GetL2WithdrawClaimed(checkpoint.L1Number, l1LatestBlockNumber.Uint64(), checkpoint.L2TokenAddress, common.Hash{}.String())

						var depositsString strings.Builder
						if len(*deposits) != 0 {
							for _, deposit := range *deposits {
								if deposit.L1TokenAddress.String() == "0x0000000000000000000000000000000000000000" {
									latestDifferenceValue.Sub(latestDifferenceValue, deposit.ETHAmount)
									depositsString.WriteString(fmt.Sprintf("deposit txHash: %s\n", deposit.L1TransactionHash))
								} else {
									latestDifferenceValue.Sub(latestDifferenceValue, deposit.ERC20Amount)
									depositsString.WriteString(fmt.Sprintf("deposit txHash: %s\n", deposit.L1TransactionHash))
								}
							}
						}

						var withdrawsString strings.Builder
						if len(*withdraws) != 0 {
							for _, withdraw := range *withdraws {
								if withdraw.L1TokenAddress.String() == "0x0000000000000000000000000000000000000000" {
									latestDifferenceValue.Sub(latestDifferenceValue, withdraw.ETHAmount)
									withdrawsString.WriteString(fmt.Sprintf("withdraw txHash: %s\n", withdraw.L2TransactionHash))
								} else {
									latestDifferenceValue.Sub(latestDifferenceValue, withdraw.ERC20Amount)
									withdrawsString.WriteString(fmt.Sprintf("withdraw txHash: %s\n", withdraw.L2TransactionHash))
								}
							}
						}
						var l2NotRelayedDepositString strings.Builder
						if len(*l2NotRelayedDeposits) != 0 {
							for _, l2NotRelayedDeposit := range *l2NotRelayedDeposits {
								if l2NotRelayedDeposit.L1TokenAddress.String() == "0x0000000000000000000000000000000000000000" {
									latestDifferenceValue.Add(latestDifferenceValue, l2NotRelayedDeposit.ETHAmount)
									l2NotRelayedDepositString.WriteString(fmt.Sprintf("deposit and not relayed in l2,txHash: %s\n", l2NotRelayedDeposit.L1TransactionHash))
								} else {
									latestDifferenceValue.Add(latestDifferenceValue, l2NotRelayedDeposit.ERC20Amount)
									l2NotRelayedDepositString.WriteString(fmt.Sprintf("deposit and not relayed in l2,txHash: %s\n", l2NotRelayedDeposit.L1TransactionHash))
								}
							}
						}
						var claimedWithdrawsString strings.Builder
						if len(*claimedWithdraws) != 0 {
							for _, claimedWithdraw := range *claimedWithdraws {
								if claimedWithdraw.L1TokenAddress.String() == "0x0000000000000000000000000000000000000000" {
									latestDifferenceValue.Add(latestDifferenceValue, claimedWithdraw.ETHAmount)
									claimedWithdrawsString.WriteString(fmt.Sprintf("claimed withdraw txHash: %s\n", claimedWithdraw.L2TransactionHash))
								} else {
									latestDifferenceValue.Add(latestDifferenceValue, claimedWithdraw.ERC20Amount)
									claimedWithdrawsString.WriteString(fmt.Sprintf("claimed withdraw txHash: %s\n", claimedWithdraw.L2TransactionHash))
								}
							}
						}

						if latestDifferenceValue.Cmp(differenceValue) == 0 {
							newCheckpoint.Checked = true
							err := bp.db.CheckPoint.StoreBridgeCheckpoint(newCheckpoint)
							if err != nil {
								bp.log.Error(err.Error())
								return err
							}
							bp.log.Info(fmt.Sprintf("Layer 1 and layer 2 token:%s amounts aligned. Latest bridge balance: %s, totalSupply: %s.\nCheckpoint bridge balance: %s, checkpoint totalsupply: %s.\nThere are some deposit transaction that is not confirmed in Layer 2, %s.\nThere are some withdraw transaction that is not claimed in Layer 1 %s",
								checkpoint.L2TokenAddress,
								latestBalance.String(),
								latestTotalSupply.String(),
								checkpoint.L1BridgeBalance,
								checkpoint.TotalSupply,
								depositsString.String(),
								withdrawsString.String()))

						} else {

							err := bp.db.CheckPoint.StoreBridgeCheckpoint(newCheckpoint)

							if err != nil {
								bp.log.Error(err.Error())
								return err
							}

							bp.log.Warn(fmt.Sprintf("Layer 1 and layer 2 token:%s amounts are not aligned.amount difference is: %s, Latest bridge balance: %s, totalSupply: %s.\n checkpoint bridge balance: %s, checkpoint totalsupply: %s.\n There are some deposit transaction that is not confirmed in Layer 2, %s.\nThere are some withdraw transaction that is not claimed in Layer 1 %s\n Claimed withdraws txs %s\n l2 not relayed txs before checkpoint %s",
								checkpoint.L2TokenAddress,
								new(big.Int).Sub(latestDifferenceValue, differenceValue).String(),
								latestBalance.String(),
								latestTotalSupply.String(),
								checkpoint.L1BridgeBalance,
								checkpoint.TotalSupply,
								depositsString.String(),
								withdrawsString.String(),
								claimedWithdrawsString.String(),
								l2NotRelayedDepositString.String()))
						}
					}
				}
			}
		}
	}
	bp.log.Info("done")
	return nil
}

func (bp *BusinessProcessor) syncTokenList() error {
	var myClient = &http.Client{Timeout: 10 * time.Second}
	var loader = make(map[string]json.RawMessage)
	var resp *http.Response
	var err error

	if resp, err = myClient.Get(bp.tokenListUrl); err != nil {
		bp.log.Error(err.Error())
		return err
	}
	defer resp.Body.Close()

	// parse return string
	bytes, _ := io.ReadAll(resp.Body)
	if err = json.Unmarshal(bytes, &loader); err != nil {
		bp.log.Error(err.Error())
		return err
	}
	tokenList := make([]business.TokenList, 0)
	if err = json.Unmarshal(loader["tokens"], &tokenList); err != nil {
		bp.log.Error(err.Error())
		return err
	}

	for _, token := range tokenList {
		token.Timestamp = uint64(time.Now().Unix())
		err = bp.db.TokenList.SaveTokenList(token)
		if err != nil {
			bp.log.Error(err.Error())
			continue
		}
	}
	bp.log.Info("token list updated!")

	return nil

}
