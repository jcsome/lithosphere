package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/mantlenetworkio/lithosphere/cache"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ethereum/go-ethereum/log"

	"github.com/mantlenetworkio/lithosphere/api/common/httputil"
	"github.com/mantlenetworkio/lithosphere/api/routes"
	"github.com/mantlenetworkio/lithosphere/api/service"
	"github.com/mantlenetworkio/lithosphere/config"
	"github.com/mantlenetworkio/lithosphere/database"
	metrics2 "github.com/mantlenetworkio/lithosphere/metrics"
)

const ethereumAddressRegex = `^0x[a-fA-F0-9]{40}$`

const (
	MetricsNamespace = "lithosphere_api"
	idParam          = "{id}"
	indexParam       = "{index}"

	HealthPath           = "/healthz"
	MetricsPath          = "/api/metrics"
	DepositsV1Path       = "/api/v1/deposits"
	WithdrawalsV1Path    = "/api/v1/withdrawals"
	DataStoreListPath    = "/api/v1/datastore/list"
	DataStoreByIDPath    = "/api/v1/datastore/id/"
	DataStoreTxByIDPath  = "/api/v1/datastore/transaction/id/"
	StateRootListPath    = "/api/v1/stateroot/list"
	StateRootByIndexPath = "/api/v1/stateroot/index/"
)

type APIConfig struct {
	HTTPServer    config.ServerConfig
	MetricsServer config.ServerConfig
}

type API struct {
	log             log.Logger
	router          *chi.Mux
	metricsRegistry *prometheus.Registry
	apiServer       *httputil.HTTPServer
	metricsServer   *httputil.HTTPServer
	db              *database.DB
	stopped         atomic.Bool
}

func chiMetricsMiddleware(rec metrics2.HTTPRecorder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return metrics2.NewHTTPRecordingMiddleware(rec, next)
	}
}

func NewApi(ctx context.Context, log log.Logger, cfg *config.Config) (*API, error) {
	out := &API{log: log, metricsRegistry: metrics2.NewRegistry()}
	if err := out.initFromConfig(ctx, cfg, log); err != nil {
		return nil, errors.Join(err, out.Stop(ctx))
	}
	return out, nil
}

func (a *API) initFromConfig(ctx context.Context, cfg *config.Config, log log.Logger) error {
	if err := a.initDB(ctx, cfg, log); err != nil {
		return fmt.Errorf("failed to init DB: %w", err)
	}
	if err := a.startMetricsServer(cfg.MetricsServer); err != nil {
		return fmt.Errorf("failed to start metrics server: %w", err)
	}
	a.initRouter(cfg.HTTPServer, cfg)
	if err := a.startServer(cfg.HTTPServer); err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}
	return nil
}

func (a *API) initRouter(conf config.ServerConfig, cfg *config.Config) {
	v := new(service.Validator)

	var lruCache = new(cache.LruCache)
	if cfg.ApiCacheEnable {
		lruCache = cache.NewLruCache(cfg.CacheConfig)
	}

	svc := service.New(v, a.db.DataStore, a.db.L1ToL2, a.db.L2ToL1, a.db.Blocks, a.db.StateRoots, a.log)
	apiRouter := chi.NewRouter()
	h := routes.NewRoutes(a.log, apiRouter, svc, cfg.ApiCacheEnable, lruCache)

	mr := metrics2.NewRegistry()
	promRecorder := metrics2.NewPromHTTPRecorder(mr, MetricsNamespace)

	apiRouter.Use(chiMetricsMiddleware(promRecorder))
	apiRouter.Use(middleware.Timeout(time.Second * 12))
	apiRouter.Use(middleware.Recoverer)

	apiRouter.Use(middleware.Heartbeat(HealthPath))

	apiRouter.Get(fmt.Sprintf(DataStoreListPath), h.DataStoreListHandler)
	apiRouter.Get(fmt.Sprintf(DepositsV1Path), h.L1ToL2ListHandler)
	apiRouter.Get(fmt.Sprintf(WithdrawalsV1Path), h.L2ToL1ListHandler)
	apiRouter.Get(fmt.Sprintf(DataStoreByIDPath+idParam), h.DataStoreByIdHandler)
	apiRouter.Get(fmt.Sprintf(DataStoreTxByIDPath+idParam), h.DataStoreBlockByIDHandler)
	apiRouter.Get(fmt.Sprintf(StateRootListPath), h.StateRootListHandler)
	apiRouter.Get(fmt.Sprintf(StateRootByIndexPath+indexParam), h.StateRootByIndexHandler)

	a.router = apiRouter
}

func (a *API) initDB(ctx context.Context, cfg *config.Config, log log.Logger) error {
	var initDb *database.DB
	var err error
	if !cfg.SlaveDbEnable {
		initDb, err = database.NewDB(ctx, log, cfg.MasterDB)
		if err != nil {
			log.Error("failed to connect to master database", "err", err)
			return err
		}
	} else {
		initDb, err = database.NewDB(ctx, log, cfg.SlaveDB)
		if err != nil {
			log.Error("failed to connect to slave database", "err", err)
			return err
		}
	}
	a.db = initDb
	return nil
}

func (a *API) Start(ctx context.Context) error {
	return nil
}

func (a *API) Stop(ctx context.Context) error {
	var result error
	if a.apiServer != nil {
		if err := a.apiServer.Stop(ctx); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to stop API server: %w", err))
		}
	}
	if a.metricsServer != nil {
		if err := a.metricsServer.Stop(ctx); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to stop metrics server: %w", err))
		}
	}
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close DB: %w", err))
		}
	}
	a.stopped.Store(true)
	a.log.Info("API service shutdown complete")
	return result
}

func (a *API) startServer(serverConfig config.ServerConfig) error {
	a.log.Debug("API server listening...", "port", serverConfig.Port)
	addr := net.JoinHostPort(serverConfig.Host, strconv.Itoa(serverConfig.Port))
	srv, err := httputil.StartHTTPServer(addr, a.router)
	if err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}
	a.log.Info("API server started", "addr", srv.Addr().String())
	a.apiServer = srv
	return nil
}

func (a *API) startMetricsServer(metricsConfig config.ServerConfig) error {
	a.log.Debug("starting metrics server...", "port", metricsConfig.Port)
	srv, err := metrics2.StartServer(a.metricsRegistry, metricsConfig.Host, metricsConfig.Port)
	if err != nil {
		return fmt.Errorf("failed to start metrics server: %w", err)
	}
	a.log.Info("Metrics server started", "addr", srv.Addr().String())
	a.metricsServer = srv
	return nil
}

func (a *API) Stopped() bool {
	return a.stopped.Load()
}
