package contracts

import (
	"encoding/hex"
	"math/big"

	"github.com/google/uuid"

	"github.com/ethereum/go-ethereum/common"

	"github.com/mantlenetworkio/lithosphere/database"
	"github.com/mantlenetworkio/lithosphere/database/business"
	"github.com/mantlenetworkio/lithosphere/database/event"
	"github.com/mantlenetworkio/lithosphere/event/op-bindings/bindings"
)

func L2OutputProposedEvent(contractAddress common.Address, db *database.DB, fromHeight, toHeight *big.Int) ([]business.StateRoot, error) {
	l2OutputAbi, err := bindings.L2OutputOracleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	outputProposedEventAbi := l2OutputAbi.Events["OutputProposed"]
	contractEventFilter := event.ContractEvent{ContractAddress: contractAddress, EventSignature: outputProposedEventAbi.ID}
	outputProposedEvents, err := db.ContractEvents.L1ContractEventsWithFilter(contractEventFilter, fromHeight, toHeight)
	if err != nil {
		return nil, err
	}
	stateRoots := make([]business.StateRoot, len(outputProposedEvents))
	for i := range outputProposedEvents {
		outputProposed := bindings.L2OutputOracleOutputProposed{Raw: *outputProposedEvents[i].RLPLog}
		err = UnpackLog(&outputProposed, outputProposedEvents[i].RLPLog, outputProposedEventAbi.Name, l2OutputAbi)
		if err != nil {
			return nil, err
		}
		l1BlockNumber, err := db.L1ToL2.GetBlockNumberFromHash(outputProposedEvents[i].BlockHash)
		if err != nil {
			return nil, err
		}
		stateRoots[i] = business.StateRoot{
			GUID:            uuid.New(),
			TransactionHash: outputProposedEvents[i].TransactionHash,
			BlockHash:       outputProposedEvents[i].BlockHash,
			OutputRoot:      hex.EncodeToString(outputProposed.OutputRoot[:]),
			OutputIndex:     outputProposed.L2OutputIndex,
			BatchSize:       outputProposed.L2OutputIndex,
			L1BlockNumber:   l1BlockNumber,
			L2BlockNumber:   outputProposed.L2BlockNumber,
			Timestamp:       outputProposed.L1Timestamp.Uint64(),
		}
	}
	return stateRoots, nil
}
