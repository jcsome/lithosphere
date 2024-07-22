package business

import (
	"math/big"

	"github.com/google/uuid"
)

type NormalStat struct {
	GUID               uuid.UUID `gorm:"primaryKey" json:"guid"`
	TxCount            *big.Int  `gorm:"serializer:u256" json:"txCount"`
	ActiveUser         *big.Int  `gorm:"serializer:u256" json:"activeUser"`
	NftHolders         *big.Int  `gorm:"serializer:u256" json:"nftHolders"`
	NewUser            *big.Int  `gorm:"serializer:u256" json:"newUser"`
	DepositCount       *big.Int  `gorm:"serializer:u256" json:"depositCount"`
	WithdrawCount      *big.Int  `gorm:"serializer:u256" json:"withdrawCount"`
	DeveloperCount     *big.Int  `gorm:"serializer:u256" json:"developerCount"`
	SmartContractCount *big.Int  `gorm:"serializer:u256" json:"smartContractCount"`
	L1CostAmount       *big.Int  `gorm:"serializer:u256" json:"l1CostAmount"`
	L2FeeAmount        *big.Int  `gorm:"serializer:u256" json:"l2FeeAmount"`
	Timestamp          uint64    `json:"timestamp"`
}
