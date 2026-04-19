package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"

	bclb "github.com/betting-platform/internal/compliance/bclb"
	compliance "github.com/betting-platform/internal/compliance/bclb"
)

// ComplianceHandler handles compliance-related HTTP requests
type ComplianceHandler struct {
	bclbService *bclb.BCLBService
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(bclbService *bclb.BCLBService) *ComplianceHandler {
	return &ComplianceHandler{
		bclbService: bclbService,
	}
}

// ValidateBetPlacement validates a bet placement request
func (h *ComplianceHandler) ValidateBetPlacement(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID     string          `json:"user_id"`
		BetAmount  decimal.Decimal `json:"bet_amount"`
		BetType    string          `json:"bet_type"`
		Selections int             `json:"selections"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	check, err := h.bclbService.ValidateBetPlacement(r.Context(), req.UserID, req.BetAmount, req.BetType, req.Selections)
	if err != nil {
		WriteError(w, err, "Failed to validate bet placement", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, check, http.StatusOK)
}

// ValidateTransaction validates a financial transaction
func (h *ComplianceHandler) ValidateTransaction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID          string          `json:"user_id"`
		Amount          decimal.Decimal `json:"amount"`
		TransactionType string          `json:"transaction_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	check, err := h.bclbService.ValidateTransaction(r.Context(), req.UserID, req.Amount, req.TransactionType)
	if err != nil {
		WriteError(w, err, "Failed to validate transaction", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, check, http.StatusOK)
}

// GetComplianceReport generates a compliance report
func (h *ComplianceHandler) GetComplianceReport(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly" // default period
	}

	report, err := h.bclbService.GenerateComplianceReport(r.Context(), period)
	if err != nil {
		WriteError(w, err, "Failed to generate compliance report", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, report, http.StatusOK)
}

// GetComplianceStatus returns the current compliance status
func (h *ComplianceHandler) GetComplianceStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.bclbService.GetComplianceStatus(r.Context())
	if err != nil {
		WriteError(w, err, "Failed to get compliance status", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, status, http.StatusOK)
}

// SetUserSelfExclusion sets a user's self-exclusion status
func (h *ComplianceHandler) SetUserSelfExclusion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		Duration string `json:"duration"` // e.g., "6m", "1y", "permanent"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	duration, err := parseDuration(req.Duration)
	if err != nil {
		WriteError(w, err, "Invalid duration format", http.StatusBadRequest)
		return
	}

	err = h.bclbService.SetUserSelfExclusion(r.Context(), req.UserID, duration)
	if err != nil {
		WriteError(w, err, "Failed to set self-exclusion", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "Self-exclusion set successfully"}, http.StatusOK)
}

// RemoveUserSelfExclusion removes a user's self-exclusion status
func (h *ComplianceHandler) RemoveUserSelfExclusion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.bclbService.RemoveUserSelfExclusion(r.Context(), req.UserID)
	if err != nil {
		WriteError(w, err, "Failed to remove self-exclusion", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "Self-exclusion removed successfully"}, http.StatusOK)
}

// UpdateUserLimits updates a user's betting limits
func (h *ComplianceHandler) UpdateUserLimits(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID       string          `json:"user_id"`
		DailyLimit   decimal.Decimal `json:"daily_limit"`
		WeeklyLimit  decimal.Decimal `json:"weekly_limit"`
		MonthlyLimit decimal.Decimal `json:"monthly_limit"`
		MaxBetSize   decimal.Decimal `json:"max_bet_size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	limits := compliance.UserLimits{
		UserID:       req.UserID,
		DailyLimit:   req.DailyLimit,
		WeeklyLimit:  req.WeeklyLimit,
		MonthlyLimit: req.MonthlyLimit,
		MaxBetSize:   req.MaxBetSize,
	}

	err := h.bclbService.UpdateUserLimits(r.Context(), req.UserID, limits)
	if err != nil {
		WriteError(w, err, "Failed to update user limits", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]string{"message": "User limits updated successfully"}, http.StatusOK)
}

// GetComplianceViolations returns recent compliance violations
func (h *ComplianceHandler) GetComplianceViolations(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	severity := r.URL.Query().Get("severity")
	_ = r.URL.Query().Get("user_id") // userID parameter for future use

	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// In a real implementation, this would query the database for violations
	// For now, return sample data
	violations := []compliance.ComplianceViolation{
		{
			Type:        "AGE_VERIFICATION",
			Description: "Underage betting attempt",
			Severity:    "CRITICAL",
			Action:      "BLOCK_BET",
		},
		{
			Type:        "DAILY_LIMIT",
			Description: "Daily stake limit exceeded",
			Severity:    "MEDIUM",
			Amount:      decimal.NewFromInt(25000),
			Action:      "BLOCK_BET",
		},
	}

	// Filter violations based on parameters
	var filteredViolations []compliance.ComplianceViolation
	for _, violation := range violations {
		if severity != "" && string(violation.Severity) != severity {
			continue
		}
		filteredViolations = append(filteredViolations, violation)
		if len(filteredViolations) >= limit {
			break
		}
	}

	WriteJSON(w, filteredViolations, http.StatusOK)
}

// GetAMLReport generates an AML (Anti-Money Laundering) report
func (h *ComplianceHandler) GetAMLReport(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly" // default period
	}

	// In a real implementation, this would generate a comprehensive AML report
	// For now, return sample data
	amlReport := map[string]interface{}{
		"period":               period,
		"total_transactions":   10000,
		"flagged_transactions": 25,
		"suspicious_patterns":  8,
		"large_transactions":   100,
		"high_risk_users":      15,
		"recommendations": []string{
			"Enhanced monitoring for high-risk users",
			"Implement real-time transaction screening",
			"Improve SAR (Suspicious Activity Report) filing process",
		},
		"generated_at": time.Now(),
	}

	WriteJSON(w, amlReport, http.StatusOK)
}

// GetResponsibleGamingReport generates a responsible gaming report
func (h *ComplianceHandler) GetResponsibleGamingReport(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly" // default period
	}

	// In a real implementation, this would generate a comprehensive responsible gaming report
	// For now, return sample data
	rgReport := map[string]interface{}{
		"period":                period,
		"total_users":           10000,
		"self_excluded_users":   150,
		"limit_breached_users":  200,
		"cooling_off_users":     75,
		"intervention_contacts": 50,
		"responsible_gaming_tools_usage": map[string]interface{}{
			"deposit_limits": 1200,
			"bet_limits":     800,
			"time_limits":    300,
			"self_exclusion": 150,
		},
		"recommendations": []string{
			"Increase awareness of responsible gaming tools",
			"Implement proactive intervention for at-risk users",
			"Enhance cooling-off period enforcement",
		},
		"generated_at": time.Now(),
	}

	WriteJSON(w, rgReport, http.StatusOK)
}

// Helper functions

// parseDuration parses a duration string (e.g., "6m", "1y", "permanent")
func parseDuration(duration string) (time.Duration, error) {
	switch duration {
	case "permanent":
		return time.Duration(1<<63 - 1), nil // Maximum duration
	case "6m":
		return 6 * 30 * 24 * time.Hour, nil // 6 months
	case "1y":
		return 365 * 24 * time.Hour, nil // 1 year
	case "2y":
		return 2 * 365 * 24 * time.Hour, nil // 2 years
	case "5y":
		return 5 * 365 * 24 * time.Hour, nil // 5 years
	default:
		return 0, fmt.Errorf("invalid duration format: %s", duration)
	}
}

// RegisterComplianceRoutes registers compliance routes
func RegisterComplianceRoutes(mux *http.ServeMux, complianceHandler *ComplianceHandler) {
	// Validation endpoints
	mux.HandleFunc("/compliance/validate/bet", complianceHandler.ValidateBetPlacement)
	mux.HandleFunc("/compliance/validate/transaction", complianceHandler.ValidateTransaction)

	// Reporting endpoints
	mux.HandleFunc("/compliance/report", complianceHandler.GetComplianceReport)
	mux.HandleFunc("/compliance/status", complianceHandler.GetComplianceStatus)
	mux.HandleFunc("/compliance/violations", complianceHandler.GetComplianceViolations)

	// AML endpoints
	mux.HandleFunc("/compliance/aml/report", complianceHandler.GetAMLReport)

	// Responsible gaming endpoints
	mux.HandleFunc("/compliance/responsible-gaming/report", complianceHandler.GetResponsibleGamingReport)

	// User management endpoints
	mux.HandleFunc("/compliance/users/self-exclusion", complianceHandler.SetUserSelfExclusion)
	mux.HandleFunc("/compliance/users/self-exclusion", complianceHandler.RemoveUserSelfExclusion)
	mux.HandleFunc("/compliance/users/limits", complianceHandler.UpdateUserLimits)
}
