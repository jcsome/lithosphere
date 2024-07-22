package contracts

import (
	"github.com/mantlenetworkio/lithosphere/database/utils"
	"github.com/mantlenetworkio/lithosphere/event/op-bindings/bindings"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/mantlenetworkio/lithosphere/database"
	"github.com/mantlenetworkio/lithosphere/database/event"
)

type L2ToL1MessagePasserMessagePassed struct {
	Event          *event.ContractEvent
	WithdrawalHash common.Hash
	GasLimit       *big.Int
	Nonce          *big.Int
	FromAddress    common.Address
	ToAddress      common.Address
	ETHAmount      *big.Int
	ERC20Amount    *big.Int
	Data           utils.Bytes
	Timestamp      uint64
}

func L2ToL1MessagePasserMessagePassedEvents(contractAddress common.Address, db *database.DB, fromHeight, toHeight *big.Int) ([]L2ToL1MessagePasserMessagePassed, error) {
	l2ToL1MessagePasserAbi, err := bindings.L2ToL1MessagePasserMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	messagePassedAbi := l2ToL1MessagePasserAbi.Events["MessagePassed"]
	contractEventFilter := event.ContractEvent{ContractAddress: contractAddress, EventSignature: messagePassedAbi.ID}
	messagePassedEvents, err := db.ContractEvents.L2ContractEventsWithFilter(contractEventFilter, fromHeight, toHeight)
	if err != nil {
		return nil, err
	}

	messagesPassed := make([]L2ToL1MessagePasserMessagePassed, len(messagePassedEvents))
	for i := range messagePassedEvents {
		messagePassed := bindings.L2ToL1MessagePasserMessagePassed{Raw: *messagePassedEvents[i].RLPLog}
		err := UnpackLog(&messagePassed, messagePassedEvents[i].RLPLog, messagePassedAbi.Name, l2ToL1MessagePasserAbi)
		if err != nil {
			return nil, err
		}

		messagesPassed[i] = L2ToL1MessagePasserMessagePassed{
			Event:          &messagePassedEvents[i].ContractEvent,
			WithdrawalHash: messagePassed.WithdrawalHash,
			Nonce:          messagePassed.Nonce,
			GasLimit:       messagePassed.GasLimit,
			FromAddress:    messagePassed.Sender,
			ToAddress:      messagePassed.Target,
			ETHAmount:      messagePassed.EthValue,
			Data:           messagePassed.Data,
			Timestamp:      messagePassedEvents[i].Timestamp,
			ERC20Amount:    messagePassed.MntValue,
		}
	}

	return messagesPassed, nil
}
