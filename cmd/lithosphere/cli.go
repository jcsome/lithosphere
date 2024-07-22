package main

import (
	"context"
	"github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/params"

	"github.com/mantlenetworkio/lithosphere"
	"github.com/mantlenetworkio/lithosphere/api"
	"github.com/mantlenetworkio/lithosphere/common/cliapp"
	oplog "github.com/mantlenetworkio/lithosphere/common/log"
	"github.com/mantlenetworkio/lithosphere/common/opio"
	"github.com/mantlenetworkio/lithosphere/config"
	"github.com/mantlenetworkio/lithosphere/database"
	"github.com/mantlenetworkio/lithosphere/exporter"
	flag2 "github.com/mantlenetworkio/lithosphere/flag"
)

func runIndexer(ctx *cli.Context, shutdown context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	log := oplog.NewLogger(oplog.AppOut(ctx), oplog.ReadCLIConfig(ctx)).New("role", "lithosphere")
	oplog.SetGlobalLogHandler(log.GetHandler())
	log.Info("running indexer...")

	cfg, err := config.LoadConfig(log, ctx)
	if err != nil {
		log.Error("failed to load config", "err", err)
		return nil, err
	}

	return lithosphere.NewLithosphere(ctx.Context, log, &cfg, shutdown)
}

func runApi(ctx *cli.Context, _ context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	log := oplog.NewLogger(oplog.AppOut(ctx), oplog.ReadCLIConfig(ctx)).New("role", "api")
	oplog.SetGlobalLogHandler(log.GetHandler())
	log.Info("running api...")
	cfg, err := config.LoadConfig(log, ctx)
	if err != nil {
		log.Error("failed to load config", "err", err)
		return nil, err
	}
	return api.NewApi(ctx.Context, log, &cfg)
}

func runExporter(ctx *cli.Context, shutdown context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	log := oplog.NewLogger(oplog.AppOut(ctx), oplog.ReadCLIConfig(ctx)).New("role", "exporter")
	oplog.SetGlobalLogHandler(log.GetHandler())
	cfg, err := config.LoadConfig(log, ctx)
	if err != nil {
		log.Error("failed to load config", "err", err)
		return nil, err
	}
	db, err := database.NewDB(ctx.Context, log, cfg.MasterDB)
	if err != nil {
		log.Error("failed to connect to database", "err", err)
		return nil, err
	}
	log.Info("running exporter...")

	return exporter.NewExporter(cfg.ExporterConfig, db, shutdown)
}

func runMigrations(ctx *cli.Context) error {
	ctx.Context = opio.CancelOnInterrupt(ctx.Context)
	log := oplog.NewLogger(oplog.AppOut(ctx), oplog.ReadCLIConfig(ctx)).New("role", "migrations")
	oplog.SetGlobalLogHandler(log.GetHandler())
	log.Info("running migrations...")
	cfg, err := config.LoadConfig(log, ctx)
	if err != nil {
		log.Error("failed to load config", "err", err)
		return err
	}
	db, err := database.NewDB(ctx.Context, log, cfg.MasterDB)
	if err != nil {
		log.Error("failed to connect to database", "err", err)
		return err
	}
	defer db.Close()
	return db.ExecuteSQLMigration(cfg.Migrations)
}

func newCli(GitCommit string, GitDate string) *cli.App {
	flags := oplog.CLIFlags("LITHOSPHERE")
	flags = append(flags, flag2.Flags...)
	return &cli.App{
		Version:              params.VersionWithCommit(GitCommit, GitDate),
		Description:          "An indexer of all optimism events with a serving api layer",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:        "api",
				Flags:       flags,
				Description: "Runs the api service",
				Action:      cliapp.LifecycleCmd(runApi),
			},
			{
				Name:        "index",
				Flags:       flags,
				Description: "Runs the indexing service",
				Action:      cliapp.LifecycleCmd(runIndexer),
			},
			{
				Name:        "migrate",
				Flags:       flags,
				Description: "Runs the database migrations",
				Action:      runMigrations,
			},
			{
				Name:        "exporter",
				Flags:       flags,
				Description: "Runs the exporter service",
				Action:      cliapp.LifecycleCmd(runExporter),
			},
			{
				Name:        "version",
				Description: "print version",
				Action: func(ctx *cli.Context) error {
					cli.ShowVersion(ctx)
					return nil
				},
			},
		},
	}
}
