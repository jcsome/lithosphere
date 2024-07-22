package event

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math/big"

	"github.com/google/uuid"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	common2 "github.com/mantlenetworkio/lithosphere/database/common"
)

type RelayMessage struct {
	GUID                 uuid.UUID      `gorm:"primaryKey"`
	BlockNumber          *big.Int       `gorm:"serializer:u256;column:block_number" db:"block_number"`
	DepositHash          common.Hash    `gorm:"serializer:bytes"`
	RelayTransactionHash common.Hash    `gorm:"serializer:bytes"`
	MessageHash          common.Hash    `gorm:"serializer:bytes"`
	L1TokenAddress       common.Address `gorm:"column:l1_token_address;serializer:bytes"`
	L2TokenAddress       common.Address `gorm:"column:l2_token_address;serializer:bytes"`
	ETHAmount            *big.Int       `gorm:"serializer:u256;column:eth_amount"`
	ERC20Amount          *big.Int       `gorm:"serializer:u256;column:erc20_amount"`
	Related              bool           `json:"related"`
	Timestamp            uint64
}

func (RelayMessage) TableName() string {
	return "relay_message"
}

type RelayMessageDB interface {
	RelayMessageView
	StoreRelayMessage([]RelayMessage) error
	MarkedRelayMessageRelated(relayMessageList []RelayMessage) error
	UpdateRelayMessageInfo(relayMessageList []RelayMessage) error
}

type RelayMessageView interface {
	RelayMessageL1BlockHeader() (*common2.L1BlockHeader, error)
	RelayMessageUnRelatedList() ([]RelayMessage, error)
}

type relayMessageDB struct {
	gorm *gorm.DB
}

func NewRelayMessageDB(db *gorm.DB) RelayMessageDB {
	return &relayMessageDB{gorm: db}
}

func (rm relayMessageDB) RelayMessageL1BlockHeader() (*common2.L1BlockHeader, error) {
	l1Query := rm.gorm.Where("number = (?)", rm.gorm.Table("relay_message").Select("MAX(block_number)"))
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

func (rm relayMessageDB) StoreRelayMessage(relayMessageList []RelayMessage) error {
	result := rm.gorm.CreateInBatches(&relayMessageList, len(relayMessageList))
	return result.Error
}

func (rm relayMessageDB) MarkedRelayMessageRelated(relayMessageList []RelayMessage) error {
	for i := 0; i < len(relayMessageList); i++ {
		var relayMessages = RelayMessage{}
		result := rm.gorm.Where(&RelayMessage{MessageHash: relayMessageList[i].MessageHash}).Take(&relayMessages)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			return result.Error
		}
		relayMessages.Related = true
		err := rm.gorm.Save(relayMessages).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (rm relayMessageDB) UpdateRelayMessageInfo(relayMessageList []RelayMessage) error {
	for i := 0; i < len(relayMessageList); i++ {
		var relayMessages = RelayMessage{}
		result := rm.gorm.Where(&RelayMessage{MessageHash: relayMessageList[i].MessageHash}).Take(&relayMessages)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			return result.Error
		}
		relayMessages.L1TokenAddress = relayMessageList[i].L1TokenAddress
		relayMessages.L2TokenAddress = relayMessageList[i].L2TokenAddress
		relayMessages.ETHAmount = relayMessageList[i].ETHAmount
		relayMessages.ERC20Amount = relayMessageList[i].ERC20Amount
		relayMessages.MessageHash = relayMessageList[i].MessageHash
		err := rm.gorm.Save(relayMessages).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (rm relayMessageDB) RelayMessageUnRelatedList() ([]RelayMessage, error) {
	var unRelatedRelayList []RelayMessage
	err := rm.gorm.Table("relay_message").Where("related = ?", false).Find(&unRelatedRelayList).Error
	if err != nil {
		log.Error("get unrelated deposit finalized fail", "err", err)
	}
	return unRelatedRelayList, nil
}
