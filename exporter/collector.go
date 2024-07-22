package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	gasBaseFeeMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "op_baseFee",
			Help: "Gas base fee."},
		[]string{"network", "layer"},
	)
	gasUsedMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "op_gasUsed",
			Help: "Gas Used."},
		[]string{"network", "layer"},
	)
	gasPrice = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "op_gasPrice",
			Help: "Gas price."},
		[]string{"network", "layer"},
	)
	blockNumber = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "op_blocknumber",
			Help: "Current block number."},
		[]string{"network", "layer"},
	)
	healthySequencer = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "op_healthy_sequencer",
			Help: "Is the sequencer healthy?"},
		[]string{"network"},
	)
	opExporterVersion = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "op_exporter_version",
			Help: "Verion of the op-exporter software"},
		[]string{"version", "commit", "goVersion", "buildDate"},
	)
	bridgeAccountChecking = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bridge_account_checking",
			Help: ""},
		[]string{"l1Token", "l2Token", "symbol"},
	)
	depositsAmount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lithosphere_deposit_amount",
			Help: "The total amount of deposits"},
		[]string{"token_address", "symbol"},
	)
	withdrawsUnclaimedAmount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lithosphere_withdraws_unclaimed_amount",
			Help: "The total amount of withdraws unclaimed"},
		[]string{"token_address", "symbol"},
	)
	withdrawsClaimedAmount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lithosphere_withdraws_claimed_amount",
			Help: "The total amount of withdraws claimed"},
		[]string{"token_address", "symbol"},
	)
	transferBigValueInEthereum = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lithosphere_transfer_big_value_in_ethereum",
			Help: "The count of bigValue transfer"},
		[]string{"token_address", "symbol"},
	)
	transferBigValueInMantle = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lithosphere_transfer_big_value_in_mantle",
			Help: "The count of bigValue transfer"},
		[]string{"token_address", "symbol"},
	)
)

func init() {
	prometheus.MustRegister(gasPrice)
	prometheus.MustRegister(blockNumber)
	prometheus.MustRegister(healthySequencer)
	prometheus.MustRegister(opExporterVersion)
	prometheus.MustRegister(gasBaseFeeMetric)
	prometheus.MustRegister(gasUsedMetric)
	prometheus.MustRegister(bridgeAccountChecking)
	prometheus.MustRegister(depositsAmount)
	prometheus.MustRegister(withdrawsUnclaimedAmount)
	prometheus.MustRegister(withdrawsClaimedAmount)
	prometheus.MustRegister(transferBigValueInEthereum)
	prometheus.MustRegister(transferBigValueInMantle)
}
