package exporter

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/mantlenetworkio/lithosphere/database/business"
)

type BridgeCheckpoint struct {
	ID              uint64    `gorm:"primary_key;column:id"`
	SnapshotTime    time.Time `gorm:"column:snapshot_time"`
	L1Number        uint64    `gorm:"index;column:l1_number"`
	L1TokenAddress  string    `gorm:"index;column:l1_token_address"`
	L2Number        uint64    `gorm:"index;column:l2_number"`
	L2TokenAddress  string    `gorm:"index;column:l2_token_address"`
	L1BridgeBalance string    `gorm:"column:l1_bridge_balance"`
	TotalSupply     string    `gorm:"column:total_supply"`
	Checked         bool      `gorm:"column:checked"`
	Status          int       `gorm:"column:status"`
}

type BridgeCheckpoints []*BridgeCheckpoint

func (BridgeCheckpoint) TableName() string {
	return "bridge_checkpoints"
}

type BridgeCheckpointDB interface {
	BridgeCheckpointView
	StoreBridgeCheckpoint(checkpoint BridgeCheckpoint) error
}

type BridgeCheckpointView interface {
	GetLatestBridgeCheckpoint() []BridgeCheckpoint
	GetL1DepositUnrelay(L1Number uint64, l1LatestBlockNumber uint64, L1TokenAddress string, l2TransactionHash string) *business.L1ToL2s
	GetL2WithdrawUnclaimed(L2Number uint64, l2LatestBlockNumber uint64, L2TokenAddress string, l1FinalizeTxHash string) *business.L2ToL1s
	GetL1DepositRelayed(L1Number uint64, L2Number uint64, L1TokenAddress string, l2TransactionHash string) *business.L1ToL2s
	GetL2WithdrawClaimed(L1Number uint64, l1LatestBlockNumber uint64, L2TokenAddress string, l1FinalizeTxHash string) *business.L2ToL1s
}

type bridgeCheckpointDB struct {
	gorm *gorm.DB
}

func NewBridgeCheckpointDB(db *gorm.DB) BridgeCheckpointDB {
	return &bridgeCheckpointDB{gorm: db}
}

func (bc bridgeCheckpointDB) GetLatestBridgeCheckpoint() []BridgeCheckpoint {
	var bridgeCheckpoints []BridgeCheckpoint
	subQuery := bc.gorm.Table("bridge_checkpoints").Select("MAX(id)").Group("l1_token_address")
	bc.gorm.Where("id IN (?)", subQuery).Find(&bridgeCheckpoints)

	return bridgeCheckpoints
}

func (bc bridgeCheckpointDB) StoreBridgeCheckpoint(checkpoint BridgeCheckpoint) error {
	result := bc.gorm.Create(&checkpoint)
	return result.Error
}

func (bc bridgeCheckpointDB) GetL1DepositUnrelay(L1Number uint64, l1LatestBlockNumber uint64, L1TokenAddress string, l2TransactionHash string) *business.L1ToL2s {
	var l1ToL2s business.L1ToL2s
	err := bc.gorm.Table("l1_to_l2").Where("l1_block_number > ? AND l1_block_number <= ? AND l1_token_address = ? AND l2_transaction_hash = ?", L1Number, l1LatestBlockNumber, L1TokenAddress, l2TransactionHash).Find(&l1ToL2s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
	}
	return &l1ToL2s
}

func (bc bridgeCheckpointDB) GetL2WithdrawUnclaimed(L2Number uint64, l2LatestBlockNumber uint64, L2TokenAddress string, l1FinalizeTxHash string) *business.L2ToL1s {
	var l2Tol1s business.L2ToL1s

	err := bc.gorm.Table("l2_to_l1").Where("l2_block_number > ? AND l2_block_number <= ? AND l2_token_address = ? AND l1_finalize_tx_hash = ?", L2Number, l2LatestBlockNumber, L2TokenAddress, l1FinalizeTxHash).Find(&l2Tol1s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
	}
	return &l2Tol1s
}

func (bc bridgeCheckpointDB) GetL1DepositRelayed(L1Number uint64, L2Number uint64, L1TokenAddress string, l2TransactionHash string) *business.L1ToL2s {
	var l1ToL2s business.L1ToL2s
	err := bc.gorm.Table("l1_to_l2").Where("l1_block_number < ? AND l2_block_number > ? AND l1_token_address = ? AND l2_transaction_hash != ?", L1Number, L2Number, L1TokenAddress, l2TransactionHash).Find(&l1ToL2s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
	}
	return &l1ToL2s
}

func (bc bridgeCheckpointDB) GetL2WithdrawClaimed(L1Number uint64, l1LatestBlockNumber uint64, L2TokenAddress string, l1FinalizeTxHash string) *business.L2ToL1s {
	var l2Tol1s business.L2ToL1s

	err := bc.gorm.Table("l2_to_l1").Where("l1_block_number > ? AND l1_block_number <= ? AND l2_token_address = ? AND l1_finalize_tx_hash != ?", L1Number, l1LatestBlockNumber, L2TokenAddress, l1FinalizeTxHash).Find(&l2Tol1s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
	}
	return &l2Tol1s
}
