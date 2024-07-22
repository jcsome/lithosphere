package monthly

import "github.com/mantlenetworkio/lithosphere/database/business"

type ProtocolMonthlyTvl struct {
	business.ProtocolTvl
}

func (ProtocolMonthlyTvl) TableName() string {
	return "protocol_monthly_tvl"
}
