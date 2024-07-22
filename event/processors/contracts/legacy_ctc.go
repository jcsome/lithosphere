package contracts

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/mantlenetworkio/lithosphere/common/bigint"
	"github.com/mantlenetworkio/lithosphere/database"
	"github.com/mantlenetworkio/lithosphere/database/event"
	"github.com/mantlenetworkio/lithosphere/database/utils"
	"github.com/mantlenetworkio/lithosphere/event/op-bindings/legacy-bindings"
)

type LegacyCTCDepositEvent struct {
	Event       *event.ContractEvent
	FromAddress common.Address
	ToAddress   common.Address
	ETHAmount   *big.Int
	ERC20Amount *big.Int
	Data        utils.Bytes
	Timestamp   uint64
	TxHash      common.Hash
	GasLimit    *big.Int
}

func LegacyCTCDepositEvents(contractAddress common.Address, db *database.DB, fromHeight, toHeight *big.Int) ([]LegacyCTCDepositEvent, error) {
	ctcAbi, err := legacy_bindings.CanonicalTransactionChainMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	transactionEnqueuedEventAbi := ctcAbi.Events["TransactionEnqueued"]
	contractEventFilter := event.ContractEvent{ContractAddress: contractAddress, EventSignature: transactionEnqueuedEventAbi.ID}
	events, err := db.ContractEvents.L1ContractEventsWithFilter(contractEventFilter, fromHeight, toHeight)
	if err != nil {
		return nil, err
	}

	ctcTxDeposits := make([]LegacyCTCDepositEvent, len(events))
	for i := range events {
		txEnqueued := legacy_bindings.CanonicalTransactionChainTransactionEnqueued{Raw: *events[i].RLPLog}
		err = UnpackLog(&txEnqueued, events[i].RLPLog, transactionEnqueuedEventAbi.Name, ctcAbi)
		if err != nil {
			return nil, err
		}

		ctcTxDeposits[i] = LegacyCTCDepositEvent{
			Event:       &events[i].ContractEvent,
			GasLimit:    txEnqueued.GasLimit,
			TxHash:      types.NewTransaction(0, txEnqueued.Target, bigint.Zero, txEnqueued.GasLimit.Uint64(), nil, txEnqueued.Data).Hash(),
			FromAddress: txEnqueued.L1TxOrigin,
			ToAddress:   txEnqueued.Target,
			ETHAmount:   bigint.Zero,
			Data:        txEnqueued.Data,
			Timestamp:   events[i].Timestamp,
			ERC20Amount: bigint.Zero,
		}
	}
	return ctcTxDeposits, nil
}
