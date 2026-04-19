package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/infrastructure/kyc"
	"github.com/betting-platform/internal/infrastructure/logging"
	"github.com/betting-platform/internal/infrastructure/validation"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// UserRepository is the minimal user repository contract this handler needs.
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateVerificationStatus(ctx context.Context, id uuid.UUID, isVerified bool) error
}

// KYCHandler exposes endpoints that kick off Smile ID verification flows.
type KYCHandler struct {
	provider kyc.Provider
	users    UserRepository
}

func NewKYCHandler(provider kyc.Provider, users UserRepository) *KYCHandler {
	return &KYCHandler{provider: provider, users: users}
}

// RegisterRoutes wires KYC endpoints under /api/kyc.
func (h *KYCHandler) RegisterRoutes(r *mux.Router) {
	s := r.PathPrefix("/api/kyc").Subrouter()
	s.HandleFunc("/verify-user", h.verifyUser).Methods(http.MethodPost)
	s.HandleFunc("/verify-id", h.verifyID).Methods(http.MethodPost)
}

type verifyUserRequest struct {
	UserID   string `json:"user_id"`
	IDType   string `json:"id_type"`
	IDNumber string `json:"id_number"`
}

type verifyIDRequest struct {
	CountryCode string `json:"country_code"`
	IDType      string `json:"id_type"`
	IDNumber    string `json:"id_number"`
}

func (h *KYCHandler) verifyUser(w http.ResponseWriter, r *http.Request) {
	logger := logging.FromContext(r.Context())
	var req verifyUserRequest
	if err := validation.DecodeJSON(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	errs := validation.Errors{}
	if msg, ok := validation.Required("user_id", req.UserID); !ok {
		errs.Add("user_id", msg)
	}
	if msg, ok := validation.Required("id_type", req.IDType); !ok {
		errs.Add("id_type", msg)
	}
	if msg, ok := validation.Required("id_number", req.IDNumber); !ok {
		errs.Add("id_number", msg)
	}
	if errs.HasAny() {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user_id"})
		return
	}
	user, err := h.users.GetByID(r.Context(), uid)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	result, err := h.provider.VerifyUser(r.Context(), user, req.IDType, req.IDNumber)
	if err != nil {
		logger.Error("kyc verify user failed", "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "kyc verification failed"})
		return
	}
	if result.Verified {
		if err := h.users.UpdateVerificationStatus(r.Context(), user.ID, true); err != nil {
			logger.Error("failed to persist verification status", "error", err)
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *KYCHandler) verifyID(w http.ResponseWriter, r *http.Request) {
	logger := logging.FromContext(r.Context())
	var req verifyIDRequest
	if err := validation.DecodeJSON(w, r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	errs := validation.Errors{}
	if msg, ok := validation.Required("country_code", req.CountryCode); !ok {
		errs.Add("country_code", msg)
	}
	if msg, ok := validation.Required("id_type", req.IDType); !ok {
		errs.Add("id_type", msg)
	}
	if msg, ok := validation.Required("id_number", req.IDNumber); !ok {
		errs.Add("id_number", msg)
	}
	if errs.HasAny() {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}
	result, err := h.provider.VerifyID(r.Context(), req.CountryCode, req.IDType, req.IDNumber)
	if err != nil {
		logger.Error("kyc verify id failed", "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "kyc verification failed"})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
