package business

import (
	"gorm.io/gorm"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	common2 "github.com/mantlenetworkio/lithosphere/database/common"
	"github.com/mantlenetworkio/lithosphere/database/utils"
)

type DataStore struct {
	GUID                 uuid.UUID   `gorm:"primaryKey" json:"guid"`
	DataStoreId          uint64      `gorm:"column:data_store_id" json:"dataStoreId"`
	DurationDataStoreId  uint64      `gorm:"column:duration_data_store_id" json:"durationDataStoreId"`
	DataInitHash         common.Hash `gorm:"serializer:bytes" json:"dataInitHash"`
	DataConfirmHash      common.Hash `gorm:"serializer:bytes" json:"dataConfirmHash"`
	FromStoreNumber      *big.Int    `gorm:"serializer:u256" json:"fromStoreNumber"`
	StakeFromBlockNumber *big.Int    `gorm:"serializer:u256;column:stake_from_block_number" json:"stakeFromBlockNumber"`
	InitGasUsed          *big.Int    `gorm:"serializer:u256" json:"initGasUsed"`
	InitBlockNumber      *big.Int    `gorm:"serializer:u256" json:"initBlockNumber"`
	ConfirmGasUsed       *big.Int    `gorm:"serializer:u256" json:"confirmGasUsed"`
	DataHash             string      `json:"dataHash"`
	EthSign              string      `json:"ethSign"`
	MantleSign           string      `json:"mantleSign"`
	SignatoryRecord      string      `json:"signatoryRecord"`
	InitTime             uint64      `json:"initTime"`
	ExpireTime           uint64      `json:"expireTime"`
	NumSys               uint32      `json:"numSys"`
	NumPar               uint32      `json:"numPar"`
	LowDegree            string      `json:"lowDegree"`
	Status               bool        `json:"status"`
	Confirmer            string      `json:"confirmer"`
	Header               string      `json:"header"`
	DataCommitment       string      `json:"dataCommitment"`
	Timestamp            uint64      `json:"timestamp"`
	DataSize             *big.Int    `gorm:"serializer:u256" json:"dataSize"`
}

func (DataStore) TableName() string {
	return "data_store"
}

type DataStoreDB interface {
	DataStoreView
	StoreBatchDataStores([]DataStore) error
	StoreBatchDataStoreBlocks([]DataStoreBlock) error
}

type DataStoreView interface {
	DataStoreList(int, int, string) ([]DataStore, int64)
	DataStoreById(id *big.Int) (*DataStore, error)
	DataStoreBlockById(id *big.Int) ([]DataStoreBlock, error)
	LatestDataStoreId() uint64
	DataStoreL1BlockHeader() (*common2.L1BlockHeader, error)
}

type dataStoreDB struct {
	gorm *gorm.DB
}

func NewDataStoreDB(db *gorm.DB) DataStoreDB {
	return &dataStoreDB{gorm: db}
}

func (d dataStoreDB) DataStoreBlockById(id *big.Int) ([]DataStoreBlock, error) {
	var daBlockList []DataStoreBlock
	err := d.gorm.Table("data_store_block").Where("data_store_id=(?)", id.Uint64()).Order("data_store_id desc").Find(&daBlockList).Error
	if err != nil {
		log.Error("get l2 to l1 list fail", "err", err)
	}
	return daBlockList, nil
}

func (d dataStoreDB) StoreBatchDataStoreBlocks(blocks []DataStoreBlock) error {
	result := d.gorm.CreateInBatches(&blocks, len(blocks))
	return result.Error
}

func (d dataStoreDB) StoreBatchDataStores(stores []DataStore) error {
	result := d.gorm.CreateInBatches(&stores, utils.BatchInsertSize)
	return result.Error
}

func (d dataStoreDB) DataStoreById(id *big.Int) (*DataStore, error) {
	var dataStore DataStore
	dataStoreQuery := d.gorm.Where("data_store_id=?", id.Uint64())
	result := dataStoreQuery.Take(&dataStore)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Error("Record not found")
			return nil, nil
		}
		log.Error("Query data fail", "err", result.Error)
	}
	return &dataStore, nil
}

func (d dataStoreDB) LatestDataStoreId() uint64 {
	var dataStore DataStore
	result := d.gorm.Order("data_store_id desc").Take(&dataStore).Limit(1)
	if result.Error != nil {
		return 0
	}
	return dataStore.DataStoreId
}

func (d dataStoreDB) DataStoreList(page int, pageSize int, order string) (dataStoreList []DataStore, total int64) {
	var totalRecord int64
	var daList []DataStore
	err := d.gorm.Table("data_store").Select("data_store_id").Count(&totalRecord).Error
	if err != nil {
		log.Error("get l2 to l1 count fail")
	}
	queryStateRoot := d.gorm.Table("data_store").Offset((page - 1) * pageSize).Limit(pageSize)
	if strings.ToLower(order) == "asc" {
		queryStateRoot.Order("timestamp asc")
	} else {
		queryStateRoot.Order("timestamp desc")
	}
	qErr := queryStateRoot.Find(&daList).Error
	if qErr != nil {
		log.Error("get l2 to l1 list fail", "err", err)
	}
	return daList, totalRecord
}

func (d dataStoreDB) DataStoreL1BlockHeader() (*common2.L1BlockHeader, error) {
	l1Query := d.gorm.Where("number = (?)", d.gorm.Table("data_store").Select("MAX(from_store_number)"))
	var l1Header common2.L1BlockHeader
	result := l1Query.Take(&l1Header)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &l1Header, nil
}
