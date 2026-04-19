package http

import (
	"context"
	"net/http"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/core/usecase/wallet"
	"github.com/betting-platform/internal/infrastructure/events"
	"github.com/betting-platform/internal/infrastructure/logging"
	"github.com/betting-platform/internal/infrastructure/validation"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// DepositLookup resolves an M-Pesa CheckoutRequestID back to the initiating user.
// This is populated when we first call InitiateDeposit and persisted so the
// async callback can find its match.
type DepositLookup interface {
	FindByCheckoutID(ctx context.Context, checkoutID string) (uuid.UUID, decimal.Decimal, error)
	MarkCompleted(ctx context.Context, checkoutID, providerTxnID string) error
	MarkFailed(ctx context.Context, checkoutID, reason string) error
}

// MPesaHandler handles Safaricom Daraja async callbacks:
//
//	POST /api/mpesa/stk-callback   — STK Push (C2B deposit) confirmation
//	POST /api/mpesa/b2c-result     — B2C (withdrawal) result
//	POST /api/mpesa/b2c-timeout    — B2C queue timeout
type MPesaHandler struct {
	wallets  *wallet.Service
	deposits DepositLookup
	bus      events.Bus
}

// NewMPesaHandler constructs the callback handler. Pass nil for `bus` to skip
// event publishing.
func NewMPesaHandler(wallets *wallet.Service, deposits DepositLookup, bus events.Bus) *MPesaHandler {
	return &MPesaHandler{wallets: wallets, deposits: deposits, bus: bus}
}

func (h *MPesaHandler) RegisterRoutes(r *mux.Router) {
	s := r.PathPrefix("/api/mpesa").Subrouter()
	s.HandleFunc("/stk-callback", h.stkCallback).Methods(http.MethodPost)
	s.HandleFunc("/b2c-result", h.b2cResult).Methods(http.MethodPost)
	s.HandleFunc("/b2c-timeout", h.b2cTimeout).Methods(http.MethodPost)
}

// Daraja STK Push callback payload (trimmed to the fields we actually use).
type stkCallback struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  struct {
				Item []struct {
					Name  string `json:"Name"`
					Value any    `json:"Value"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

// stkAck is the response Safaricom requires — non-zero ResultCode aborts retries.
type stkAck struct {
	ResultCode int    `json:"ResultCode"`
	ResultDesc string `json:"ResultDesc"`
}

func (h *MPesaHandler) stkCallback(w http.ResponseWriter, r *http.Request) {
	logger := logging.FromContext(r.Context())

	var cb stkCallback
	if err := validation.DecodeJSON(w, r, &cb); err != nil {
		writeJSON(w, http.StatusBadRequest, stkAck{ResultCode: 1, ResultDesc: err.Error()})
		return
	}

	checkoutID := cb.Body.StkCallback.CheckoutRequestID
	if checkoutID == "" {
		writeJSON(w, http.StatusBadRequest, stkAck{ResultCode: 1, ResultDesc: "missing CheckoutRequestID"})
		return
	}

	if cb.Body.StkCallback.ResultCode != 0 {
		if err := h.deposits.MarkFailed(r.Context(), checkoutID, cb.Body.StkCallback.ResultDesc); err != nil {
			logger.Error("mpesa mark failed", "error", err, "checkout_id", checkoutID)
		}
		writeJSON(w, http.StatusOK, stkAck{ResultCode: 0, ResultDesc: "noted"})
		return
	}

	userID, amount, err := h.deposits.FindByCheckoutID(r.Context(), checkoutID)
	if err != nil {
		logger.Error("mpesa deposit lookup failed", "error", err, "checkout_id", checkoutID)
		// Ack anyway; retrying would just repeat the miss.
		writeJSON(w, http.StatusOK, stkAck{ResultCode: 0, ResultDesc: "ignored"})
		return
	}

	providerTxnID := metaString(cb.Body.StkCallback.CallbackMetadata.Item, "MpesaReceiptNumber")

	ref := uuid.New()
	if _, err := h.wallets.Credit(r.Context(), userID, amount, wallet.Movement{
		Type:          domain.TransactionTypeDeposit,
		ReferenceID:   &ref,
		ReferenceType: "DEPOSIT",
		ProviderName:  "MPESA",
		ProviderTxnID: providerTxnID,
		Description:   "M-Pesa STK deposit",
		CountryCode:   "KE",
	}); err != nil {
		logger.Error("mpesa credit failed", "error", err, "user_id", userID)
		writeJSON(w, http.StatusOK, stkAck{ResultCode: 1, ResultDesc: "credit failed"})
		return
	}

	if err := h.deposits.MarkCompleted(r.Context(), checkoutID, providerTxnID); err != nil {
		logger.Error("mpesa mark completed failed", "error", err, "checkout_id", checkoutID)
	}

	if h.bus != nil {
		_ = h.bus.Publish(r.Context(), events.SubjectDepositOK, map[string]any{
			"user_id":         userID,
			"amount":          amount,
			"provider":        "MPESA",
			"provider_txn_id": providerTxnID,
		})
	}

	writeJSON(w, http.StatusOK, stkAck{ResultCode: 0, ResultDesc: "accepted"})
}

func (h *MPesaHandler) b2cResult(w http.ResponseWriter, r *http.Request) {
	logger := logging.FromContext(r.Context())

	// We only log and ack for now — the actual debit happens when we initiate
	// the B2C. Reconciliation lives in a separate job.
	var body map[string]any
	_ = validation.DecodeJSON(w, r, &body)
	logger.Info("mpesa b2c result", "body", body)
	writeJSON(w, http.StatusOK, stkAck{ResultCode: 0, ResultDesc: "accepted"})
}

func (h *MPesaHandler) b2cTimeout(w http.ResponseWriter, r *http.Request) {
	logger := logging.FromContext(r.Context())
	var body map[string]any
	_ = validation.DecodeJSON(w, r, &body)
	logger.Warn("mpesa b2c timeout", "body", body)
	writeJSON(w, http.StatusOK, stkAck{ResultCode: 0, ResultDesc: "accepted"})
}

// metaString extracts a string value from the Daraja CallbackMetadata slice.
func metaString(items []struct {
	Name  string `json:"Name"`
	Value any    `json:"Value"`
}, name string) string {
	for _, it := range items {
		if it.Name == name {
			if s, ok := it.Value.(string); ok {
				return s
			}
		}
	}
	return ""
}
