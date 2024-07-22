package ovm1

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mantlenetworkio/lithosphere/event/processors/bridge/ovm1/crossdomain"
)

func CalcTransaction(legacyWithdrawal *crossdomain.LegacyWithdrawal, l1CrossdomainMessengerAddress *common.Address, l2ChainID *big.Int) (common.Hash, error) {
	withdrawal, err := crossdomain.CalcWithdrawalHash(legacyWithdrawal, l1CrossdomainMessengerAddress, l2ChainID)
	if err != nil {
		return common.Hash{}, err
	}
	hash, err := withdrawal.Hash() // WithdrawalHash
	if err != nil {
		return common.Hash{}, err
	}
	log.Info("withdraw hash", hash)
	return hash, nil
}
