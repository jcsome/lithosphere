package bridge

import (
	"fmt"

	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/google/uuid"

	"github.com/ethereum-optimism/optimism/indexer/bigint"
	"github.com/mantlenetworkio/lithosphere/config"
	"github.com/mantlenetworkio/lithosphere/database"
	"github.com/mantlenetworkio/lithosphere/database/business"
	contracts2 "github.com/mantlenetworkio/lithosphere/event/processors/contracts"
	"github.com/mantlenetworkio/lithosphere/synchronizer/node"
)

func LegacyL1ProcessInitiatedBridgeEvents(log log.Logger, db *database.DB, metrics L1Metricer, l1Contracts config.L1Contracts, fromHeight, toHeight *big.Int) error {
	// (1) CanonicalTransactionChain
	ctcTxDepositEvents, err := contracts2.LegacyCTCDepositEvents(l1Contracts.LegacyCanonicalTransactionChain, db, fromHeight, toHeight)
	if err != nil {
		return err
	}
	if len(ctcTxDepositEvents) > 0 {
		log.Info("detected legacy transaction deposits", "size", len(ctcTxDepositEvents))
	}

	mintedWEI := bigint.Zero
	ctcTxDeposits := make(map[logKey]*contracts2.LegacyCTCDepositEvent, len(ctcTxDepositEvents))
	l1ToL2s := make([]business.L1ToL2, len(ctcTxDeposits))
	for i := range ctcTxDepositEvents {
		depositTx := ctcTxDepositEvents[i]
		mintedWEI = new(big.Int).Add(mintedWEI, depositTx.ETHAmount)
		blockNumber, err := db.L1ToL2.GetBlockNumberFromHash(depositTx.Event.BlockHash)
		if err != nil {
			log.Error("can not get l1 blockNumber", "blockHash", depositTx.Event.BlockHash)
			return err
		}
		l1ToL2s[i] = business.L1ToL2{
			L1BlockNumber:     blockNumber,
			QueueIndex:        nil,
			L1TransactionHash: depositTx.Event.TransactionHash,
			L2TransactionHash: common.Hash{},
			MessageHash:       common.Hash{},
			L1TxOrigin:        depositTx.FromAddress,
			Status:            0,
			L1TokenAddress:    common.Address{},
			L2TokenAddress:    common.Address{},
			ETHAmount:         depositTx.ETHAmount,
			ERC20Amount:       depositTx.ERC20Amount,
			GasLimit:          depositTx.GasLimit,
			Timestamp:         int64(depositTx.Event.Timestamp),
		}
	}
	if len(ctcTxDepositEvents) > 0 {
		if err := db.L1ToL2.StoreL1ToL2Transactions(l1ToL2s); err != nil {
			return err
		}
		mintedETH, _ := bigint.WeiToETH(mintedWEI).Float64()
		metrics.RecordL1TransactionDeposits(len(ctcTxDepositEvents), mintedETH)
	}

	// (2) L1CrossDomainMessenger
	crossDomainSentMessages, err := contracts2.CrossDomainMessengerSentMessageEvents("l1", l1Contracts.L1CrossDomainMessengerProxy, db, fromHeight, toHeight)
	if err != nil {
		return err
	}
	if len(crossDomainSentMessages) > 0 {
		log.Info("detected legacy sent messages", "size", len(crossDomainSentMessages))
	}

	sentMessages := make(map[logKey]*contracts2.CrossDomainMessengerSentMessageEvent, len(crossDomainSentMessages))
	l1ToL2c2 := make([]business.L1ToL2, len(crossDomainSentMessages))
	for i := range crossDomainSentMessages {
		sentMessage := crossDomainSentMessages[i]
		sentMessages[logKey{sentMessage.Event.BlockHash, sentMessage.Event.LogIndex}] = &sentMessage
		l1ToL2c2[i].L1TransactionHash = sentMessage.Event.TransactionHash
		l1ToL2c2[i].MessageHash = sentMessage.MessageHash
	}
	if len(crossDomainSentMessages) > 0 {
		if err := db.L1ToL2.UpdateMessageHashByTxHash(l1ToL2c2); err != nil {
			return err
		}
		metrics.RecordL1CrossDomainSentMessages(len(crossDomainSentMessages))
	}

	// (3) L1StandardBridge
	initiatedBridges, err := contracts2.L1StandardBridgeLegacyDepositInitiatedEvents(l1Contracts.L1StandardBridgeProxy, db, fromHeight, toHeight)
	if err != nil {
		return err
	}
	if len(initiatedBridges) > 0 {
		log.Info("detected iegacy bridge deposits", "size", len(initiatedBridges))
	}

	bridgedTokens := make(map[common.Address]int)
	l1ToL2s2 := make([]business.L1ToL2, len(initiatedBridges))
	for i := range initiatedBridges {
		initiatedBridge := initiatedBridges[i]
		sentMessage, ok := sentMessages[logKey{initiatedBridge.Event.BlockHash, initiatedBridge.Event.LogIndex - 1}]
		if !ok {
			log.Error("missing cross domain message for bridge transfer", "tx_hash", initiatedBridge.Event.TransactionHash.String())
			return fmt.Errorf("expected SentMessage preceding DepositInitiated event. tx_hash = %s", initiatedBridge.Event.TransactionHash.String())
		}
		ctcTxDeposit, ok := ctcTxDeposits[logKey{initiatedBridge.Event.BlockHash, initiatedBridge.Event.LogIndex - 2}]
		if !ok {
			log.Error("missing transaction deposit for bridge transfer", "tx_hash", initiatedBridge.Event.TransactionHash.String())
			return fmt.Errorf("expected TransactionEnqueued preceding DepostInitiated event. tx_hash = %s", initiatedBridge.Event.TransactionHash.String())
		}

		initiatedBridge.CrossDomainMessageHash = &sentMessage.MessageHash
		bridgedTokens[initiatedBridge.LocalTokenAddress]++

		l1ToL2s2[i].L1TransactionHash = ctcTxDeposit.Event.TransactionHash
		l1ToL2s2[i].FromAddress = initiatedBridge.FromAddress
		l1ToL2s2[i].ToAddress = initiatedBridge.ToAddress
		l1ToL2s2[i].L1TokenAddress = initiatedBridge.LocalTokenAddress
		l1ToL2s2[i].L2TokenAddress = initiatedBridge.RemoteTokenAddress
		l1ToL2s2[i].ERC20Amount = initiatedBridge.ERC20Amount
	}
	if len(initiatedBridges) > 0 {
		if err := db.L1ToL2.UpdateL1ToL2InfoByTxHash(l1ToL2s2); err != nil {
			return err
		}
		for tokenAddr, size := range bridgedTokens {
			metrics.RecordL1InitiatedBridgeTransfers(tokenAddr, size)
		}
	}
	return nil
}

func LegacyL2ProcessInitiatedBridgeEvents(log log.Logger, db *database.DB, metrics L2Metricer, l2Contracts config.L2Contracts, fromHeight, toHeight *big.Int) error {
	// (1) L2CrossDomainMessenger
	crossDomainSentMessages, err := contracts2.CrossDomainMessengerSentMessageEvents("l2", l2Contracts.L2CrossDomainMessenger, db, fromHeight, toHeight)
	if err != nil {
		return err
	}
	if len(crossDomainSentMessages) > 0 {
		log.Info("detected legacy transaction withdrawals (via L2CrossDomainMessenger)", "size", len(crossDomainSentMessages))
	}

	l2ToL1Cs := make([]business.L2ToL1, len(crossDomainSentMessages))
	withdrawnWEI := bigint.Zero
	sentMessages := make(map[logKey]*contracts2.CrossDomainMessengerSentMessageEvent, len(crossDomainSentMessages))
	for i := range crossDomainSentMessages {
		sentMessage := crossDomainSentMessages[i]
		sentMessages[logKey{sentMessage.Event.BlockHash, sentMessage.Event.LogIndex}] = &sentMessage
		withdrawnWEI = new(big.Int).Add(withdrawnWEI, sentMessage.ETHAmount)

		// To ensure consistency in the schema, we duplicate this as the "root" transaction withdrawal. The storage key in the message
		// passer contract is sha3(calldata + sender). The sender always being the L2CrossDomainMessenger pre-bedrock.
		withdrawalHash := crypto.Keccak256Hash(append(sentMessage.MessageCalldata, l2Contracts.L2CrossDomainMessenger[:]...))
		l1blockNumber, err := db.L1ToL2.GetBlockNumberFromHash(sentMessage.Event.BlockHash)
		if err != nil {
			log.Error("can not get l1 blockNumber", "blockHash", sentMessage.Event.BlockHash)
			return err
		}

		l2ToL1Cs[i] = business.L2ToL1{
			GUID:                    uuid.New(),
			L1BlockNumber:           l1blockNumber,
			MsgNonce:                sentMessage.Nonce,
			L2TransactionHash:       sentMessage.Event.TransactionHash,
			WithdrawTransactionHash: withdrawalHash,
			L1ProveTxHash:           common.Hash{},
			L1FinalizeTxHash:        common.Hash{},
			Status:                  0,
			ETHAmount:               sentMessage.ETHAmount,
			ERC20Amount:             sentMessage.ERC20Amount,
			GasLimit:                sentMessage.GasLimit,
			TimeLeft:                new(big.Int).SetUint64(0),
			L1TokenAddress:          common.Address{},
			L2TokenAddress:          common.Address{},
			Timestamp:               int64(sentMessage.Event.Timestamp),
		}
	}
	if len(crossDomainSentMessages) > 0 {
		if err := db.L2ToL1.StoreL2ToL1Transactions(l2ToL1Cs); err != nil {
			return err
		}
		withdrawnETH, _ := bigint.WeiToETH(withdrawnWEI).Float64()
		metrics.RecordL2TransactionWithdrawals(len(crossDomainSentMessages), withdrawnETH)
		metrics.RecordL2CrossDomainSentMessages(len(crossDomainSentMessages))
	}

	// (2) L2StandardBridge
	initiatedBridges, err := contracts2.L2StandardBridgeLegacyWithdrawalInitiatedEvents(l2Contracts.L2StandardBridge, db, fromHeight, toHeight)
	if err != nil {
		return err
	}
	if len(initiatedBridges) > 0 {
		log.Info("detected legacy bridge withdrawals", "size", len(initiatedBridges))
	}

	bridgedTokens := make(map[common.Address]int)

	l2ToL1Bs := make([]business.L2ToL1, len(initiatedBridges))

	for i := range initiatedBridges {
		initiatedBridge := initiatedBridges[i]

		// extract the cross domain message hash & deposit source hash from the following events
		// Unlike bedrock, the bridge events are emitted AFTER sending the cross domain message
		// 	- Event Flow: TransactionEnqueued -> SentMessage -> DepositInitiated
		sentMessage, ok := sentMessages[logKey{initiatedBridge.Event.BlockHash, initiatedBridge.Event.LogIndex - 1}]
		if !ok {
			log.Error("expected SentMessage preceding DepositInitiated event", "tx_hash", initiatedBridge.Event.TransactionHash.String())
			return fmt.Errorf("expected SentMessage preceding DepositInitiated event. tx_hash = %s", initiatedBridge.Event.TransactionHash)
		}
		initiatedBridge.CrossDomainMessageHash = &sentMessage.MessageHash
		bridgedTokens[initiatedBridge.LocalTokenAddress]++
		l2ToL1Bs[i].MessageHash = sentMessage.MessageHash
		l2ToL1Bs[i].FromAddress = initiatedBridge.FromAddress
		l2ToL1Bs[i].ToAddress = initiatedBridge.ToAddress
		l2ToL1Bs[i].L1TokenAddress = initiatedBridge.LocalTokenAddress
		l2ToL1Bs[i].L2TokenAddress = initiatedBridge.RemoteTokenAddress
	}
	if len(initiatedBridges) > 0 {
		if err := db.L2ToL1.UpdateL2ToL1InfoByTxHash(l2ToL1Bs); err != nil {
			return err
		}
		for tokenAddr, size := range bridgedTokens {
			metrics.RecordL2InitiatedBridgeTransfers(tokenAddr, size)
		}
	}
	return nil
}

func LegacyL1ProcessFinalizedBridgeEvents(log log.Logger, db *database.DB, metrics L1Metricer, l1Client node.EthClient, l1Contracts config.L1Contracts, fromHeight, toHeight *big.Int) error {
	crossDomainRelayedMessages, err := contracts2.CrossDomainMessengerRelayedMessageEvents("l1", l1Contracts.L1CrossDomainMessengerProxy, db, fromHeight, toHeight)
	if err != nil {
		return err
	}
	if len(crossDomainRelayedMessages) > 0 {
		log.Info("detected relayed messages", "size", len(crossDomainRelayedMessages))
	}

	l2ToL1Finalized := make([]business.L2ToL1, len(crossDomainRelayedMessages))
	for i := range crossDomainRelayedMessages {
		relayedMessage := crossDomainRelayedMessages[i]

		l2ToL1Finalized[i].MessageHash = relayedMessage.MessageHash
		l2ToL1Finalized[i].L1FinalizeTxHash = relayedMessage.Event.TransactionHash
	}
	if len(crossDomainRelayedMessages) > 0 {
		if err = db.L2ToL1.MarkL2ToL1TransactionWithdrawalFinalized(l2ToL1Finalized); err != nil {
			return err
		}
		metrics.RecordL1ProvenWithdrawals(len(crossDomainRelayedMessages))
		metrics.RecordL1FinalizedWithdrawals(len(crossDomainRelayedMessages))
		metrics.RecordL1CrossDomainRelayedMessages(len(crossDomainRelayedMessages))
	}

	return nil
}

func LegacyL2ProcessFinalizedBridgeEvents(log log.Logger, db *database.DB, metrics L2Metricer, l2Contracts config.L2Contracts, fromHeight, toHeight *big.Int) error {
	// (1) L2CrossDomainMessenger
	crossDomainRelayedMessages, err := contracts2.CrossDomainMessengerRelayedMessageEvents("l2", l2Contracts.L2CrossDomainMessenger, db, fromHeight, toHeight)
	if err != nil {
		return err
	}
	if len(crossDomainRelayedMessages) > 0 {
		log.Info("detected relayed legacy messages", "size", len(crossDomainRelayedMessages))
	}

	L1ToL2Fz := make([]business.L1ToL2, len(crossDomainRelayedMessages))
	for i := range crossDomainRelayedMessages {
		relayedMessage := crossDomainRelayedMessages[i]
		L1ToL2Fz[i].MessageHash = relayedMessage.MessageHash
		L1ToL2Fz[i].L2TransactionHash = relayedMessage.Event.TransactionHash
	}
	if len(crossDomainRelayedMessages) > 0 {
		if err := db.L1ToL2.UpdateMessageHashByTxHash(L1ToL2Fz); err != nil {
			log.Error("failed to relay cross domain message", "err", err)
			return err
		}
		metrics.RecordL2CrossDomainRelayedMessages(len(crossDomainRelayedMessages))
	}
	return nil
}
