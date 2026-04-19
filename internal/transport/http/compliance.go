package http

import (
	"encoding/json"
	"net/http"
	"time"
)

// ValidateBetPlacement validates a bet placement request
func (h *ComplianceHandler) ValidateBetPlacement(w http.ResponseWriter, r *http.Request) {
	var req BetValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	check, err := h.bclbService.ValidateBetPlacement(r.Context(), req.UserID, req.BetAmount, req.BetType, req.Selections)
	if err != nil {
		WriteError(w, err, "Failed to validate bet placement", http.StatusInternalServerError)
		return
	}

	response := &BetValidationResponse{
		Success: true,
		Data:    check,
	}

	WriteJSON(w, response, http.StatusOK)
}

// GetUserComplianceStatus returns user compliance status
func (h *ComplianceHandler) GetUserComplianceStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user ID from URL path
	path := r.URL.Path[len("/api/compliance/user/status/"):]
	if path == "" {
		WriteError(w, nil, "User ID is required", http.StatusBadRequest)
		return
	}

	status, err := h.bclbService.GetUserComplianceStatus(ctx, path)
	if err != nil {
		WriteError(w, err, "Failed to get user compliance status", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    status,
	}, http.StatusOK)
}

// SetUserLimits sets user-specific limits
func (h *ComplianceHandler) SetUserLimits(w http.ResponseWriter, r *http.Request) {
	var req UserLimits
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract user ID from URL path
	path := r.URL.Path[len("/api/compliance/user/limits/"):]
	if path == "" {
		WriteError(w, nil, "User ID is required", http.StatusBadRequest)
		return
	}

	err := h.bclbService.SetUserLimits(r.Context(), path, &req)
	if err != nil {
		WriteError(w, err, "Failed to set user limits", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"message": "User limits set successfully",
	}, http.StatusOK)
}

// AddUserRestriction adds a restriction to a user
func (h *ComplianceHandler) AddUserRestriction(w http.ResponseWriter, r *http.Request) {
	var req UserRestriction
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract user ID from URL path
	path := r.URL.Path[len("/api/compliance/user/restrictions/"):]
	if path == "" {
		WriteError(w, nil, "User ID is required", http.StatusBadRequest)
		return
	}

	err := h.bclbService.AddUserRestriction(r.Context(), path, &req)
	if err != nil {
		WriteError(w, err, "Failed to add user restriction", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"message": "User restriction added successfully",
	}, http.StatusOK)
}

// GenerateComplianceReport generates a compliance report
func (h *ComplianceHandler) GenerateComplianceReport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type   string      `json:"type"`
		Period *TimePeriod `json:"period"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	report, err := h.bclbService.GenerateComplianceReport(r.Context(), req.Type)
	if err != nil {
		WriteError(w, err, "Failed to generate compliance report", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    report,
	}, http.StatusOK)
}

// GetComplianceMetrics returns compliance metrics
func (h *ComplianceHandler) GetComplianceMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	fromDateStr := r.URL.Query().Get("from_date")
	toDateStr := r.URL.Query().Get("to_date")

	var fromDate, toDate *time.Time
	if fromDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromDateStr); err == nil {
			fromDate = &parsed
		}
	}
	if toDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toDateStr); err == nil {
			toDate = &parsed
		}
	}

	period := &TimePeriod{}
	if fromDate != nil {
		period.From = *fromDate
	}
	if toDate != nil {
		period.To = *toDate
	}

	metrics, err := h.bclbService.GetComplianceMetrics(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get compliance metrics", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    metrics,
	}, http.StatusOK)
}

// GetComplianceAlerts returns compliance alerts
func (h *ComplianceHandler) GetComplianceAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	alerts, err := h.bclbService.GetComplianceAlerts(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get compliance alerts", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    alerts,
	}, http.StatusOK)
}

// AcknowledgeComplianceAlert acknowledges a compliance alert
func (h *ComplianceHandler) AcknowledgeComplianceAlert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AlertID string `json:"alert_id"`
		Notes   string `json:"notes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.bclbService.AcknowledgeComplianceAlert(r.Context(), req.AlertID)
	if err != nil {
		WriteError(w, err, "Failed to acknowledge compliance alert", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"message": "Compliance alert acknowledged successfully",
	}, http.StatusOK)
}

// ResolveComplianceAlert resolves a compliance alert
func (h *ComplianceHandler) ResolveComplianceAlert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AlertID    string `json:"alert_id"`
		Resolution string `json:"resolution"`
		Notes      string `json:"notes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.bclbService.ResolveComplianceAlert(r.Context(), req.AlertID)
	if err != nil {
		WriteError(w, err, "Failed to resolve compliance alert", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"message": "Compliance alert resolved successfully",
	}, http.StatusOK)
}

// GetComplianceRules returns compliance rules
func (h *ComplianceHandler) GetComplianceRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	rules, err := h.bclbService.GetComplianceRules(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get compliance rules", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    rules,
	}, http.StatusOK)
}

// CreateComplianceRule creates a new compliance rule
func (h *ComplianceHandler) CreateComplianceRule(w http.ResponseWriter, r *http.Request) {
	var req ComplianceRule
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.bclbService.CreateComplianceRule(r.Context(), &req)
	if err != nil {
		WriteError(w, err, "Failed to create compliance rule", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    req,
		"message": "Compliance rule created successfully",
	}, http.StatusCreated)
}

// UpdateComplianceRule updates an existing compliance rule
func (h *ComplianceHandler) UpdateComplianceRule(w http.ResponseWriter, r *http.Request) {
	var req ComplianceRule
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract rule ID from URL path
	path := r.URL.Path[len("/api/compliance/rules/"):]
	if path == "" {
		WriteError(w, nil, "Rule ID is required", http.StatusBadRequest)
		return
	}

	err := h.bclbService.UpdateComplianceRule(r.Context(), path, &req)
	if err != nil {
		WriteError(w, err, "Failed to update compliance rule", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    req,
		"message": "Compliance rule updated successfully",
	}, http.StatusOK)
}

// DeleteComplianceRule deletes a compliance rule
func (h *ComplianceHandler) DeleteComplianceRule(w http.ResponseWriter, r *http.Request) {
	// Extract rule ID from URL path
	path := r.URL.Path[len("/api/compliance/rules/"):]
	if path == "" {
		WriteError(w, nil, "Rule ID is required", http.StatusBadRequest)
		return
	}

	err := h.bclbService.DeleteComplianceRule(r.Context(), path)
	if err != nil {
		WriteError(w, err, "Failed to delete compliance rule", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"message": "Compliance rule deleted successfully",
	}, http.StatusOK)
}

// GetComplianceSettings returns compliance settings
func (h *ComplianceHandler) GetComplianceSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	settings, err := h.bclbService.GetComplianceSettings(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get compliance settings", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    settings,
	}, http.StatusOK)
}

// UpdateComplianceSettings updates compliance settings
func (h *ComplianceHandler) UpdateComplianceSettings(w http.ResponseWriter, r *http.Request) {
	var req ComplianceSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.bclbService.UpdateComplianceSettings(r.Context(), &req)
	if err != nil {
		WriteError(w, err, "Failed to update compliance settings", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"message": "Compliance settings updated successfully",
	}, http.StatusOK)
}

// RegisterRoutes registers compliance routes
func (h *ComplianceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/compliance/bet/validate", h.ValidateBetPlacement)
	mux.HandleFunc("/api/compliance/user/status/", h.GetUserComplianceStatus)
	mux.HandleFunc("/api/compliance/user/limits/", h.SetUserLimits)
	mux.HandleFunc("/api/compliance/user/restrictions/", h.AddUserRestriction)
	mux.HandleFunc("/api/compliance/report/generate", h.GenerateComplianceReport)
	mux.HandleFunc("/api/compliance/metrics", h.GetComplianceMetrics)
	mux.HandleFunc("/api/compliance/alerts", h.GetComplianceAlerts)
	mux.HandleFunc("/api/compliance/alerts/acknowledge", h.AcknowledgeComplianceAlert)
	mux.HandleFunc("/api/compliance/alerts/resolve", h.ResolveComplianceAlert)
	mux.HandleFunc("/api/compliance/rules", h.GetComplianceRules)
	mux.HandleFunc("/api/compliance/rules/", h.UpdateComplianceRule)
	mux.HandleFunc("/api/compliance/rules/create", h.CreateComplianceRule)
	mux.HandleFunc("/api/compliance/rules/delete/", h.DeleteComplianceRule)
	mux.HandleFunc("/api/compliance/settings", h.GetComplianceSettings)
	mux.HandleFunc("/api/compliance/settings/update", h.UpdateComplianceSettings)
}
