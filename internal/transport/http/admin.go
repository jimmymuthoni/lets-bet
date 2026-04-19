package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/betting-platform/internal/admin"
)

// GetDashboardData returns comprehensive dashboard data
func (h *AdminHandler) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dashboardData, err := h.adminService.GetDashboardData(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get dashboard data", http.StatusInternalServerError)
		return
	}

	response := &DashboardResponse{
		Success: true,
		Data:    dashboardData,
	}

	WriteJSON(w, response, http.StatusOK)
}

// GetUserManagementData returns user management data with pagination
func (h *AdminHandler) GetUserManagementData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")
	sortBy := r.URL.Query().Get("sort_by")
	sortDir := r.URL.Query().Get("sort_dir")

	limit := 20 // default
	offset := 0 // default

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	request := &admin.UserManagementRequest{
		Limit:   limit,
		Offset:  offset,
		Search:  search,
		Status:  status,
		SortBy:  sortBy,
		SortDir: sortDir,
	}

	userData, err := h.adminService.GetUserManagementData(ctx, request)
	if err != nil {
		WriteError(w, err, "Failed to get user management data", http.StatusInternalServerError)
		return
	}

	response := &UserManagementResponse{
		Success: true,
		Data:    userData,
	}

	WriteJSON(w, response, http.StatusOK)
}

// GetUserDetails returns detailed information about a specific user
func (h *AdminHandler) GetUserDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	if path == "" {
		WriteError(w, nil, "User ID is required", http.StatusBadRequest)
		return
	}

	userDetails, err := h.adminService.GetUserDetails(ctx, path)
	if err != nil {
		WriteError(w, err, "Failed to get user details", http.StatusNotFound)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    userDetails,
	}, http.StatusOK)
}

// PerformUserAction performs an action on a user (ban, unban, verify, etc.)
func (h *AdminHandler) PerformUserAction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req UserActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.UserID == "" || req.Action == "" {
		WriteError(w, nil, "User ID and action are required", http.StatusBadRequest)
		return
	}

	actionReq := &admin.UserActionRequest{
		UserID: req.UserID,
		Action: req.Action,
		Reason: req.Reason,
	}

	err := h.adminService.PerformUserAction(ctx, actionReq)
	if err != nil {
		WriteError(w, err, "Failed to perform user action", http.StatusInternalServerError)
		return
	}

	response := &UserActionResponse{
		Success: true,
		Message: "User action completed successfully",
		UserID:  req.UserID,
	}

	WriteJSON(w, response, http.StatusOK)
}

// GetBettingMetrics returns betting metrics and analytics
func (h *AdminHandler) GetBettingMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	timeRange := r.URL.Query().Get("time_range")
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

	request := &admin.BettingMetricsRequest{
		TimeRange: timeRange,
		FromDate:  fromDate,
		ToDate:    toDate,
	}

	metrics, err := h.adminService.GetBettingMetrics(ctx, request)
	if err != nil {
		WriteError(w, err, "Failed to get betting metrics", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    metrics,
	}, http.StatusOK)
}

// GetFinancialReports returns financial reports and analytics
func (h *AdminHandler) GetFinancialReports(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	reportType := r.URL.Query().Get("report_type")
	timeRange := r.URL.Query().Get("time_range")
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

	request := &admin.FinancialReportRequest{
		ReportType: reportType,
		TimeRange:  timeRange,
		FromDate:   fromDate,
		ToDate:     toDate,
	}

	report, err := h.adminService.GetFinancialReport(ctx, request)
	if err != nil {
		WriteError(w, err, "Failed to get financial report", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    report,
	}, http.StatusOK)
}

// GetSystemHealth returns system health status
func (h *AdminHandler) GetSystemHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	health, err := h.adminService.GetSystemHealth(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get system health", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    health,
	}, http.StatusOK)
}

// GetSystemConfig returns system configuration
func (h *AdminHandler) GetSystemConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	config, err := h.adminService.GetSystemConfig(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get system config", http.StatusInternalServerError)
		return
	}

	response := &SystemConfigResponse{
		Success: true,
		Config:  config,
	}

	WriteJSON(w, response, http.StatusOK)
}

// UpdateSystemConfig updates system configuration
func (h *AdminHandler) UpdateSystemConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req SystemConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.adminService.UpdateSystemConfig(ctx, req.Config)
	if err != nil {
		WriteError(w, err, "Failed to update system config", http.StatusInternalServerError)
		return
	}

	response := &SystemConfigResponse{
		Success: true,
		Message: "System configuration updated successfully",
		Config:  req.Config,
	}

	WriteJSON(w, response, http.StatusOK)
}

// GetAuditLogs returns audit logs with filtering
func (h *AdminHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	action := r.URL.Query().Get("action")
	userID := r.URL.Query().Get("user_id")
	fromDateStr := r.URL.Query().Get("from_date")
	toDateStr := r.URL.Query().Get("to_date")

	limit := 50 // default
	offset := 0 // default

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

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

	request := &admin.AuditLogRequest{
		Limit:    limit,
		Offset:   offset,
		Action:   action,
		UserID:   userID,
		FromDate: fromDate,
		ToDate:   toDate,
	}

	logs, err := h.adminService.GetAuditLogs(ctx, request)
	if err != nil {
		WriteError(w, err, "Failed to get audit logs", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]interface{}{
		"success": true,
		"data":    logs,
	}, http.StatusOK)
}

// RegisterRoutes registers admin routes
func (h *AdminHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/admin/dashboard", h.GetDashboardData)
	mux.HandleFunc("/api/admin/users", h.GetUserManagementData)
	mux.HandleFunc("/api/admin/users/", h.GetUserDetails)
	mux.HandleFunc("/api/admin/users/action", h.PerformUserAction)
	mux.HandleFunc("/api/admin/betting/metrics", h.GetBettingMetrics)
	mux.HandleFunc("/api/admin/financial/reports", h.GetFinancialReports)
	mux.HandleFunc("/api/admin/system/health", h.GetSystemHealth)
	mux.HandleFunc("/api/admin/system/config", h.GetSystemConfig)
	mux.HandleFunc("/api/admin/system/config/update", h.UpdateSystemConfig)
	mux.HandleFunc("/api/admin/audit/logs", h.GetAuditLogs)
}
