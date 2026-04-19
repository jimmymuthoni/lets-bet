package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

// Checker is a single dependency health check.
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

// Handler wires /healthz and /readyz endpoints.
type Handler struct {
	service  string
	version  string
	started  time.Time
	checkers []Checker
	mu       sync.RWMutex
}

func NewHandler(service, version string) *Handler {
	return &Handler{
		service: service,
		version: version,
		started: time.Now(),
	}
}

// Register adds a dependency checker.
func (h *Handler) Register(c Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, c)
}

// RegisterRoutes wires the health endpoints.
func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/healthz", h.liveness).Methods(http.MethodGet)
	r.HandleFunc("/readyz", h.readiness).Methods(http.MethodGet)
}

func (h *Handler) liveness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": h.service,
		"version": h.version,
		"uptime":  time.Since(h.started).String(),
	})
}

func (h *Handler) readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	h.mu.RLock()
	checkers := h.checkers
	h.mu.RUnlock()

	results := make(map[string]string, len(checkers))

	overall := http.StatusOK

	for _, c := range checkers {
		if err := c.Check(ctx); err != nil {
			results[c.Name()] = "fail: " + err.Error()
			overall = http.StatusServiceUnavailable
		} else {
			results[c.Name()] = "ok"
		}
	}

	status := "ready"
	if overall != http.StatusOK {
		status = "not_ready"
	}
	writeJSON(w, overall, map[string]any{
		"status":  status,
		"service": h.service,
		"version": h.version,
		"checks":  results,
	})
}

// PostgresChecker checks a PostgreSQL connection.
type PostgresChecker struct {
	DB *sql.DB
}

func (p *PostgresChecker) Name() string { return "postgres" }
func (p *PostgresChecker) Check(ctx context.Context) error {
	if p.DB == nil {
		return nil
	}
	return p.DB.PingContext(ctx)
}

// RedisChecker checks a Redis connection.
type RedisChecker struct {
	Client *redis.Client
}

func (r *RedisChecker) Name() string { return "redis" }
func (r *RedisChecker) Check(ctx context.Context) error {
	if r.Client == nil {
		return nil
	}
	return r.Client.Ping(ctx).Err()
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
