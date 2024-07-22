package cumulative

import "github.com/mantlenetworkio/lithosphere/database/business"

type ProtocolCumulativeTvl struct {
	business.ProtocolTvl
}

func (ProtocolCumulativeTvl) TableName() string {
	return "protocol_cumulative_tvl"
}
