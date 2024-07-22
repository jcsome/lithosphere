package business

import (
	"errors"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	common3 "github.com/mantlenetworkio/lithosphere/common"
	common2 "github.com/mantlenetworkio/lithosphere/database/common"
)

type L2ToL1 struct {
	GUID                    uuid.UUID      `gorm:"primaryKey" json:"guid"`
	L1BlockNumber           *big.Int       `gorm:"serializer:u256;column:l1_block_number" db:"l1_block_number" json:"l1BlockNumber" form:"l1_block_number"`
	L2BlockNumber           *big.Int       `gorm:"serializer:u256;column:l2_block_number;primaryKey" db:"l2_block_number" json:"l2BlockNumber" form:"l2_block_number"`
	MsgNonce                *big.Int       `gorm:"column:msg_nonce;serializer:u256" db:"msg_nonce" json:"msgNonce" form:"msg_nonce"`
	L2TransactionHash       common.Hash    `gorm:"column:l2_transaction_hash;serializer:bytes" db:"l2_transaction_hash" json:"l2TransactionHash" form:"l2_transaction_hash"`
	WithdrawTransactionHash common.Hash    `gorm:"column:withdraw_transaction_hash;serializer:bytes" db:"withdraw_transaction_hash" json:"withdrawTransactionHash" form:"withdraw_transaction_hash"`
	L1ProveTxHash           common.Hash    `gorm:"column:l1_prove_tx_hash;serializer:bytes" db:"l1_prove_tx_hash" json:"l1ProveTxHash" form:"l1_prove_tx_hash"`
	L1FinalizeTxHash        common.Hash    `gorm:"column:l1_finalize_tx_hash;serializer:bytes" db:"l1_finalize_tx_hash" json:"l1FinalizeTxHash" form:"l1_finalize_tx_hash"`
	MessageHash             common.Hash    `gorm:"column:message_hash;serializer:bytes" db:"message_hash" json:"messageHash" form:"message_hash"`
	Status                  int64          `gorm:"column:status" db:"status" json:"status" form:"status"`
	FromAddress             common.Address `gorm:"column:from_address;serializer:bytes" db:"from_address" json:"fromAddress" form:"from_address"`
	ETHAmount               *big.Int       `gorm:"serializer:u256;column:eth_amount" json:"ETHAmount"`
	ERC20Amount             *big.Int       `gorm:"serializer:u256;column:erc20_amount" json:"ERC20Amount"`
	GasLimit                *big.Int       `gorm:"serializer:u256;column:gas_limit" json:"gasLimit"`
	TimeLeft                *big.Int       `gorm:"serializer:u256;column:time_left" json:"timeLeft"`
	ToAddress               common.Address `gorm:"column:to_address;serializer:bytes" db:"to_address" json:"toAddress" form:"to_address"`
	L1TokenAddress          common.Address `gorm:"column:l1_token_address;serializer:bytes" db:"l1_token_address" json:"l1TokenAddress" form:"l1_token_address"`
	L2TokenAddress          common.Address `gorm:"column:l2_token_address;serializer:bytes" db:"l2_token_address" json:"l2TokenAddress" form:"l2_token_address"`
	Version                 int64          `gorm:"column:version" json:"version"`
	Timestamp               int64          `gorm:"column:timestamp" db:"timestamp" json:"timestamp" form:"timestamp"`
}

type L2ToL1s []*L2ToL1

func (L2ToL1) TableName() string {
	return "l2_to_l1"
}

type L2ToL1DB interface {
	L2ToL1View
	StoreL2ToL1Transactions([]L2ToL1) error
	UpdateL2ToL1InfoByTxHash(l2L1List []L2ToL1) error
	UpdateL2ToL1InfoByMessageHash(l2L1List []L2ToL1) error
	UpdateReadyForProvedStatus(l2BlockNumber uint64, fraudProofWindows uint64) error
	UpdateTimeLeft() error
	MarkL2ToL1TransactionWithdrawalProven(l2L1List []L2ToL1) error
	MarkL2ToL1TransactionWithdrawalFinalized(l2L1List []L2ToL1) error
	MarkL2ToL1TransactionWithdrawalProvenV0(l2L1List []L2ToL1) error
	MarkL2ToL1TransactionWithdrawalFinalizedV0(l2L1List []L2ToL1) error
	UpdateV1L2Tol1WithdrawalHash(txHash common.Hash, withdrawHash common.Hash) error
	GetWithdrawsUnclaimedAmount(l1FinalizeTxHash string) (L2ToL1s, error)
}

type L2ToL1View interface {
	L2ToL1List(string, int, int, string) ([]L2ToL1, int64)
	L2ToL1TransactionWithdrawal(common.Hash) (*L2ToL1, error)
	GetBlockNumberFromHash(blockHash common.Hash) (*big.Int, error)
	L2L1LatestBlockL2Header() (*common2.L2BlockHeader, error)
	L2L1LatestBlockL1Header() (*common2.L1BlockHeader, error)
	L2L1LatestFinalizedBlockL1Header() (*common2.L1BlockHeader, error)
	L2ToL1TransactionTxHash(common.Hash) (*L2ToL1, error)
	L2L1LatestFinalizedL1BlockNumber() int
	GetWithdrawsClaimedAmount(l1FinalizeTxHash string, startTimestamp int, endTimestamp int) (L2ToL1s, error)
}

type l2ToL1DB struct {
	gorm *gorm.DB
}

func NewL21ToL1DB(db *gorm.DB) L2ToL1DB {
	return &l2ToL1DB{gorm: db}
}

func (l2l1 l2ToL1DB) StoreL2ToL1Transactions(l1L2List []L2ToL1) error {
	result := l2l1.gorm.CreateInBatches(&l1L2List, len(l1L2List))
	return result.Error
}

func (l2l1 l2ToL1DB) L2ToL1List(address string, page int, pageSize int, order string) (l2L1List []L2ToL1, total int64) {
	var totalRecord int64
	var l2ToL1List []L2ToL1
	queryStateRoot := l2l1.gorm.Table("l2_to_l1")
	if address != "0x00" {
		err := l2l1.gorm.Table("l2_to_l1").Select("l2_block_number").Where("from_address = ?", address).Or(" to_address = ?", address).Count(&totalRecord).Error
		if err != nil {
			log.Error("get l2 to l1 by address count fail")
		}
		queryStateRoot.Where("from_address = ?", address).Or(" to_address = ?", address).Offset((page - 1) * pageSize).Limit(pageSize)
	} else {
		err := l2l1.gorm.Table("l2_to_l1").Select("l2_block_number").Count(&totalRecord).Error
		if err != nil {
			log.Error("get l1 to l2 no address count fail ")
		}
		queryStateRoot.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	if strings.ToLower(order) == "asc" {
		queryStateRoot.Order("timestamp asc")
	} else {
		queryStateRoot.Order("timestamp desc")
	}
	qErr := queryStateRoot.Find(&l2ToL1List).Error
	if qErr != nil {
		log.Error("get l2 to l1 list fail", "err", qErr)
	}
	return l2ToL1List, totalRecord
}

func (l2l1 l2ToL1DB) UpdateL2ToL1InfoByTxHash(l2L1List []L2ToL1) error {
	for i := 0; i < len(l2L1List); i++ {
		var l2ToL1 = L2ToL1{}
		result := l2l1.gorm.Where(&L2ToL1{L2TransactionHash: l2L1List[i].L2TransactionHash}).Take(&l2ToL1)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			return result.Error
		}
		l2ToL1.L1BlockNumber = l2L1List[i].L1BlockNumber
		l2ToL1.FromAddress = l2L1List[i].FromAddress
		l2ToL1.ToAddress = l2L1List[i].ToAddress
		l2ToL1.ETHAmount = l2L1List[i].ETHAmount
		l2ToL1.ERC20Amount = l2L1List[i].ERC20Amount
		l2ToL1.L1TokenAddress = l2L1List[i].L1TokenAddress
		l2ToL1.L2TokenAddress = l2L1List[i].L2TokenAddress
		err := l2l1.gorm.Save(&l2ToL1).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (l2l1 l2ToL1DB) UpdateL2ToL1InfoByMessageHash(l2L1List []L2ToL1) error {
	for i := 0; i < len(l2L1List); i++ {
		var l2ToL1 = L2ToL1{}
		result := l2l1.gorm.Where(&L2ToL1{MessageHash: l2L1List[i].MessageHash}).Take(&l2ToL1)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			return result.Error
		}
		l2ToL1.L1BlockNumber = l2L1List[i].L1BlockNumber
		l2ToL1.FromAddress = l2L1List[i].FromAddress
		l2ToL1.ToAddress = l2L1List[i].ToAddress
		l2ToL1.ETHAmount = l2L1List[i].ETHAmount
		l2ToL1.ERC20Amount = l2L1List[i].ERC20Amount
		l2ToL1.L1TokenAddress = l2L1List[i].L1TokenAddress
		l2ToL1.L2TokenAddress = l2L1List[i].L2TokenAddress
		err := l2l1.gorm.Save(&l2ToL1).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (l2l1 l2ToL1DB) MarkL2ToL1TransactionWithdrawalProven(l2L1List []L2ToL1) error {
	for i := 0; i < len(l2L1List); i++ {
		var l2ToL1 = L2ToL1{}
		if l2L1List[i].L1BlockNumber.Uint64() <= 0 {
			continue
		}
		result := l2l1.gorm.Where(&L2ToL1{WithdrawTransactionHash: l2L1List[i].WithdrawTransactionHash}).Take(&l2ToL1)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			return result.Error
		}
		log.Info("mark transaction prove", "L1BlockNumber",
			l2L1List[i].L1BlockNumber, "L1ProveTxHash", l2L1List[i].L1ProveTxHash, "WithdrawTransactionHash",
			l2L1List[i].WithdrawTransactionHash, "l2ToL1WithdrawTransactionHash", l2ToL1.WithdrawTransactionHash)
		l2ToL1.L1BlockNumber = l2L1List[i].L1BlockNumber
		l2ToL1.L1ProveTxHash = l2L1List[i].L1ProveTxHash
		if l2ToL1.TimeLeft.Uint64() > 0 {
			l2ToL1.Status = common3.L2ToL1InChallengePeriod // in challenge period
		} else {
			l2ToL1.Status = common3.L2ToL1ReadyForClaim // ready for claim
		}
		err := l2l1.gorm.Save(&l2ToL1).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (l2l1 l2ToL1DB) MarkL2ToL1TransactionWithdrawalProvenV0(l2L1List []L2ToL1) error {
	for i := 0; i < len(l2L1List); i++ {
		var l2ToL1 = L2ToL1{}
		if l2L1List[i].L1BlockNumber.Uint64() <= 0 {
			continue
		}
		result := l2l1.gorm.Where(&L2ToL1{MessageHash: l2L1List[i].MessageHash}).Take(&l2ToL1)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			return result.Error
		}
		log.Info("mark transaction v0 prove", "L1BlockNumber",
			l2L1List[i].L1BlockNumber, "L1ProveTxHash", l2L1List[i].L1ProveTxHash,
			"WithdrawTransactionHash", l2L1List[i].WithdrawTransactionHash)
		l2ToL1.L1BlockNumber = l2L1List[i].L1BlockNumber
		l2ToL1.L1ProveTxHash = l2L1List[i].L1ProveTxHash
		if l2ToL1.TimeLeft.Uint64() > 0 {
			l2ToL1.Status = common3.L2ToL1InChallengePeriod // in challenge period
		} else {
			l2ToL1.Status = common3.L2ToL1ReadyForClaim // ready for claim
		}
		err := l2l1.gorm.Save(&l2ToL1).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (l2l1 l2ToL1DB) MarkL2ToL1TransactionWithdrawalFinalized(l2L1List []L2ToL1) error {
	for i := 0; i < len(l2L1List); i++ {
		var l2ToL1 = L2ToL1{}
		if l2L1List[i].L1BlockNumber.Uint64() <= 0 {
			continue
		}
		result := l2l1.gorm.Where(&L2ToL1{WithdrawTransactionHash: l2L1List[i].WithdrawTransactionHash}).Take(&l2ToL1)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			return result.Error
		}
		log.Info("mark transaction finalized", "L1BlockNumber",
			l2L1List[i].L1BlockNumber, "L1FinalizeTxHash", l2L1List[i].L1FinalizeTxHash,
			"WithdrawTransactionHash", l2L1List[i].WithdrawTransactionHash, "l2ToL1WithdrawHash", l2ToL1.WithdrawTransactionHash)
		l2ToL1.L1BlockNumber = l2L1List[i].L1BlockNumber
		l2ToL1.L1FinalizeTxHash = l2L1List[i].L1FinalizeTxHash
		l2ToL1.Status = common3.L2ToL1Claimed // relayed
		err := l2l1.gorm.Save(&l2ToL1).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (l2l1 l2ToL1DB) MarkL2ToL1TransactionWithdrawalFinalizedV0(l2L1List []L2ToL1) error {
	for i := 0; i < len(l2L1List); i++ {
		var l2ToL1 = L2ToL1{}
		if l2L1List[i].L1BlockNumber.Uint64() <= 0 {
			continue
		}
		result := l2l1.gorm.Where(&L2ToL1{MessageHash: l2L1List[i].MessageHash}).Take(&l2ToL1)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return nil
			}
			return result.Error
		}
		log.Info("mark transaction v0 finalized",
			"L1BlockNumber", l2L1List[i].L1BlockNumber, "L1FinalizeTxHash", l2L1List[i].L1FinalizeTxHash,
			"WithdrawTransactionHash", l2L1List[i].WithdrawTransactionHash)
		l2ToL1.L1BlockNumber = l2L1List[i].L1BlockNumber
		l2ToL1.L1FinalizeTxHash = l2L1List[i].L1FinalizeTxHash
		l2ToL1.Status = common3.L2ToL1Claimed // relayed
		err := l2l1.gorm.Save(&l2ToL1).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (l2l1 l2ToL1DB) UpdateReadyForProvedStatus(l2BlockNumber uint64, fraudProofWindows uint64) error {
	var l2ToL1 = L2ToL1{}
	err := l2l1.gorm.Model(&l2ToL1).Where("l2_block_number <= ? AND status = ?", l2BlockNumber, 0).Updates(map[string]interface{}{"status": common3.L2ToL1ReadyForProved, "time_left": fraudProofWindows}).Error
	if err != nil {
		return err
	}
	return nil
}

func (l2l1 l2ToL1DB) UpdateTimeLeft() error {
	result := l2l1.gorm.Model(&L2ToL1{}).Where("time_left > ? and status = ?", 0, common3.L2ToL1InChallengePeriod).Updates(map[string]interface{}{"time_left": gorm.Expr("GREATEST(time_left - 1, 0)")})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return result.Error
	}
	result = l2l1.gorm.Model(&L2ToL1{}).Where("time_left = ? and status = ?", 0, common3.L2ToL1InChallengePeriod).Updates(map[string]interface{}{"status": common3.L2ToL1ReadyForClaim})
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return result.Error
	}
	return nil
}

func (l2l1 l2ToL1DB) GetBlockNumberFromHash(blockHash common.Hash) (*big.Int, error) {
	var l2BlockNumber uint64
	result := l2l1.gorm.Table("l2_block_headers").Where("hash = ?", blockHash.String()).Select("number").Take(&l2BlockNumber)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return new(big.Int).SetUint64(l2BlockNumber), nil
}

func (l2l1 l2ToL1DB) L2ToL1TransactionWithdrawal(withdrawalHash common.Hash) (*L2ToL1, error) {
	var l2ToL1Withdrawal L2ToL1
	result := l2l1.gorm.Where(&L2ToL1{WithdrawTransactionHash: withdrawalHash}).Take(&l2ToL1Withdrawal)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &l2ToL1Withdrawal, nil
}

func (l2l1 l2ToL1DB) L2ToL1TransactionTxHash(txHash common.Hash) (*L2ToL1, error) {
	var l2ToL1Withdrawal L2ToL1
	result := l2l1.gorm.Where(&L2ToL1{L2TransactionHash: txHash}).Take(&l2ToL1Withdrawal)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &l2ToL1Withdrawal, nil
}

func (l2l1 l2ToL1DB) L2L1LatestBlockL2Header() (*common2.L2BlockHeader, error) {
	l2Query := l2l1.gorm.Where("number = (?)", l2l1.gorm.Table("l2_to_l1").Select("MAX(l2_block_number)"))
	var l2Header common2.L2BlockHeader
	result := l2Query.Take(&l2Header)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &l2Header, nil
}

func (l2l1 l2ToL1DB) L2L1LatestBlockL1Header() (*common2.L1BlockHeader, error) {
	var l1Header common2.L1BlockHeader
	result := l2l1.gorm.Table("l1_block_headers").Order("number desc").Limit(1).Take(&l1Header)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &l1Header, nil
}

func (l2l1 l2ToL1DB) L2L1LatestFinalizedBlockL1Header() (*common2.L1BlockHeader, error) {
	l1Query := l2l1.gorm.Where("number = (?)", l2l1.gorm.Table("l2_to_l1").Where("status = (?)", common3.L2ToL1Claimed).Select("MAX(l1_block_number)"))
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

func (l2l1 l2ToL1DB) UpdateV1L2Tol1WithdrawalHash(txHash common.Hash, withdrawHash common.Hash) error {
	var l2ToL1 = L2ToL1{}
	err := l2l1.gorm.Model(&l2ToL1).Where("l2_transaction_hash = ?", txHash.String()).Updates(map[string]interface{}{"withdraw_transaction_hash": withdrawHash.String()}).Error
	if err != nil {
		return err
	}
	return nil
}

func (l2l1 *l2ToL1DB) L2L1LatestFinalizedL1BlockNumber() int {
	var timestamp int
	Query := l2l1.gorm.Table("l2_to_l1").Select("l1_block_number").Where("l1_block_number > ?", 0).Order("l1_block_number desc").Limit(1)
	Query.Take(&timestamp)

	return timestamp
}

func (l2l1 *l2ToL1DB) GetWithdrawsUnclaimedAmount(l1FinalizeTxHash string) (L2ToL1s, error) {
	var l2Tol1s L2ToL1s
	Query := l2l1.gorm.Table("l2_to_l1").Select("sum(eth_amount) as eth_amount,sum(erc20_amount) as erc20_amount,l1_token_address,l2_token_address").Where("l1_finalize_tx_hash = ?", l1FinalizeTxHash).Group("l1_token_address,l2_token_address")
	err := Query.Find(&l2Tol1s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return l2Tol1s, nil
}

func (l2l1 *l2ToL1DB) GetWithdrawsClaimedAmount(l1FinalizeTxHash string, startBlockNumber int, endBlockNumber int) (L2ToL1s, error) {
	var l2Tol1s L2ToL1s
	Query := l2l1.gorm.Table("l2_to_l1").Select("sum(eth_amount) as eth_amount,sum(erc20_amount) as erc20_amount,l1_token_address,l2_token_address").Where("l1_finalize_tx_hash != ? and l1_block_number > ? and l1_block_number <= ?", l1FinalizeTxHash, startBlockNumber, endBlockNumber).Group("l1_token_address,l2_token_address")
	err := Query.Find(&l2Tol1s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return l2Tol1s, nil
}
