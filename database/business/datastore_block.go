package business

import (
	"math/big"

	"github.com/google/uuid"

	"github.com/ethereum/go-ethereum/common"
)

type DataStoreBlock struct {
	GUID          uuid.UUID   `gorm:"primaryKey" json:"guid"`
	DataStoreID   uint64      `gorm:"column:data_store_id;primaryKey" json:"dataStoreID"`
	BlockData     common.Hash `gorm:"serializer:bytes" json:"blockData"`
	L2TxHash      common.Hash `gorm:"serializer:bytes;column:l2_transaction_hash" json:"l2TxHash"`
	L2BlockNumber *big.Int    `gorm:"serializer:u256" json:"l2BlockNumber"`
	Canonical     bool        `json:"canonical"`
	Timestamp     uint64      `json:"timestamp"`
}

func (DataStoreBlock) TableName() string {
	return "data_store_block"
}
