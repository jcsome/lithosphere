package business

import (
	"strings"

	"gorm.io/gorm"
)

// TokenList represents bridge token list
type TokenList struct {
	ID        uint64 `gorm:"primary_key;column:id"`
	ChainID   uint64 `json:"chainId" gorm:"index:idx_token,unique;column:chain_id"`
	Address   string `json:"address" gorm:"index:idx_token,unique,type:varchar(42);column:address"`
	Name      string `json:"name" gorm:"column:name"`
	Symbol    string `json:"symbol" gorm:"column:symbol"`
	Decimals  uint64 `json:"decimals" gorm:"column:decimals"`
	Timestamp uint64 `json:"timestamp" gorm:"column:timestamp"`
}

func (TokenList) TableName() string {
	return "token_lists"
}

type TokenListDB interface {
	TokenListView
	SaveTokenList(list TokenList) error
}

type TokenListView interface {
	GetSymbolByAddress(address string) (string, error)
}

type tokenListDB struct {
	gorm *gorm.DB
}

func NewTokenListDB(db *gorm.DB) TokenListDB {
	return &tokenListDB{gorm: db}
}

func (tl tokenListDB) SaveTokenList(list TokenList) error {
	tokenlist := TokenList{}
	list.Address = strings.ToLower(list.Address)
	tl.gorm.Table("token_lists").Where("name = ? and chain_id = ?", list.Name, list.ChainID).Find(&tokenlist)
	if tokenlist.ID != 0 {
		result := tl.gorm.Table("token_lists").Where("name = ? and chain_id = ?", list.Name, list.ChainID).Updates(map[string]interface{}{"address": list.Address, "symbol": list.Symbol, "decimals": list.Decimals, "timestamp": list.Timestamp})
		return result.Error
	}
	result := tl.gorm.Table("token_lists").Where("name = ? and chain_id = ?", list.Name, list.ChainID).Save(&list)
	return result.Error
}

func (tl tokenListDB) GetSymbolByAddress(address string) (string, error) {
	address = strings.ToLower(address)
	symbol := ""
	result := tl.gorm.Table("token_lists").Where("address = ?", address).Select("symbol").Take(&symbol)
	return symbol, result.Error
}
