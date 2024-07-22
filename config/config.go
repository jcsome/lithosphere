package config

import (
	"reflect"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/mantlenetworkio/lithosphere/event/op-bindings/predeploys"
	"github.com/mantlenetworkio/lithosphere/flag"
)

const (
	defaultLoopInterval     = 5000
	defaultHeaderBufferSize = 500
)

type Config struct {
	Migrations         string
	Chain              ChainConfig
	RPCs               RPCsConfig
	DA                 DAConfig
	MasterDB           DBConfig
	SlaveDB            DBConfig
	SlaveDbEnable      bool
	ApiCacheEnable     bool
	CacheConfig        CacheConfig
	HTTPServer         ServerConfig
	MetricsServer      ServerConfig
	ExporterConfig     ExporterConfig
	StartDataStoreId   uint32
	FraudProofWindows  uint64
	WithdrawCalcEnable bool
	CheckingAddress    CheckingConfig
	TokenListUrl       string
}

type L1Contracts struct {
	AddressManager                  common.Address
	SystemConfigProxy               common.Address
	OptimismPortalProxy             common.Address
	L2OutputOracleProxy             common.Address
	L1CrossDomainMessengerProxy     common.Address
	L1StandardBridgeProxy           common.Address
	L1ERC721BridgeProxy             common.Address
	DataLayrServiceManagerAddr      common.Address
	LegacyCanonicalTransactionChain common.Address
	LegacyStateCommitmentChain      common.Address
}

func (c L1Contracts) ForEach(cb func(string, common.Address) error) error {
	contracts := reflect.ValueOf(c)
	fields := reflect.VisibleFields(reflect.TypeOf(c))
	for _, field := range fields {
		addr := (contracts.FieldByName(field.Name).Interface()).(common.Address)
		if err := cb(field.Name, addr); err != nil {
			return err
		}
	}
	return nil
}

type L2Contracts struct {
	L2ToL1MessagePasser    common.Address
	L2CrossDomainMessenger common.Address
	L2StandardBridge       common.Address
	L2ERC721Bridge         common.Address
}

func L2ContractsFromPredeploys() L2Contracts {
	return L2Contracts{
		L2ToL1MessagePasser:    predeploys.L2ToL1MessagePasserAddr,
		L2CrossDomainMessenger: predeploys.L2CrossDomainMessengerAddr,
		L2StandardBridge:       predeploys.L2StandardBridgeAddr,
		L2ERC721Bridge:         predeploys.L2ERC721BridgeAddr,
	}
}

func (c L2Contracts) ForEach(cb func(string, common.Address) error) error {
	contracts := reflect.ValueOf(c)
	fields := reflect.VisibleFields(reflect.TypeOf(c))
	for _, field := range fields {
		addr := (contracts.FieldByName(field.Name).Interface()).(common.Address)
		if err := cb(field.Name, addr); err != nil {
			return err
		}
	}
	return nil
}

type ChainConfig struct {
	ChainID                 uint
	L1StartingHeight        uint
	L2StartingHeight        uint
	L1BedrockStartingHeight uint
	L2BedrockStartingHeight uint
	L1Contracts             L1Contracts
	L2Contracts             L2Contracts
	L1ConfirmationDepth     uint
	L2ConfirmationDepth     uint
	L1PollingInterval       uint
	L2PollingInterval       uint
	L1HeaderBufferSize      uint
	L2HeaderBufferSize      uint
}

type RPCsConfig struct {
	L1RPC string
	L2RPC string
}

type DBConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

type CacheConfig struct {
	ListSize         int
	DetailSize       int
	ListExpireTime   time.Duration
	DetailExpireTime time.Duration
}

type ServerConfig struct {
	Host string
	Port int
}

type DAConfig struct {
	RetrieverSocket          string
	RetrieverTimeout         time.Duration
	GraphProvider            string
	DataStorePollingDuration time.Duration
}

func LoadConfig(log log.Logger, cliCtx *cli.Context) (Config, error) {
	var cfg Config
	cfg = NewConfig(cliCtx)
	cfg.Chain.L2Contracts = L2ContractsFromPredeploys()

	if cfg.Chain.L1PollingInterval == 0 {
		cfg.Chain.L1PollingInterval = defaultLoopInterval
	}

	if cfg.Chain.L2PollingInterval == 0 {
		cfg.Chain.L2PollingInterval = defaultLoopInterval
	}

	if cfg.Chain.L1HeaderBufferSize == 0 {
		cfg.Chain.L1HeaderBufferSize = defaultHeaderBufferSize
	}

	if cfg.Chain.L2HeaderBufferSize == 0 {
		cfg.Chain.L2HeaderBufferSize = defaultHeaderBufferSize
	}

	log.Info("loaded chain config", "config", cfg.Chain)
	return cfg, nil
}

type ExporterConfig struct {
	ExportAddress                     string
	RpcProvider                       string
	NetworkLabel                      string
	Version                           bool
	UnhealthyTimePeriod               time.Duration
	SequencerPollingSeconds           time.Duration
	EnableK8sQuery                    bool
	EnableRollUpGasPrices             bool
	EnableGasBaseFee                  bool
	TransferBigValueAddressInEthereum string
	TransferBigValueInEthereum        string
	TransferBigValueAddressInMantle   string
	TransferBigValueInMantle          string
	WithdrawBigValueAddress           string
}

type CheckingConfig struct {
	L1AccountCheckingAddress string
	L2AccountCheckingAddress string
}

func NewConfig(ctx *cli.Context) Config {
	return Config{
		Migrations: ctx.String(flag.MigrationsFlag.Name),
		Chain: ChainConfig{
			ChainID:                 ctx.Uint(flag.ChainIdFlag.Name),
			L1StartingHeight:        ctx.Uint(flag.L1StartingHeightFlag.Name),
			L2StartingHeight:        ctx.Uint(flag.L2StartingHeightFlag.Name),
			L1BedrockStartingHeight: ctx.Uint(flag.L1BedrockStartingHeightFlag.Name),
			L2BedrockStartingHeight: ctx.Uint(flag.L2BedrockStartingHeightFlag.Name),
			L1Contracts: L1Contracts{
				AddressManager:                  common.HexToAddress(ctx.String(flag.AddressManagerFlag.Name)),
				SystemConfigProxy:               common.HexToAddress(ctx.String(flag.SystemConfigProxyFlag.Name)),
				OptimismPortalProxy:             common.HexToAddress(ctx.String(flag.OptimismPortalProxyFlag.Name)),
				L2OutputOracleProxy:             common.HexToAddress(ctx.String(flag.L2OutputOracleProxyFlag.Name)),
				L1CrossDomainMessengerProxy:     common.HexToAddress(ctx.String(flag.L1CrossDomainMessengerProxyFlag.Name)),
				L1StandardBridgeProxy:           common.HexToAddress(ctx.String(flag.L1StandardBridgeProxyFlag.Name)),
				L1ERC721BridgeProxy:             common.HexToAddress(ctx.String(flag.L1ERC721BridgeProxyFlag.Name)),
				DataLayrServiceManagerAddr:      common.HexToAddress(ctx.String(flag.DataLayrServiceManagerAddrFlag.Name)),
				LegacyCanonicalTransactionChain: common.HexToAddress(ctx.String(flag.LegacyCanonicalTransactionChainFlag.Name)),
				LegacyStateCommitmentChain:      common.HexToAddress(ctx.String(flag.LegacyStateCommitmentChainFlag.Name)),
			},
			L1ConfirmationDepth: ctx.Uint(flag.L1ConfirmationDepthFlag.Name),
			L2ConfirmationDepth: ctx.Uint(flag.L2ConfirmationDepthFlag.Name),
			L1PollingInterval:   ctx.Uint(flag.L1PollingIntervalFlag.Name),
			L2PollingInterval:   ctx.Uint(flag.L2PollingIntervalFlag.Name),
			L1HeaderBufferSize:  ctx.Uint(flag.L1HeaderBufferSizeFlag.Name),
			L2HeaderBufferSize:  ctx.Uint(flag.L2HeaderBufferSizeFlag.Name),
		},
		RPCs: RPCsConfig{
			L1RPC: ctx.String(flag.L1EthRpcFlag.Name),
			L2RPC: ctx.String(flag.L2EthRpcFlag.Name),
		},
		DA: DAConfig{
			RetrieverSocket:          ctx.String(flag.RetrieverSocketFlag.Name),
			RetrieverTimeout:         ctx.Duration(flag.RetrieverTimeoutFlag.Name),
			GraphProvider:            ctx.String(flag.GraphProviderFlag.Name),
			DataStorePollingDuration: ctx.Duration(flag.DataStorePollingDurationFlag.Name),
		},
		MasterDB: DBConfig{
			Host:     ctx.String(flag.MasterDbHostFlag.Name),
			Port:     ctx.Int(flag.MasterDbPortFlag.Name),
			Name:     ctx.String(flag.MasterDbNameFlag.Name),
			User:     ctx.String(flag.MasterDbUserFlag.Name),
			Password: ctx.String(flag.MasterDbPasswordFlag.Name),
		},
		SlaveDB: DBConfig{
			Host:     ctx.String(flag.SlaveDbHostFlag.Name),
			Port:     ctx.Int(flag.SlaveDbPortFlag.Name),
			Name:     ctx.String(flag.SlaveDbNameFlag.Name),
			User:     ctx.String(flag.SlaveDbUserFlag.Name),
			Password: ctx.String(flag.SlaveDbPasswordFlag.Name),
		},
		SlaveDbEnable:  ctx.Bool(flag.SlaveDbEnableFlag.Name),
		ApiCacheEnable: ctx.Bool(flag.EnableApiCacheFlag.Name),
		CacheConfig: CacheConfig{
			ListSize:         ctx.Int(flag.ApiCacheListSize.Name),
			DetailSize:       ctx.Int(flag.ApiCacheDetailSize.Name),
			ListExpireTime:   ctx.Duration(flag.ApiCacheListExpireTime.Name),
			DetailExpireTime: ctx.Duration(flag.ApiCacheDetailExpireTime.Name),
		},
		HTTPServer: ServerConfig{
			Host: ctx.String(flag.HttpHostFlag.Name),
			Port: ctx.Int(flag.HttpPortFlag.Name),
		},
		MetricsServer: ServerConfig{
			Host: ctx.String(flag.MetricsHostFlag.Name),
			Port: ctx.Int(flag.MetricsPortFlag.Name),
		},
		ExporterConfig: ExporterConfig{
			ExportAddress:                     ctx.String(flag.ExporterAddressFlag.Name),
			RpcProvider:                       ctx.String(flag.L2EthRpcFlag.Name),
			NetworkLabel:                      ctx.String(flag.NetworkLabelFlag.Name),
			Version:                           ctx.Bool(flag.VersionEnableFlag.Name),
			UnhealthyTimePeriod:               ctx.Duration(flag.UnhealthyTimePeriodFlag.Name),
			SequencerPollingSeconds:           ctx.Duration(flag.SequencerPollingSecondsFlag.Name),
			EnableK8sQuery:                    ctx.Bool(flag.EnableK8sQueryFlag.Name),
			EnableRollUpGasPrices:             ctx.Bool(flag.EnableRollUpGasPricesFlag.Name),
			EnableGasBaseFee:                  ctx.Bool(flag.EnableGasBaseFeeFlag.Name),
			TransferBigValueAddressInEthereum: ctx.String(flag.TransferBigValueAddressInEthereumFlag.Name),
			TransferBigValueAddressInMantle:   ctx.String(flag.TransferBigValueAddressInMantleFlag.Name),
			TransferBigValueInEthereum:        ctx.String(flag.TransferBigValueInEthereumFlag.Name),
			TransferBigValueInMantle:          ctx.String(flag.TransferBigValueInMantleFlag.Name),
			WithdrawBigValueAddress:           ctx.String(flag.WithdrawBigValueAddressFlag.Name),
		},
		CheckingAddress: CheckingConfig{
			L1AccountCheckingAddress: ctx.String(flag.L1AccountCheckingAddressFlag.Name),
			L2AccountCheckingAddress: ctx.String(flag.L2AccountCheckingAddressFlag.Name),
		},
		StartDataStoreId:   uint32(ctx.Uint64(flag.StartDataStoreIdFlag.Name)),
		FraudProofWindows:  ctx.Uint64(flag.FraudProofWindowsFlags.Name),
		WithdrawCalcEnable: ctx.Bool(flag.EnableWithdrawCalcFlag.Name),
		TokenListUrl:       ctx.String(flag.TokenListUrlFlag.Name),
	}
}
