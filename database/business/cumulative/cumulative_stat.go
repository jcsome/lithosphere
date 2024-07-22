package cumulative

import (
	"github.com/mantlenetworkio/lithosphere/database/business"
)

type CumulativeStat struct {
	business.NormalStat
}

func (CumulativeStat) TableName() string {
	return "cumulative_stat"
}
