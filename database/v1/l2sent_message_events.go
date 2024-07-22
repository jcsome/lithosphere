package v1

import (
	"math/big"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type L2SentMessageEvent struct {
	GUID         uuid.UUID      `gorm:"primaryKey;column:guid"`
	TxHash       common.Hash    `gorm:"serializer:bytes;column:tx_hash"`
	BlockNumber  *big.Int       `gorm:"column:block_number;serializer:u256"`
	Target       common.Address `gorm:"column:target;serializer:bytes"`
	Sender       common.Address `gorm:"column:sender;serializer:bytes"`
	Message      string         `gorm:"column:message"`
	MessageNonce *big.Int       `gorm:"serializer:u256;column:message_nonce"`
	GasLimit     *big.Int       `gorm:"serializer:u256;column:gas_limit"`
	Signature    common.Hash    `gorm:"serializer:bytes;column:signature"`
	L1Token      common.Address `gorm:"column:l1_token;serializer:bytes"`
	L2Token      common.Address `gorm:"column:l2_token;serializer:bytes"`
	FromAddress  common.Address `gorm:"column:from_address;serializer:bytes"`
	ToAddress    common.Address `gorm:"column:to_address;serializer:bytes"`
	Value        *big.Int       `gorm:"column:value;serializer:u256"`
	Timestamp    uint64
}

func (L2SentMessageEvent) TableName() string {
	return "l2_sent_message_events"
}

type L2SentMessageEventDB interface {
	L2SentMessageEventView
}

type L2SentMessageEventView interface {
	L2SentMessageList() []L2SentMessageEvent
}

type l2SentMessageEventDB struct {
	gorm *gorm.DB
}

func NewL2SentMessageEventDB(db *gorm.DB) L2SentMessageEventDB {
	return &l2SentMessageEventDB{gorm: db}
}

func (l l2SentMessageEventDB) L2SentMessageList() []L2SentMessageEvent {
	var l2SentMessageEvents []L2SentMessageEvent
	err := l.gorm.Table("l2_sent_message_events").Find(&l2SentMessageEvents).Error
	if err != nil {
		log.Error("get l2 send message events error", "err", err.Error())
		return nil
	}
	return l2SentMessageEvents
}
