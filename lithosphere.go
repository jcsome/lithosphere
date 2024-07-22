package lithosphere

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ethereum/go-ethereum/log"

	"github.com/mantlenetworkio/lithosphere/api/common/httputil"
	"github.com/mantlenetworkio/lithosphere/business"
	"github.com/mantlenetworkio/lithosphere/business/mantle_da"
	"github.com/mantlenetworkio/lithosphere/config"
	"github.com/mantlenetworkio/lithosphere/database"
	"github.com/mantlenetworkio/lithosphere/event/processors"
	"github.com/mantlenetworkio/lithosphere/event/processors/bridge"
	"github.com/mantlenetworkio/lithosphere/event/processors/bridge/ovm1"
	"github.com/mantlenetworkio/lithosphere/event/processors/bridge/ovm1/crossdomain"
	metrics2 "github.com/mantlenetworkio/lithosphere/metrics"
	"github.com/mantlenetworkio/lithosphere/synchronizer"
	"github.com/mantlenetworkio/lithosphere/synchronizer/node"
)

// Lithosphere contains the necessary resources for
// indexing the configured L1 and L2 chains
type Lithosphere struct {
	log      log.Logger
	DB       *database.DB
	l1Client node.EthClient
	l2Client node.EthClient

	apiServer *httputil.HTTPServer

	metricsServer *httputil.HTTPServer

	metricsRegistry *prometheus.Registry

	L1Sync *synchronizer.L1Sync
	L2Sync *synchronizer.L2Sync

	BridgeProcessor   *processors.EventProcessor
	BusinessProcessor *business.BusinessProcessor

	shutdown context.CancelCauseFunc

	stopped atomic.Bool
}

// NewLithosphere initializes an instance of the Lithosphere
func NewLithosphere(ctx context.Context, log log.Logger, cfg *config.Config, shutdown context.CancelCauseFunc) (*Lithosphere, error) {

	out := &Lithosphere{
		log:             log,
		metricsRegistry: metrics2.NewRegistry(),
		shutdown:        shutdown,
	}
	if err := out.initFromConfig(ctx, cfg); err != nil {
		return nil, errors.Join(err, out.Stop(ctx))
	}
	return out, nil
}

func (i *Lithosphere) Start(ctx context.Context) error {
	// If any of these services has a critical failure,
	// the service can request a shutdown, while providing the error cause.
	if err := i.L1Sync.Start(); err != nil {
		return fmt.Errorf("failed to start L1 Sync: %w", err)
	}
	if err := i.L2Sync.Start(); err != nil {
		return fmt.Errorf("failed to start L2 Sync: %w", err)
	}
	if err := i.BridgeProcessor.Start(); err != nil {
		return fmt.Errorf("failed to start bridge processor: %w", err)
	}
	if err := i.BusinessProcessor.Start(); err != nil {
		return fmt.Errorf("failed to start Business processor: %w", err)
	}
	return nil
}

func (i *Lithosphere) Stop(ctx context.Context) error {
	var result error

	if i.L1Sync != nil {
		if err := i.L1Sync.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close L1 Sync: %w", err))
		}
	}

	if i.L2Sync != nil {
		if err := i.L2Sync.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close L2 Sync: %w", err))
		}
	}

	if i.BridgeProcessor != nil {
		if err := i.BridgeProcessor.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close bridge processor: %w", err))
		}
	}

	if i.BusinessProcessor != nil {
		if err := i.BusinessProcessor.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close business processor: %w", err))
		}
	}

	// Now that the ETLs are closed, we can stop the RPC clients
	if i.l1Client != nil {
		i.l1Client.Close()
	}
	if i.l2Client != nil {
		i.l2Client.Close()
	}

	if i.apiServer != nil {
		if err := i.apiServer.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close lithosphere API server: %w", err))
		}
	}

	// DB connection can be closed last, after all its potential users have shut down
	if i.DB != nil {
		if err := i.DB.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close DB: %w", err))
		}
	}

	if i.metricsServer != nil {
		if err := i.metricsServer.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close metrics server: %w", err))
		}
	}

	i.stopped.Store(true)

	i.log.Info("lithosphere stopped")

	return result
}

func (i *Lithosphere) Stopped() bool {
	return i.stopped.Load()
}

func (i *Lithosphere) initFromConfig(ctx context.Context, cfg *config.Config) error {
	if err := i.initRPCClients(ctx, cfg.RPCs); err != nil {
		return fmt.Errorf("failed to start RPC clients: %w", err)
	}
	if err := i.initDB(ctx, cfg.MasterDB); err != nil {
		return fmt.Errorf("failed to init DB: %w", err)
	}
	if err := i.initL1Syncer(*cfg); err != nil {
		return fmt.Errorf("failed to init L1 Sync: %w", err)
	}
	if err := i.initL2ETL(*cfg); err != nil {
		return fmt.Errorf("failed to init L2 Sync: %w", err)
	}
	if err := i.initBridgeProcessor(cfg.Chain); err != nil {
		return fmt.Errorf("failed to init Bridge Processor: %w", err)
	}
	if err := i.initBusinessProcessor(*cfg); err != nil {
		return fmt.Errorf("failed to init Business Processor: %w", err)
	}
	if err := i.startHttpServer(ctx, cfg.HTTPServer); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	if err := i.startMetricsServer(ctx, cfg.MetricsServer); err != nil {
		return fmt.Errorf("failed to start Metrics server: %w", err)
	}
	if cfg.WithdrawCalcEnable {
		if err := i.initV1Withdraw(ctx, cfg); err != nil {
			return fmt.Errorf("fialed to init v1 withdraw: %w", err)
		}
	}
	return nil
}

func (i *Lithosphere) initRPCClients(ctx context.Context, rpcsConfig config.RPCsConfig) error {
	l1EthClient, err := node.DialEthClient(ctx, rpcsConfig.L1RPC, metrics2.NewNodeMetrics(i.metricsRegistry, "l1"))
	if err != nil {
		return fmt.Errorf("failed to dial L1 client: %w", err)
	}
	i.l1Client = l1EthClient

	l2EthClient, err := node.DialEthClient(ctx, rpcsConfig.L2RPC, metrics2.NewNodeMetrics(i.metricsRegistry, "l2"))
	if err != nil {
		return fmt.Errorf("failed to dial L2 client: %w", err)
	}
	i.l2Client = l2EthClient
	return nil
}

func (i *Lithosphere) initDB(ctx context.Context, cfg config.DBConfig) error {
	db, err := database.NewDB(ctx, i.log, cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	i.DB = db
	return nil
}

func (i *Lithosphere) initL1Syncer(cfg config.Config) error {
	l1Cfg := synchronizer.Config{
		LoopIntervalMsec:  cfg.Chain.L1PollingInterval,
		HeaderBufferSize:  cfg.Chain.L1HeaderBufferSize,
		ConfirmationDepth: big.NewInt(int64(cfg.Chain.L1ConfirmationDepth)),
		StartHeight:       big.NewInt(int64(cfg.Chain.L1StartingHeight)),
	}
	l1Sync, err := synchronizer.NewL1Sync(l1Cfg, i.log, i.DB, metrics2.NewMetrics(i.metricsRegistry, "l1"),
		i.l1Client, cfg.Chain.L1Contracts, i.shutdown, cfg.ExporterConfig.TransferBigValueAddressInEthereum)
	if err != nil {
		return err
	}
	i.L1Sync = l1Sync
	return nil
}

func (i *Lithosphere) initL2ETL(cfg config.Config) error {
	// L2 (defaults to predeploy contracts)
	l2Cfg := synchronizer.Config{
		LoopIntervalMsec:  cfg.Chain.L2PollingInterval,
		HeaderBufferSize:  cfg.Chain.L2HeaderBufferSize,
		ConfirmationDepth: big.NewInt(int64(cfg.Chain.L2ConfirmationDepth)),
		StartHeight:       big.NewInt(int64(cfg.Chain.L2StartingHeight)),
	}
	l2Sync, err := synchronizer.NewL2Sync(l2Cfg, i.log, i.DB, metrics2.NewMetrics(i.metricsRegistry, "l2"),
		i.l2Client, cfg.Chain.L2Contracts, i.shutdown, cfg.ExporterConfig.TransferBigValueAddressInMantle)
	if err != nil {
		return err
	}
	i.L2Sync = l2Sync
	return nil
}

func (i *Lithosphere) initBridgeProcessor(chainConfig config.ChainConfig) error {
	bridgeProcessor, err := processors.NewBridgeProcessor(
		i.log, i.DB, bridge.NewMetrics(i.metricsRegistry), i.L1Sync, i.L2Sync, chainConfig, i.shutdown)
	if err != nil {
		return err
	}
	i.BridgeProcessor = bridgeProcessor
	return nil
}

func (i *Lithosphere) initBusinessProcessor(cfg config.Config) error {
	mantleDACfg, err := mantle_da.NewMantleDataStoreConfig(cfg.DA)
	mantleDA, err := mantle_da.NewMantleDataStore(&mantleDACfg)
	if err != nil {
		return err
	}
	businessProcessor := business.NewBusinessProcessor(
		i.log, i.DB, i.l1Client, i.l2Client, mantleDA, cfg, i.shutdown)

	i.BusinessProcessor = businessProcessor
	return nil
}

func (i *Lithosphere) startHttpServer(ctx context.Context, cfg config.ServerConfig) error {
	i.log.Debug("starting http server...", "port", cfg.Port)

	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/healthz"))

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	srv, err := httputil.StartHTTPServer(addr, r)
	if err != nil {
		return fmt.Errorf("http server failed to start: %w", err)
	}
	i.apiServer = srv
	i.log.Info("http server started", "addr", srv.Addr())
	return nil
}

func (i *Lithosphere) startMetricsServer(ctx context.Context, cfg config.ServerConfig) error {
	i.log.Debug("starting metrics server...", "port", cfg.Port)
	srv, err := metrics2.StartServer(i.metricsRegistry, cfg.Host, cfg.Port)
	if err != nil {
		return fmt.Errorf("metrics server failed to start: %w", err)
	}
	i.metricsServer = srv
	i.log.Info("metrics server started", "addr", srv.Addr())
	return nil
}

func (i *Lithosphere) initV1Withdraw(ctx context.Context, cfg *config.Config) error {
	l2SendMessageList := i.DB.L2SentMessageEvent.L2SentMessageList()
	var legacyWithdrawal crossdomain.LegacyWithdrawal

	for _, l2SendMessage := range l2SendMessageList {
		data, _ := hex.DecodeString(l2SendMessage.Message[2:])
		legacyWithdrawal.MessageSender = l2SendMessage.Sender
		legacyWithdrawal.XDomainData = data
		legacyWithdrawal.XDomainNonce = l2SendMessage.MessageNonce
		legacyWithdrawal.XDomainSender = l2SendMessage.Sender
		legacyWithdrawal.XDomainTarget = l2SendMessage.Target

		hash, err := ovm1.CalcTransaction(&legacyWithdrawal, &cfg.Chain.L1Contracts.L1CrossDomainMessengerProxy, new(big.Int).SetUint64(uint64(cfg.Chain.ChainID)))
		if err != nil {
			i.log.Error("calc withdrawal hash fail", "err", err)
		}

		err = i.DB.L2ToL1.UpdateV1L2Tol1WithdrawalHash(l2SendMessage.TxHash, hash)
		if err != nil {
			i.log.Error("update withdrawal hash fail", "err", err)
		}
	}

	i.log.Info("calc all withdrawal hash success")
	return nil
}
