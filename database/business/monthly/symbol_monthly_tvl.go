package monthly

import "github.com/mantlenetworkio/lithosphere/database/business"

type SymbolMonthlyTvl struct {
	business.SymbolTvl
}

func (SymbolMonthlyTvl) TableName() string {
	return "symbol_monthly_tvl"
}
