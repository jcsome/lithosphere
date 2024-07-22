package weekly

import "github.com/mantlenetworkio/lithosphere/database/business"

type ProtocolWeeklyTvl struct {
	business.ProtocolTvl
}

func (ProtocolWeeklyTvl) TableName() string {
	return "protocol_weekly_tvl"
}
