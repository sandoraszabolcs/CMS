package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"github.com/szabolcs/cms/internal/infrastructure"
	"github.com/szabolcs/cms/internal/repository"
	"github.com/szabolcs/cms/internal/service"
	"github.com/szabolcs/cms/internal/simulator"
	transporthttp "github.com/szabolcs/cms/internal/transport/http"
	"github.com/szabolcs/cms/internal/transport/ws"
)

func main() {
	// Load config.
	cfg, err := infrastructure.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Set up structured logger.
	var logLevel slog.Level
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	// Root context cancelled on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Connect to Postgres.
	db, err := infrastructure.NewPostgresDB(ctx, cfg.DBURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}

	// Connect to Redis.
	rdb, err := infrastructure.NewRedisClient(ctx, cfg.RedisAddr)
	if err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	// Build repositories.
	passengerRepo := repository.NewPostgresPassengerRepo(db)
	validationRepo := repository.NewPostgresValidationRepo(db)
	vehicleRepo := repository.NewPostgresVehicleRepo(db)
	stopRepo := repository.NewPostgresStopRepo(db)
	odMatrixRepo := repository.NewPostgresODMatrixRepo(db)
	statsRepo := repository.NewPostgresStatsRepo(db)

	// Build services.
	validationSvc := service.NewValidationService(passengerRepo, validationRepo, stopRepo, rdb, logger)
	vehicleSvc := service.NewVehicleService(vehicleRepo)
	stopSvc := service.NewStopService(stopRepo)
	odMatrixSvc := service.NewODMatrixService(odMatrixRepo)
	statsSvc := service.NewStatsService(statsRepo)
	eventSvc := service.NewEventService(validationRepo)

	// Build HTTP handler.
	handler := transporthttp.NewHandler(validationSvc, vehicleSvc, stopSvc, odMatrixSvc, statsSvc, eventSvc, logger)

	// Build WebSocket hub.
	hub := ws.NewHub(rdb, validationRepo, logger)

	// Set up Gin.
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(
		transporthttp.RequestID(),
		transporthttp.AccessLog(logger),
		transporthttp.Recovery(logger),
	)

	handler.RegisterRoutes(router)
	router.GET("/ws", hub.HandleWS)

	// Serve static frontend.
	router.StaticFile("/", "./frontend/index.html")
	router.Static("/static", "./frontend")

	// Build simulator.
	sim := simulator.New(simulator.Deps{
		Validations: validationRepo,
		Vehicles:    vehicleRepo,
		Stops:       stopRepo,
		Redis:       rdb,
		Logger:      logger,
		Interval:    cfg.SimulatorInterval,
	})

	// HTTP server.
	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	// Run everything with errgroup.
	g, gCtx := errgroup.WithContext(ctx)

	// HTTP server.
	g.Go(func() error {
		logger.Info("starting HTTP server", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	// Graceful HTTP shutdown.
	g.Go(func() error {
		<-gCtx.Done()
		logger.Info("shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	})

	// WebSocket hub.
	g.Go(func() error {
		return hub.Run(gCtx)
	})

	// Simulator.
	g.Go(func() error {
		if err := sim.Start(gCtx); err != nil {
			return err
		}
		<-gCtx.Done()
		return sim.Stop()
	})

	// OD matrix refresh.
	g.Go(func() error {
		ticker := time.NewTicker(cfg.ODRefreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-gCtx.Done():
				return nil
			case <-ticker.C:
				if err := odMatrixRepo.Refresh(gCtx); err != nil {
					logger.Error("failed to refresh OD matrix", "error", err)
				} else {
					logger.Debug("OD matrix refreshed")
				}
			}
		}
	})

	// Wait for all goroutines.
	if err := g.Wait(); err != nil {
		logger.Error("server exited with error", "error", err)
	}

	// Cleanup.
	logger.Info("closing database connection")
	if err := db.Close(); err != nil {
		logger.Error("failed to close database", "error", err)
	}

	logger.Info("closing redis connection")
	if err := rdb.Close(); err != nil {
		logger.Error("failed to close redis", "error", err)
	}

	logger.Info("shutdown complete")
}
