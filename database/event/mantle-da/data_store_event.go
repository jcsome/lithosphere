package mantle_da

import (
	"gorm.io/gorm"

	"github.com/acmestack/gorm-plus/gplus"
	"github.com/google/uuid"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type DataStoreEvent struct {
	GUID        uuid.UUID   `gorm:"primaryKey" json:"guid"`
	DataStoreId uint64      `gorm:"column:data_store_id;primaryKey" json:"dataStoreId"`
	BlockHash   common.Hash `gorm:"column:block_hash;serializer:bytes"`
	Timestamp   uint64      `json:"timestamp"`
}

func (DataStoreEvent) TableName() string {
	return "data_store_event"
}

type DataStoreEventDB interface {
	DataStoreEventView
	StoreBatchDataStoreEvent([]DataStoreEvent) error
}

type DataStoreEventView interface {
	DataStoreEventListByRange(uint64, uint64) ([]*DataStoreEvent, error)
}

type dataStoreEventDB struct {
	gorm *gorm.DB
}

func NewDataStoreEvnetDB(db *gorm.DB) DataStoreEventDB {
	gplus.Init(db)
	return &dataStoreEventDB{gorm: db}
}

func (de dataStoreEventDB) DataStoreEventListByRange(fromDataStoreId uint64, endDataStoreId uint64) (des []*DataStoreEvent, err error) {
	query, dEvents := gplus.NewQuery[DataStoreEvent]()
	query.Gt(&dEvents.DataStoreId, fromDataStoreId)
	query.Lt(&dEvents.DataStoreId, endDataStoreId)
	query.OrderByAsc("timestamp")
	rstEventList, rstDB := gplus.SelectList[DataStoreEvent](query)
	if rstDB.Error != nil {
		return nil, rstDB.Error
	}
	log.Info("Query data store event success", "fromDataStoreId", fromDataStoreId, "endDataStoreId", endDataStoreId, "len(rstEventList)", len(rstEventList))
	return rstEventList, nil
}

func (de dataStoreEventDB) StoreBatchDataStoreEvent(events []DataStoreEvent) error {
	result := de.gorm.CreateInBatches(&events, len(events))
	return result.Error
}
