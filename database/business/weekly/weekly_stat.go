package weekly

import (
	"github.com/mantlenetworkio/lithosphere/database/business"
)

type WeeklyStat struct {
	business.NormalStat
}

func (WeeklyStat) TableName() string {
	return "weekly_stat"
}
