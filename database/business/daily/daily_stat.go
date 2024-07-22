package daily

import (
	"github.com/mantlenetworkio/lithosphere/database/business"
)

type DailyStat struct {
	business.NormalStat
}

func (DailyStat) TableName() string {
	return "daily_stat"
}
