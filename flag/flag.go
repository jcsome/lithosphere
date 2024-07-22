package flag

import (
	"time"

	"github.com/urfave/cli/v2"
)

const envVarPrefix = "LITHOSPHERE"

func prefixEnvVars(name string) []string {
	return []string{envVarPrefix + "_" + name}
}

var (
	// Required flags
	MigrationsFlag = &cli.StringFlag{
		Name:    "migrations-dir",
		Value:   "./migrations",
		Usage:   "path to migrations folder",
		EnvVars: prefixEnvVars("MIGRATIONS_DIR"),
	}
	L1EthRpcFlag = &cli.StringFlag{
		Name:     "l1-eth-rpc",
		Usage:    "HTTP provider URL for L1",
		EnvVars:  prefixEnvVars("L1_RPC"),
		Required: true,
	}
	L2EthRpcFlag = &cli.StringFlag{
		Name:     "l2-eth-rpc",
		Usage:    "HTTP provider URL for L2",
		EnvVars:  prefixEnvVars("L2_PRC"),
		Required: true,
	}
	HttpHostFlag = &cli.StringFlag{
		Name:     "http-host",
		Usage:    "The host of the api",
		EnvVars:  prefixEnvVars("HTTP_HOST"),
		Required: true,
	}
	HttpPortFlag = &cli.IntFlag{
		Name:     "http-port",
		Usage:    "The port of the api",
		EnvVars:  prefixEnvVars("HTTP_PORT"),
		Value:    8987,
		Required: true,
	}
	MetricsHostFlag = &cli.StringFlag{
		Name:     "metrics-host",
		Usage:    "The host of the metrics",
		EnvVars:  prefixEnvVars("METRICS_HOST"),
		Required: true,
	}
	MetricsPortFlag = &cli.IntFlag{
		Name:     "metrics-port",
		Usage:    "The port of the metrics",
		EnvVars:  prefixEnvVars("METRICS_PORT"),
		Value:    7214,
		Required: true,
	}
	SlaveDbEnableFlag = &cli.BoolFlag{
		Name:     "slave-db-enable",
		Usage:    "Whether to use slave db",
		EnvVars:  prefixEnvVars("SLAVE_DB_ENABLE"),
		Required: true,
	}
	MasterDbHostFlag = &cli.StringFlag{
		Name:     "master-db-host",
		Usage:    "The host of the master database",
		EnvVars:  prefixEnvVars("MASTER_DB_HOST"),
		Required: true,
	}
	MasterDbPortFlag = &cli.IntFlag{
		Name:     "master-db-port",
		Usage:    "The port of the master database",
		EnvVars:  prefixEnvVars("MASTER_DB_PORT"),
		Required: true,
	}
	MasterDbUserFlag = &cli.StringFlag{
		Name:     "master-db-user",
		Usage:    "The user of the master database",
		EnvVars:  prefixEnvVars("MASTER_DB_USER"),
		Required: true,
	}
	MasterDbPasswordFlag = &cli.StringFlag{
		Name:     "master-db-password",
		Usage:    "The host of the master database",
		EnvVars:  prefixEnvVars("MASTER_DB_PASSWORD"),
		Required: true,
	}
	MasterDbNameFlag = &cli.StringFlag{
		Name:     "master-db-name",
		Usage:    "The db name of the master database",
		EnvVars:  prefixEnvVars("MASTER_DB_NAME"),
		Required: true,
	}
	StartDataStoreIdFlag = &cli.Uint64Flag{
		Name:    "start-data-store-id",
		Usage:   "Start data store id",
		EnvVars: prefixEnvVars("START_DATA_STORE_ID"),
		Value:   1,
	}
	// Optional flags
	L1PollingIntervalFlag = &cli.DurationFlag{
		Name:    "l1-polling-interval",
		Usage:   "The interval of l1 synchronization",
		EnvVars: prefixEnvVars("L1_POLLING_INTERVAL"),
		Value:   0,
	}
	L1HeaderBufferSizeFlag = &cli.IntFlag{
		Name:    "l1-header-buffer-size",
		Usage:   "The buffer size of l1 header",
		EnvVars: prefixEnvVars("L1_HEADER_BUFFER_SIZE"),
		Value:   0,
	}
	L1ConfirmationDepthFlag = &cli.IntFlag{
		Name:    "l1-confirmation-depth",
		Usage:   "The confirmation depth of l1",
		EnvVars: prefixEnvVars("L1_CONFIRMATION_DEPTH"),
		Value:   0,
	}
	L1StartingHeightFlag = &cli.IntFlag{
		Name:    "l1-starting-height",
		Usage:   "The starting height of l1",
		EnvVars: prefixEnvVars("L1_STARTING_HEIGHT"),
		Value:   0,
	}
	L2StartingHeightFlag = &cli.IntFlag{
		Name:    "l2-starting-height",
		Usage:   "The starting height of l2",
		EnvVars: prefixEnvVars("L2_STARTING_HEIGHT"),
		Value:   0,
	}
	AddressManagerFlag = &cli.StringFlag{
		Name:     "address-manager",
		Usage:    "contracts address manager.",
		Required: true,
		EnvVars:  prefixEnvVars("ADDRESS_MANAGER"),
	}
	SystemConfigProxyFlag = &cli.StringFlag{
		Name:     "system-config-proxy-address",
		Usage:    "The system config proxy address.",
		Required: true,
		EnvVars:  prefixEnvVars("SYSTEM_CONFIG_PROXY_ADDRESS"),
	}
	OptimismPortalProxyFlag = &cli.StringFlag{
		Name:     "optimism-portal-proxy-address",
		Usage:    "The optimism portal proxy address.",
		Required: true,
		EnvVars:  prefixEnvVars("OPTIMISM_PORTAL_PROXY_ADDRESS"),
	}
	L2OutputOracleProxyFlag = &cli.StringFlag{
		Name:     "l2-output-oracle-proxy-address",
		Usage:    "L2 output oracle proxy contracts address.",
		Required: true,
		EnvVars:  prefixEnvVars("L2_OUTPUT_ORACLE_PROXY_ADDRESS"),
	}
	L1CrossDomainMessengerProxyFlag = &cli.StringFlag{
		Name:     "l1-crossdomain-message-proxy-address",
		Usage:    "L1 cross domain message contract address.",
		Required: true,
		EnvVars:  prefixEnvVars("L1_CROSSDOMAIN_MESSAGE_PROXY_ADDRESS"),
	}
	L1StandardBridgeProxyFlag = &cli.StringFlag{
		Name:     "l2-standard-bridge-proxy-address",
		Usage:    "L1 standard bridge proxy contract address.",
		Required: true,
		EnvVars:  prefixEnvVars("L1_STANDARD_BRIDGE_PROXY_ADDRESS"),
	}
	L1ERC721BridgeProxyFlag = &cli.StringFlag{
		Name:     "l1-erc721-bridge-proxy-address",
		Usage:    "L1 Erc721 proxy contract address.",
		Required: true,
		EnvVars:  prefixEnvVars("L1_ERC721_BRIDGE_PROXY_ADDRESS"),
	}
	DataLayrServiceManagerAddrFlag = &cli.StringFlag{
		Name:     "mantle-da-dlsm-address",
		Usage:    "The MantleDA data layr service manager address.",
		Required: true,
		EnvVars:  prefixEnvVars("MANTLE_DA_DLSM_ADDRESS"),
	}
	LegacyCanonicalTransactionChainFlag = &cli.StringFlag{
		Name:     "legacy-ctc-address",
		Usage:    "The ctc contract address of ovm",
		Required: true,
		EnvVars:  prefixEnvVars("LEGACY_CTC_ADDRESS"),
	}
	LegacyStateCommitmentChainFlag = &cli.StringFlag{
		Name:     "legacy-scc-address",
		Usage:    "The scc contracts address of ovm",
		Required: true,
		EnvVars:  prefixEnvVars("LEGACY_SCC_ADDRESS"),
	}
	L1BedrockStartingHeightFlag = &cli.IntFlag{
		Name:    "l1-bedrock-starting-height",
		Usage:   "The starting height of l1 upgrade to bedrock",
		EnvVars: prefixEnvVars("L1_BEDROCK_STARTING_HEIGHT"),
		Value:   0,
	}
	L2BedrockStartingHeightFlag = &cli.IntFlag{
		Name:    "l2-bedrock-starting-height",
		Usage:   "The starting height of l2 upgrade to bedrock",
		EnvVars: prefixEnvVars("L2_BEDROCK_STARTING_HEIGHT"),
		Value:   0,
	}
	L2PollingIntervalFlag = &cli.DurationFlag{
		Name:    "l2-polling-interval",
		Usage:   "The interval of l2 synchronization",
		EnvVars: prefixEnvVars("L2_POLLING_INTERVAL"),
		Value:   0,
	}
	L2HeaderBufferSizeFlag = &cli.IntFlag{
		Name:    "l2-header-buffer-size",
		Usage:   "The buffer size of l2 header",
		EnvVars: prefixEnvVars("L2_HEADER_BUFFER_SIZE"),
		Value:   500,
	}
	L2ConfirmationDepthFlag = &cli.IntFlag{
		Name:    "l2-confirmation-depth",
		Usage:   "The confirmation depth of l2",
		EnvVars: prefixEnvVars("L2_CONFIRMATION_DEPTH"),
		Value:   0,
	}
	RetrieverSocketFlag = &cli.StringFlag{
		Name:    "retriever-socket",
		Usage:   "Websocket for MantleDA disperser",
		EnvVars: prefixEnvVars("RETRIEVER_SOCKET"),
	}
	RetrieverTimeoutFlag = &cli.DurationFlag{
		Name:    "retriever-timeout",
		Usage:   "Retriever timeout",
		EnvVars: prefixEnvVars("RETRIEVER_TIMEOUT"),
	}
	FraudProofWindowsFlags = &cli.Uint64Flag{
		Name:    "fraud-proof-windows",
		Usage:   "The fraud proof windows",
		Value:   1800,
		EnvVars: prefixEnvVars("FRAUD_PROOF_WINDOWS"),
	}
	DataStorePollingDurationFlag = &cli.DurationFlag{
		Name:    "data-store-polling-duration",
		Usage:   "Duration to store blob",
		EnvVars: prefixEnvVars("DATA_STORE_POLLING_DURATION"),
	}
	GraphProviderFlag = &cli.StringFlag{
		Name:    "graph-provider",
		Usage:   "Graph node url of MantleDA graph node",
		EnvVars: prefixEnvVars("GRAPH_PROVIDER"),
	}
	SlaveDbHostFlag = &cli.StringFlag{
		Name:    "slave-db-host",
		Usage:   "The host of the slave database",
		EnvVars: prefixEnvVars("SLAVE_DB_HOST"),
	}
	SlaveDbPortFlag = &cli.IntFlag{
		Name:    "slave-db-port",
		Usage:   "The port of the slave database",
		EnvVars: prefixEnvVars("SLAVE_DB_PORT"),
	}
	SlaveDbUserFlag = &cli.StringFlag{
		Name:    "slave-db-user",
		Usage:   "The user of the slave database",
		EnvVars: prefixEnvVars("SLAVE_DB_USER"),
	}
	SlaveDbPasswordFlag = &cli.StringFlag{
		Name:    "slave-db-password",
		Usage:   "The host of the slave database",
		EnvVars: prefixEnvVars("SLAVE_DB_PASSWORD"),
	}
	SlaveDbNameFlag = &cli.StringFlag{
		Name:    "slave-db-name",
		Usage:   "The db name of the slave database",
		EnvVars: prefixEnvVars("SLAVE_DB_NAME"),
	}
	ExporterAddressFlag = &cli.StringFlag{
		Name:    "exporter-address",
		Usage:   "Address on which to exporter metrics and web interface.",
		Value:   ":9100",
		EnvVars: prefixEnvVars("EXPORTER_ADDRESS"),
	}
	NetworkLabelFlag = &cli.StringFlag{
		Name:    "network-label",
		Usage:   "Label to apply to the metrics to identify the network.",
		Value:   "mainnet",
		EnvVars: prefixEnvVars("NETWORK_LABEL"),
	}
	VersionEnableFlag = &cli.BoolFlag{
		Name:    "version",
		Usage:   "Display binary version.",
		Value:   false,
		EnvVars: prefixEnvVars("VERSION_ENABLE"),
	}
	UnhealthyTimePeriodFlag = &cli.DurationFlag{
		Name:    "unhealthy-time-period",
		Usage:   "Number of minutes to wait for the next block before marking provider unhealthy.",
		Value:   120 * time.Second,
		EnvVars: prefixEnvVars("UNHEALTHY_TIME_PERIOD"),
	}
	SequencerPollingSecondsFlag = &cli.DurationFlag{
		Name:    "sequencer-polling-seconds",
		Usage:   "Number of seconds to wait between sequencer polling cycles.",
		Value:   30 * time.Second,
		EnvVars: prefixEnvVars("SEQUENCER_POLLING_SECONDS"),
	}
	EnableK8sQueryFlag = &cli.BoolFlag{
		Name:    "k8s-enable",
		Usage:   "Enable kubernetes info lookup.",
		Value:   false,
		EnvVars: prefixEnvVars("K8S_ENABLE_QUERY"),
	}
	EnableRollUpGasPricesFlag = &cli.BoolFlag{
		Name:    "rollup-gas-prices-enable",
		Usage:   "Enable rollUpGasPrices info lookup.",
		Value:   false,
		EnvVars: prefixEnvVars("ROLLUP_GAS_PRICES_ENABLE"),
	}
	EnableGasBaseFeeFlag = &cli.BoolFlag{
		Name:    "gas-base-fee-enable",
		Usage:   "Enable gaseBaseFee info lookup.",
		Value:   false,
		EnvVars: prefixEnvVars("GAS_BASE_FEE_ENABLE"),
	}
	EnableApiCacheFlag = &cli.BoolFlag{
		Name:    "api-cache-enable",
		Usage:   "Enable api cache.",
		Value:   false,
		EnvVars: prefixEnvVars("API_CACHE_ENABLE"),
	}
	ApiCacheListSize = &cli.IntFlag{
		Name:    "api-cache-list-size",
		Usage:   "Enable api cache.",
		Value:   1200000,
		EnvVars: prefixEnvVars("API_CACHE_LIST_SIZE"),
	}
	ApiCacheDetailSize = &cli.IntFlag{
		Name:    "api-cache-detail-size",
		Usage:   "Enable api cache.",
		Value:   120000,
		EnvVars: prefixEnvVars("API_CACHE_DETAIL_SIZE"),
	}
	ApiCacheListExpireTime = &cli.DurationFlag{
		Name:    "api-cache-list-expire-time",
		Usage:   "Enable api cache.",
		Value:   2 * time.Second,
		EnvVars: prefixEnvVars("API_CACHE_LIST_EXPIRE_TIME"),
	}
	ApiCacheDetailExpireTime = &cli.DurationFlag{
		Name:    "api-cache-detail-expire-time",
		Usage:   "Enable api cache.",
		Value:   60 * time.Second,
		EnvVars: prefixEnvVars("API_CACHE_DETAIL_EXPIRE_TIME"),
	}
	L1AccountCheckingAddressFlag = &cli.StringFlag{
		Name:    "l1-account-checking-address",
		Usage:   "The l1 token address that needs to be reconciled.",
		Value:   "",
		EnvVars: prefixEnvVars("L1_ACCOUNT_CHECKING_ADDRESS"),
	}
	L2AccountCheckingAddressFlag = &cli.StringFlag{
		Name:    "l2-account-checking-address",
		Usage:   "The l2 token address that needs to be reconciled.",
		Value:   "",
		EnvVars: prefixEnvVars("L2_ACCOUNT_CHECKING_ADDRESS"),
	}
	EnableWithdrawCalcFlag = &cli.BoolFlag{
		Name:    "withdraw-calc-enable",
		Usage:   "Enable withdraw calc.",
		Value:   false,
		EnvVars: prefixEnvVars("WITHDRAW_CALC_ENABLE"),
	}
	ChainIdFlag = &cli.IntFlag{
		Name:    "chain-id",
		Usage:   "The id of chain.",
		Value:   5003,
		EnvVars: prefixEnvVars("CHAIN_ID"),
	}
	TransferBigValueAddressInEthereumFlag = &cli.StringFlag{
		Name:    "transfer-big-value-address-in-ethereum",
		Usage:   "The monitored large transfer contract address in ethereum.",
		Value:   "",
		EnvVars: prefixEnvVars("TRANSFER_BIG_VALUE_ADDRESS_IN_ETHEREUM"),
	}
	TransferBigValueAddressInMantleFlag = &cli.StringFlag{
		Name:    "transfer-big-value-address-in-mantle",
		Usage:   "The monitored large transfer contract address in mantle.",
		Value:   "",
		EnvVars: prefixEnvVars("TRANSFER_BIG_VALUE_ADDRESS_IN_MANTLE"),
	}
	TransferBigValueInEthereumFlag = &cli.StringFlag{
		Name:    "transfer-big-value-in-ethereum",
		Usage:   "The monitored large transfer value address in ethereum.",
		Value:   "",
		EnvVars: prefixEnvVars("TRANSFER_BIG_VALUE_IN_ETHEREUM"),
	}
	TransferBigValueInMantleFlag = &cli.StringFlag{
		Name:    "transfer-big-value-in-mantle",
		Usage:   "The monitored large transfer value address in mantle.",
		Value:   "",
		EnvVars: prefixEnvVars("TRANSFER_BIG_VALUE_IN_MANTLE"),
	}
	WithdrawBigValueAddressFlag = &cli.StringFlag{
		Name:    "withdraw-big-value-address",
		Usage:   "The monitored withdraw big value address.",
		Value:   "",
		EnvVars: prefixEnvVars("WITHDRAW_BIG_VALUE_ADDRESS"),
	}
	TokenListUrlFlag = &cli.StringFlag{
		Name:    "token-list-url",
		Usage:   "The url of token list.",
		Value:   "",
		EnvVars: prefixEnvVars("TOKEN_LIST_URL"),
	}
)

var requiredFlags = []cli.Flag{
	MigrationsFlag,
	L1EthRpcFlag,
	L2EthRpcFlag,
	HttpPortFlag,
	HttpHostFlag,
	MetricsPortFlag,
	MetricsHostFlag,
	SlaveDbEnableFlag,
	MasterDbHostFlag,
	MasterDbPortFlag,
	MasterDbUserFlag,
	MasterDbPasswordFlag,
	MasterDbNameFlag,
	StartDataStoreIdFlag,
}

var optionalFlags = []cli.Flag{
	L1PollingIntervalFlag,
	L1HeaderBufferSizeFlag,
	L1ConfirmationDepthFlag,
	L1StartingHeightFlag,
	L2StartingHeightFlag,
	L1BedrockStartingHeightFlag,
	L2BedrockStartingHeightFlag,
	FraudProofWindowsFlags,
	AddressManagerFlag,
	SystemConfigProxyFlag,
	OptimismPortalProxyFlag,
	L2OutputOracleProxyFlag,
	L1CrossDomainMessengerProxyFlag,
	L1StandardBridgeProxyFlag,
	L1ERC721BridgeProxyFlag,
	DataLayrServiceManagerAddrFlag,
	LegacyCanonicalTransactionChainFlag,
	LegacyStateCommitmentChainFlag,
	L2PollingIntervalFlag,
	L2ConfirmationDepthFlag,
	L2HeaderBufferSizeFlag,
	RetrieverTimeoutFlag,
	RetrieverSocketFlag,
	GraphProviderFlag,
	DataStorePollingDurationFlag,
	SlaveDbHostFlag,
	SlaveDbPortFlag,
	SlaveDbUserFlag,
	SlaveDbPasswordFlag,
	SlaveDbNameFlag,
	ExporterAddressFlag,
	NetworkLabelFlag,
	VersionEnableFlag,
	UnhealthyTimePeriodFlag,
	SequencerPollingSecondsFlag,
	EnableK8sQueryFlag,
	EnableRollUpGasPricesFlag,
	EnableGasBaseFeeFlag,
	EnableApiCacheFlag,
	ApiCacheListSize,
	ApiCacheDetailSize,
	ApiCacheListExpireTime,
	ApiCacheDetailExpireTime,
	L1AccountCheckingAddressFlag,
	L2AccountCheckingAddressFlag,
	EnableWithdrawCalcFlag,
	ChainIdFlag,
	TransferBigValueAddressInEthereumFlag,
	TransferBigValueAddressInMantleFlag,
	TransferBigValueInEthereumFlag,
	TransferBigValueInMantleFlag,
	WithdrawBigValueAddressFlag,
	TokenListUrlFlag,
}

func init() {

	Flags = append(requiredFlags, optionalFlags...)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag
