package business

import (
	"github.com/google/uuid"
	"math/big"
)

type SymbolTvl struct {
	GUID        uuid.UUID `gorm:"primaryKey" json:"guid"`
	Symbol      string    `gorm:"column:symbol" json:"symbol"`
	Amount      *big.Int  `gorm:"serializer:u256" json:"amount"`
	LatestPrice string    `gorm:"column:latest_price" json:"latestPrice"`
	TransToUsd  *big.Int  `gorm:"serializer:u256" json:"transToUsd"`
	Timestamp   uint64    `json:"timestamp" json:"timestamp"`
}
