package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"

	audit "github.com/betting-platform/internal/security/audit"
	gdpr "github.com/betting-platform/internal/security/gdpr"
	pentest "github.com/betting-platform/internal/security/pentest"
	responsiblegaming "github.com/betting-platform/internal/security/responsiblegaming"
)

// SecurityHandler handles security-related HTTP requests
type SecurityHandler struct {
	auditService   *audit.SecurityAuditService
	pentestService *pentest.PenetrationTestService
	gdprService    *gdpr.GDPRService
	rgService      *responsiblegaming.ResponsibleGamingService
}

// NewSecurityHandler creates a new security handler
func NewSecurityHandler(
	auditService *audit.SecurityAuditService,
	pentestService *pentest.PenetrationTestService,
	gdprService *gdpr.GDPRService,
	rgService *responsiblegaming.ResponsibleGamingService,
) *SecurityHandler {
	return &SecurityHandler{
		auditService:   auditService,
		pentestService: pentestService,
		gdprService:    gdprService,
		rgService:      rgService,
	}
}

// PerformSecurityAudit performs a comprehensive security audit
func (h *SecurityHandler) PerformSecurityAudit(w http.ResponseWriter, r *http.Request) {
	audit, err := h.auditService.PerformSecurityAudit(r.Context())
	if err != nil {
		WriteError(w, err, "Failed to perform security audit", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, audit, http.StatusOK)
}

// GetSecurityMetrics returns security performance metrics
func (h *SecurityHandler) GetSecurityMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.auditService.GetSecurityMetrics(r.Context())
	if err != nil {
		WriteError(w, err, "Failed to get security metrics", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, metrics, http.StatusOK)
}

// PerformPenetrationTest performs a penetration test
func (h *SecurityHandler) PerformPenetrationTest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TestType string   `json:"test_type"`
		Scope    []string `json:"scope"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	test, err := h.pentestService.PerformPenetrationTest(r.Context(), pentest.TestType(req.TestType), req.Scope)
	if err != nil {
		WriteError(w, err, "Failed to perform penetration test", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, test, http.StatusOK)
}

// GetPenetrationTestHistory returns penetration test history
func (h *SecurityHandler) GetPenetrationTestHistory(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	tests, err := h.pentestService.GetPenetrationTestHistory(r.Context(), limit)
	if err != nil {
		WriteError(w, err, "Failed to get penetration test history", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, tests, http.StatusOK)
}

// PerformGDPRAssessment performs a GDPR compliance assessment
func (h *SecurityHandler) PerformGDPRAssessment(w http.ResponseWriter, r *http.Request) {
	assessment, err := h.gdprService.PerformGDPRAssessment(r.Context())
	if err != nil {
		WriteError(w, err, "Failed to perform GDPR assessment", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, assessment, http.StatusOK)
}

// ProcessDataSubjectRequest processes a data subject request
func (h *SecurityHandler) ProcessDataSubjectRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID      string `json:"user_id"`
		RequestType string `json:"request_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.gdprService.ProcessDataSubjectRequest(r.Context(), req.UserID, req.RequestType)
	if err != nil {
		WriteError(w, err, "Failed to process data subject request", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "Data subject request processed successfully"}, http.StatusOK)
}

// RecordDataBreach records a data breach
func (h *SecurityHandler) RecordDataBreach(w http.ResponseWriter, r *http.Request) {
	var req gdpr.DataBreach

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.gdprService.RecordDataBreach(r.Context(), req)
	if err != nil {
		WriteError(w, err, "Failed to record data breach", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "Data breach recorded successfully"}, http.StatusOK)
}

// PerformResponsibleGamingAssessment performs a responsible gaming assessment
func (h *SecurityHandler) PerformResponsibleGamingAssessment(w http.ResponseWriter, r *http.Request) {
	assessment, err := h.rgService.PerformResponsibleGamingAssessment(r.Context())
	if err != nil {
		WriteError(w, err, "Failed to perform responsible gaming assessment", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, assessment, http.StatusOK)
}

// SetSelfExclusion sets a user's self-exclusion status
func (h *SecurityHandler) SetSelfExclusion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		Duration string `json:"duration"`
		Reason   string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.rgService.SetSelfExclusion(r.Context(), req.UserID, req.Duration, req.Reason)
	if err != nil {
		WriteError(w, err, "Failed to set self-exclusion", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "Self-exclusion set successfully"}, http.StatusOK)
}

// SetDepositLimit sets a user's deposit limit
func (h *SecurityHandler) SetDepositLimit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string          `json:"user_id"`
		Type   string          `json:"type"`
		Amount decimal.Decimal `json:"amount"`
		Period string          `json:"period"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.rgService.SetDepositLimit(r.Context(), req.UserID, req.Type, req.Amount, req.Period)
	if err != nil {
		WriteError(w, err, "Failed to set deposit limit", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "Deposit limit set successfully"}, http.StatusOK)
}

// SetBettingLimit sets a user's betting limit
func (h *SecurityHandler) SetBettingLimit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string          `json:"user_id"`
		Type   string          `json:"type"`
		Amount decimal.Decimal `json:"amount"`
		Period string          `json:"period"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.rgService.SetBettingLimit(r.Context(), req.UserID, req.Type, req.Amount, req.Period)
	if err != nil {
		WriteError(w, err, "Failed to set betting limit", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "Betting limit set successfully"}, http.StatusOK)
}

// SetTimeLimit sets a user's time limit
func (h *SecurityHandler) SetTimeLimit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		Type     string `json:"type"`
		Duration string `json:"duration"`
		Period   string `json:"period"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse duration
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		WriteError(w, err, "Invalid duration format", http.StatusBadRequest)
		return
	}

	err = h.rgService.SetTimeLimit(r.Context(), req.UserID, req.Type, duration, req.Period)
	if err != nil {
		WriteError(w, err, "Failed to set time limit", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "Time limit set successfully"}, http.StatusOK)
}

// TriggerIntervention triggers an intervention for a user
func (h *SecurityHandler) TriggerIntervention(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID  string `json:"user_id"`
		Trigger string `json:"trigger"`
		Action  string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.rgService.TriggerIntervention(r.Context(), req.UserID, req.Trigger, req.Action)
	if err != nil {
		WriteError(w, err, "Failed to trigger intervention", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "Intervention triggered successfully"}, http.StatusOK)
}

// GetSecurityDashboard returns a comprehensive security dashboard
func (h *SecurityHandler) GetSecurityDashboard(w http.ResponseWriter, r *http.Request) {
	// Get security metrics
	metrics, err := h.auditService.GetSecurityMetrics(r.Context())
	if err != nil {
		WriteError(w, err, "Failed to get security metrics", http.StatusInternalServerError)
		return
	}

	// Get recent penetration tests
	tests, err := h.pentestService.GetPenetrationTestHistory(r.Context(), 5)
	if err != nil {
		WriteError(w, err, "Failed to get penetration test history", http.StatusInternalServerError)
		return
	}

	// Get latest GDPR assessment
	gdprAssessment, err := h.gdprService.PerformGDPRAssessment(r.Context())
	if err != nil {
		WriteError(w, err, "Failed to get GDPR assessment", http.StatusInternalServerError)
		return
	}

	// Get latest responsible gaming assessment
	rgAssessment, err := h.rgService.PerformResponsibleGamingAssessment(r.Context())
	if err != nil {
		WriteError(w, err, "Failed to get responsible gaming assessment", http.StatusInternalServerError)
		return
	}

	dashboard := map[string]interface{}{
		"security_metrics":         metrics,
		"recent_penetration_tests": tests,
		"gdpr_assessment":          gdprAssessment,
		"responsible_gaming":       rgAssessment,
		"generated_at":             time.Now(),
	}

	WriteJSON(w, dashboard, http.StatusOK)
}

// RegisterSecurityRoutes registers security routes
func RegisterSecurityRoutes(mux *http.ServeMux, securityHandler *SecurityHandler) {
	// Security audit endpoints
	mux.HandleFunc("/security/audit", securityHandler.PerformSecurityAudit)
	mux.HandleFunc("/security/metrics", securityHandler.GetSecurityMetrics)

	// Penetration testing endpoints
	mux.HandleFunc("/security/pentest", securityHandler.PerformPenetrationTest)
	mux.HandleFunc("/security/pentest/history", securityHandler.GetPenetrationTestHistory)

	// GDPR endpoints
	mux.HandleFunc("/security/gdpr/assessment", securityHandler.PerformGDPRAssessment)
	mux.HandleFunc("/security/gdpr/data-subject-request", securityHandler.ProcessDataSubjectRequest)
	mux.HandleFunc("/security/gdpr/data-breach", securityHandler.RecordDataBreach)

	// Responsible gaming endpoints
	mux.HandleFunc("/security/responsible-gaming/assessment", securityHandler.PerformResponsibleGamingAssessment)
	mux.HandleFunc("/security/responsible-gaming/self-exclusion", securityHandler.SetSelfExclusion)
	mux.HandleFunc("/security/responsible-gaming/deposit-limit", securityHandler.SetDepositLimit)
	mux.HandleFunc("/security/responsible-gaming/betting-limit", securityHandler.SetBettingLimit)
	mux.HandleFunc("/security/responsible-gaming/time-limit", securityHandler.SetTimeLimit)
	mux.HandleFunc("/security/responsible-gaming/intervention", securityHandler.TriggerIntervention)

	// Dashboard endpoint
	mux.HandleFunc("/security/dashboard", securityHandler.GetSecurityDashboard)
}
