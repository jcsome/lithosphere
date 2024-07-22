package daily

import "github.com/mantlenetworkio/lithosphere/database/business"

type SymbolDailyTvl struct {
	business.SymbolTvl
}

func (SymbolDailyTvl) TableName() string {
	return "symbol_daily_tvl"
}
