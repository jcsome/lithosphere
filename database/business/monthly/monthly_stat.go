package monthly

import (
	"github.com/mantlenetworkio/lithosphere/database/business"
)

type MonthlyStat struct {
	business.NormalStat
}

func (MonthlyStat) TableName() string {
	return "monthly_stat"
}
