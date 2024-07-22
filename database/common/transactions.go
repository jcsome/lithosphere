package common

import (
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
	"math/big"

	"github.com/google/uuid"

	"github.com/ethereum/go-ethereum/common"
)

type Transactions struct {
	GUID                 uuid.UUID      `gorm:"primaryKey" json:"GUID"`
	BlockHash            common.Hash    `gorm:"serializer:bytes" json:"blockHash"`
	BlockNumber          *big.Int       `gorm:"serializer:u256" json:"blockNumber"`
	FromAddress          common.Address `gorm:"column:from_address;serializer:bytes" json:"fromAddress"`
	ToAddress            common.Address `gorm:"column:to_address;serializer:bytes" json:"toAddress"`
	Gas                  *big.Int       `gorm:"serializer:u256" json:"gas"`
	GasPrice             *big.Int       `gorm:"serializer:u256" json:"gasPrice"`
	TransactionHash      common.Hash    `gorm:"serializer:bytes" json:"transactionHash"`
	InputData            []byte         `gorm:"serializer:bytes" json:"inputData"`
	MaxFeePerGas         *big.Int       `gorm:"serializer:u256" json:"maxFeePerGas"`
	MaxPriorityFeePerGas *big.Int       `gorm:"serializer:u256" json:"maxPriorityFeePerGas"`
	GasUsed              *big.Int       `gorm:"serializer:u256" json:"gasUsed"`
	CumulativeGasUsed    *big.Int       `gorm:"serializer:u256" json:"cumulativeGasUsed"`
	EffectiveGasPrice    *big.Int       `gorm:"serializer:u256" json:"effectiveGasPrice"`
	L1Fee                *big.Int       `gorm:"serializer:u256" json:"l1Fee"`
	L1GasUsed            *big.Int       `gorm:"serializer:u256" json:"l1GasUsed"`
	L1GasPrice           *big.Int       `gorm:"serializer:u256" json:"l1GasPrice"`
	Nonce                *big.Int       `gorm:"serializer:u256" json:"nonce"`
	TransactionIndex     *big.Int       `gorm:"serializer:u256" json:"transactionIndex"`
	TxType               int64          `gorm:"column:tx_type" db:"tx_type" form:"tx_type" json:"txType"`
	R                    *big.Int       `gorm:"serializer:u256" json:"r"`
	S                    *big.Int       `gorm:"serializer:u256" json:"s"`
	V                    *big.Int       `gorm:"serializer:u256" json:"v"`
	Status               int64          `gorm:"column:statue" db:"statue" form:"statue" json:"status"`
	ContractAddress      common.Address `gorm:"column:to_address;serializer:bytes" json:"contractAddress"`
	Amount               *big.Int       `gorm:"serializer:u256" json:"amount"`
	YParity              *big.Int       `gorm:"serializer:bytes" json:"YParity"`
	Timestamp            uint64
}

func (Transactions) TableName() string {
	return "transactions"
}

type TransactionsDB interface {
	TransactionsView
	BuildTransactions(*types.Transaction, *types.Receipt) (Transactions, error)
	StoreTransactions([]Transactions) error
}

type TransactionsView interface {
	TransactionList() ([]Transactions, error)
}

type transactionsDB struct {
	gorm *gorm.DB
}

func NewTransactionsDB(db *gorm.DB) TransactionsDB {
	return &transactionsDB{gorm: db}
}

func (tx transactionsDB) TransactionList() ([]Transactions, error) {
	return nil, nil
}

func (tx transactionsDB) StoreTransactions(transactions []Transactions) error {
	result := tx.gorm.CreateInBatches(&transactions, len(transactions))
	return result.Error
}

func (tx transactionsDB) BuildTransactions(transaction *types.Transaction, transactionReceipt *types.Receipt) (Transactions, error) {
	return Transactions{
		GUID:                 uuid.New(),
		BlockHash:            transactionReceipt.BlockHash,
		BlockNumber:          transactionReceipt.BlockNumber,
		FromAddress:          common.Address{},
		ToAddress:            *transaction.To(),
		Gas:                  big.NewInt(int64(transaction.Gas())),
		GasPrice:             transaction.GasPrice(),
		TransactionHash:      transaction.Hash(),
		InputData:            transaction.Data(),
		MaxFeePerGas:         transaction.GasFeeCap(),
		MaxPriorityFeePerGas: nil,
		GasUsed:              big.NewInt(int64(transactionReceipt.GasUsed)),
		CumulativeGasUsed:    big.NewInt(int64(transactionReceipt.CumulativeGasUsed)),
		EffectiveGasPrice:    transactionReceipt.EffectiveGasPrice,
		L1Fee:                transactionReceipt.L1Fee,
		L1GasUsed:            transactionReceipt.L1GasUsed,
		L1GasPrice:           transactionReceipt.L1GasPrice,
		Nonce:                big.NewInt(int64(transaction.Nonce())),
		TransactionIndex:     big.NewInt(int64(transactionReceipt.TransactionIndex)),
		TxType:               int64(transactionReceipt.Type),
		R:                    nil,
		S:                    nil,
		V:                    nil,
		Status:               int64(transactionReceipt.Status),
		ContractAddress:      transactionReceipt.ContractAddress,
		Amount:               transaction.Value(),
		YParity:              nil,
		Timestamp:            uint64(transaction.Time().Unix()),
	}, nil
}
