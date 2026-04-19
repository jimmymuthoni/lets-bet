package http

import (
	"net/http"

	"github.com/betting-platform/internal/core/usecase"
	"github.com/betting-platform/internal/infrastructure/validation"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// FairnessHandler exposes the /verify and /commitment endpoints players use
// to independently verify a round's crash point.
type FairnessHandler struct {
	fair *usecase.ProvablyFairService
}

func NewFairnessHandler(fair *usecase.ProvablyFairService) *FairnessHandler {
	return &FairnessHandler{fair: fair}
}

// RegisterRoutes wires the endpoints under /api/fairness.
func (h *FairnessHandler) RegisterRoutes(r *mux.Router) {
	s := r.PathPrefix("/api/fairness").Subrouter()
	s.HandleFunc("/commitment", h.commitment).Methods(http.MethodPost)
	s.HandleFunc("/verify", h.verify).Methods(http.MethodPost)
}

type commitmentRequest struct {
	ServerSeed string `json:"server_seed"`
}

type commitmentResponse struct {
	ServerSeedHash string `json:"server_seed_hash"`
}

type verifyRequest struct {
	ServerSeed   string          `json:"server_seed"`
	ClientSeed   string          `json:"client_seed"`
	RoundNumber  int64           `json:"round_number"`
	ClaimedCrash decimal.Decimal `json:"claimed_crash"`
}

type verifyResponse struct {
	Match      bool            `json:"match"`
	Calculated decimal.Decimal `json:"calculated"`
	Claimed    decimal.Decimal `json:"claimed"`
}

func (h *FairnessHandler) commitment(w http.ResponseWriter, r *http.Request) {
	var req commitmentRequest
	if err := validation.DecodeJSON(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if req.ServerSeed == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "server_seed is required"})
		return
	}
	writeJSON(w, http.StatusOK, commitmentResponse{ServerSeedHash: h.fair.HashServerSeed(req.ServerSeed)})
}

func (h *FairnessHandler) verify(w http.ResponseWriter, r *http.Request) {
	var req verifyRequest
	if err := validation.DecodeJSON(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if req.ServerSeed == "" || req.ClientSeed == "" || req.RoundNumber <= 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "server_seed, client_seed, round_number required"})
		return
	}
	calc := h.fair.CalculateCrashPoint(req.ServerSeed, req.ClientSeed, req.RoundNumber)
	writeJSON(w, http.StatusOK, verifyResponse{
		Match:      calc.Equal(req.ClaimedCrash),
		Calculated: calc,
		Claimed:    req.ClaimedCrash,
	})
}
