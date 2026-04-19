package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/betting-platform/internal/infrastructure/config"
	"github.com/betting-platform/internal/infrastructure/database"
	enginehttp "github.com/betting-platform/internal/infrastructure/http"
	"github.com/betting-platform/internal/infrastructure/http/health"
	"github.com/betting-platform/internal/infrastructure/http/middleware"
	"github.com/betting-platform/internal/infrastructure/logging"
	"github.com/betting-platform/internal/infrastructure/metrics"
	"github.com/betting-platform/internal/infrastructure/server"
	"github.com/gorilla/mux"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	logger := logging.Setup(cfg.Logging.Level, cfg.Logging.Format)
	logger.Info("starting engine", "env", cfg.Service.Environment)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	db, err := database.NewPostgresConnection(database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize MaxMind provider (optional)
	geoProvider, err := middleware.NewMaxMindProvider(middleware.DefaultDBPath())
	if err != nil {
		logger.Warn("failed to load maxmind database", "error", err)
	}

	r := mux.NewRouter()

	// Initialize metrics
	rec := metrics.New("engine")
	rec.RegisterRoutes(r)

	// Apply middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(middleware.CORS(cfg.Security))
	r.Use(rec.Middleware)
	r.Use(middleware.Geolocation(middleware.GeoConfig{
		Provider: geoProvider,
		Allowed:  cfg.Tenant.AllowedCountries,
	}))

	rl := middleware.NewRateLimiter(ctx, cfg.Security.RateLimitRequests, cfg.Security.RateLimitWindow)
	r.Use(rl.Middleware)

	// Health checks
	h := health.NewHandler("engine", "dev")
	h.Register(&health.PostgresChecker{DB: db})
	h.RegisterRoutes(r)

	// Engine handlers
	enginehttp.NewEngineHandler(nil).RegisterRoutes(r)

	go syncOddsFromProvider(ctx, logger)

	addr := fmt.Sprintf(":%d", cfg.Service.Port)
	if err := server.RunHTTP(ctx, addr, r, logger); err != nil {
		logger.Error("server terminated with error", "error", err)
		os.Exit(1)
	}
}

// syncOddsFromProvider polls the configured odds provider and refreshes the cache.
// It exits cleanly when ctx is cancelled.
func syncOddsFromProvider(ctx context.Context, logger *slog.Logger) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	logger.Info("odds sync started")
	for {
		select {
		case <-ctx.Done():
			logger.Info("odds sync stopped")
			return
		case <-ticker.C:
			// TODO: fetch from Sportradar, update Redis, publish via NATS.
			logger.Debug("odds sync tick")
		}
	}
}
