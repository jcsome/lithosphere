package stateroot

import (
	"math/big"

	"github.com/ethereum/go-ethereum/log"

	"github.com/mantlenetworkio/lithosphere/config"
	"github.com/mantlenetworkio/lithosphere/database"
	"github.com/mantlenetworkio/lithosphere/event/processors/bridge"
	"github.com/mantlenetworkio/lithosphere/event/processors/contracts"
)

func LegacyL1ProcessSCCEvent(log log.Logger, db *database.DB, metrics bridge.L1Metricer, l1Contracts config.L1Contracts, fromHeight, toHeight *big.Int) error {
	sccEvents, err := contracts.LegacySCCBatchAppendedEvent(l1Contracts.LegacyStateCommitmentChain, db, fromHeight, toHeight)
	if err != nil {
		return err
	}
	if len(sccEvents) > 0 {
		log.Info("detected legacy scc state batch appended event", "size", len(sccEvents))
		if err := db.StateRoots.StoreBatchStateRoots(sccEvents); err != nil {
			return err
		}
	}
	return nil

}

func L2OutputEvent(log log.Logger, db *database.DB, metrics bridge.L1Metricer, l1Contracts config.L1Contracts, fromHeight, toHeight *big.Int) error {
	l2OutputProposedEvents, err := contracts.L2OutputProposedEvent(l1Contracts.L2OutputOracleProxy, db, fromHeight, toHeight)
	if err != nil {
		return err
	}
	if len(l2OutputProposedEvents) > 0 {
		log.Info("detected l2output proposed event", "size", len(l2OutputProposedEvents))
		if err := db.StateRoots.StoreBatchStateRoots(l2OutputProposedEvents); err != nil {
			log.Error("Store batch state roots fail")
			return err
		}
	}
	return nil
}
