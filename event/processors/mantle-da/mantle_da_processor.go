package mantle_da

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mantlenetworkio/lithosphere/business/mantle_da"
	"github.com/mantlenetworkio/lithosphere/config"
	"github.com/mantlenetworkio/lithosphere/database"
	"github.com/mantlenetworkio/lithosphere/database/event"
)

func L1ProcessMantleDAEvents(log log.Logger, db *database.DB, l1Contracts config.L1Contracts, fromHeight, toHeight *big.Int) error {
	contractEventFilter := event.ContractEvent{ContractAddress: l1Contracts.DataLayrServiceManagerAddr}
	mantleDAEvents, err := db.ContractEvents.L1ContractEventsWithFilter(contractEventFilter, fromHeight, toHeight)
	if err != nil {
		return err
	}
	var logs = make([]types.Log, len(mantleDAEvents))
	for i := range mantleDAEvents {
		logs = append(logs, *mantleDAEvents[i].RLPLog)
	}

	if len(logs) != 0 {
		dataStoreEvents, err := mantle_da.ParseMantleDaEvent(l1Contracts.DataLayrServiceManagerAddr.String(), logs, log)
		if err != nil {
			return err
		}
		if len(dataStoreEvents) != 0 {
			if err := db.DataStoreEvent.StoreBatchDataStoreEvent(dataStoreEvents); err != nil {
				log.Error("store batch event error", "err", err)
				return nil
			}
		}
	}
	return nil
}
