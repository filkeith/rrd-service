package internal

import (
	"fmt"
	"log/slog"
	"os"

	"aerospike.com/rrd/internal/adaptors/storage"
	"aerospike.com/rrd/internal/config"
	"aerospike.com/rrd/internal/httpsrv"
	"aerospike.com/rrd/internal/httpsrv/handlers"
	"aerospike.com/rrd/internal/rrd"
)

const udfPath = "./udf/"

// App performs all services initializations.
type App struct {
	server *httpsrv.Server
	logger *slog.Logger
}

// NewApp returns new app instance.
func NewApp() (*App, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	lvl := slog.Level(0)
	if err = lvl.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return nil, fmt.Errorf("failed to parse loglevel: %w", err)
	}
	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: lvl},
		),
	)

	db, err := storage.NewStorage(
		cfg.StorageHost,
		cfg.StoragePort,
		cfg.StorageNamespace,
		cfg.StorageCapacity,
		udfPath,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	service := rrd.NewService(
		db,
		db,
	)

	router := handlers.NewRRD(
		service,
		service,
		logger,
	)

	httpServer := httpsrv.NewServer(
		cfg.HttpPort,
		router,
	)

	return &App{
		server: httpServer,
		logger: logger,
	}, nil
}

// Start starts http server.
func (app *App) Start() error {
	app.logger.Info("starting server...")
	return app.server.Start()
}
