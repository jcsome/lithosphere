package weekly

import "github.com/mantlenetworkio/lithosphere/database/business"

type SymbolWeeklyTvl struct {
	business.SymbolTvl
}

func (SymbolWeeklyTvl) TableName() string {
	return "symbol_weekly_tvl"
}
