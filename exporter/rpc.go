package exporter

type GetRollupGasPrices struct {
	L1GasPrice string `json:"l1GasPrice"`
	L2GasPrice string `json:"l2GasPrice"`
}

type GetBlockNumber struct {
	BlockNumber string `json:"result"`
}
