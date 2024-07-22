package business

import (
	"math/big"

	"github.com/google/uuid"
)

type ProtocolTvl struct {
	GUID         uuid.UUID `gorm:"primaryKey" json:"guid"`
	ProtocolName string    `gorm:"column:protocol_name" json:"protocolName"`
	Amount       *big.Int  `gorm:"serializer:u256" json:"amount"`
	LatestPrice  string    `gorm:"column:latest_price"`
	TransToUsd   *big.Int  `gorm:"serializer:u256" json:"transToUsd"`
	Timestamp    uint64    `json:"timestamp" json:"timestamp"`
}
