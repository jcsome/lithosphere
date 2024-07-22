package daily

import "github.com/mantlenetworkio/lithosphere/database/business"

type ProtocolDailyTvl struct {
	business.ProtocolTvl
}

func (ProtocolDailyTvl) TableName() string {
	return "protocol_daily_tvl"
}
