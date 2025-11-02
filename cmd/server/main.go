package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brxyxn/go-logger"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bunzerolog"

	"github.com/brxyxn/engine-care-api/config"
	"github.com/brxyxn/engine-care-api/internal"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	opts := []logger.OptsFunc{
		func(o *logger.Opts) {
			o.Level = cfg.LoggerLevel
		},
	}

	log := logger.NewLogger(opts...)
	log.Info().Msg("🖨 logger initialized")

	ctx := context.Background()

	db := configDB(log, cfg)

	// we will refactor to plug in more routes later
	routes := internal.NewRoutes(ctx, cfg, log, db)
	r := routes.ConfigRoutes()

	run(r, log, cfg)
}

func configDB(log *logger.Logger, conf config.Config) *bun.DB {
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(conf.Dsn)))

	err := pgdb.Ping()
	if err != nil {
		log.Fatal().
			Str("stage", "pgdb").
			Err(err).
			Msg("failed to ping database")
	}

	// Create a Bun db on top of it.
	db := bun.NewDB(pgdb, pgdialect.New())

	hook := bunzerolog.NewQueryHook(
		bunzerolog.WithLogger(log.Logger),
		bunzerolog.WithQueryLogLevel(zerolog.DebugLevel),
		bunzerolog.WithSlowQueryLogLevel(zerolog.WarnLevel),
		bunzerolog.WithErrorQueryLogLevel(zerolog.ErrorLevel),
		bunzerolog.WithSlowQueryThreshold(3*time.Second),
	)
	db.AddQueryHook(hook)

	err = db.Ping()
	if err != nil {
		log.Fatal().
			Str("stage", "db").
			Err(err).
			Msg("failed to ping database")
	}

	log.Info().Msg("💾 database is running")

	return db
}

func run(r *mux.Router, log *logger.Logger, conf config.Config) {
	addr := fmt.Sprintf(":%s", conf.ServerPort)

	srvLogger := log.With().
		Str("port", conf.ServerPort).
		Logger()

	srvHandler := cors.New(cors.Options{
		AllowedOrigins:   conf.CorsAllowedOrigins,
		AllowedHeaders:   conf.CorsAllowedHeaders,
		AllowedMethods:   conf.CorsAllowedMethods,
		AllowCredentials: conf.CorsAllowCredentials,
		Debug:            conf.CorsDebug,
	}).Handler(r)

	// Configure and start the HTTP server.
	srv := &http.Server{
		Handler:      srvHandler,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
		srvLogger.Info().Msg("🚀 starting server")
		serverErr <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		srvLogger.Panic().Err(err).Msg("server crashed")
	case shutdownSignal := <-shutdown:
		srvLogger.Info().Interface("shutdown_command", shutdownSignal).Msg("starting shutdown")

		// Create a deadline to wait for.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(conf.ServerShutdownDeadline))
		defer cancel()

		// Attempt to gracefully shutdown API
		err := srv.Shutdown(ctx)
		if err != nil {
			srvLogger.Error().Err(err).Msg("graceful shutdown failed")
			err = srv.Close()
		}

		switch {
		case shutdownSignal == syscall.SIGSTOP:
			srvLogger.Panic().Msg("issue caused shutdown")
		case err != nil:
			srvLogger.Panic().Err(err).Msg("shutdown failed")
		}

		srvLogger.Info().Msg("shutting down")
	}
}
