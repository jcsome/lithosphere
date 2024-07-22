package cumulative

import "github.com/mantlenetworkio/lithosphere/database/business"

type SymbolCumulativeTvl struct {
	business.SymbolTvl
}

func (SymbolCumulativeTvl) TableName() string {
	return "symbol_cumulative_tvl"
}
