package models

import "github.com/mantlenetworkio/lithosphere/database/business"

type QueryDWParams struct {
	Address  string
	Page     int
	PageSize int
	Order    string
}

type QueryPageParams struct {
	Page     int
	PageSize int
	Order    string
}

type QueryIdParams struct {
	Id uint64
}

type QueryIndexParams struct {
	Index uint64
}

type DepositsResponse struct {
	Current int               `json:"Current"`
	Size    int               `json:"Size"`
	Total   int64             `json:"Total"`
	Records []business.L1ToL2 `json:"Records"`
}

type WithdrawsResponse struct {
	Current int               `json:"Current"`
	Size    int               `json:"Size"`
	Total   int64             `json:"Total"`
	Records []business.L2ToL1 `json:"Records"`
}

type DataStoreListItem struct {
	ID        uint64 `json:"dataStoreId"`
	DataSize  uint64 `json:"dataSize"`
	Status    bool   `json:"status"`
	Timestamp uint64 `json:"timestamp"`
	DaHash    string `json:"daHash"`
}

type DataStoreList struct {
	ID        uint64 `json:"dataStoreId"`
	DataSize  uint64 `json:"dataSize"`
	Status    bool   `json:"status"`
	Timestamp uint64 `json:"timestamp"`
	DaHash    string `json:"daHash"`
}

type DataStoresResponse struct {
	Current int             `json:"Current"`
	Size    int             `json:"Size"`
	Total   int64           `json:"Total"`
	Records []DataStoreList `json:"Records"`
}

type StateRootListResponse struct {
	Current int                  `json:"Current"`
	Size    int                  `json:"Size"`
	Total   int64                `json:"Total"`
	Records []business.StateRoot `json:"Records"`
}
