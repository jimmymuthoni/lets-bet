package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/betting-platform/internal/sports/live"
)

// RegisterRoutes registers live betting routes
func (h *LiveHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/live/matches", h.GetLiveMatches)
	mux.HandleFunc("/api/live/matches/", h.GetLiveMatch)
	mux.HandleFunc("/api/live/bet", h.PlaceLiveBet)
	mux.HandleFunc("/api/live/odds/update", h.UpdateOdds)
	mux.HandleFunc("/api/live/match/suspend", h.SuspendMatch)
	mux.HandleFunc("/api/live/metrics", h.GetMetrics)
}

// GetLiveMatches returns all live matches
func (h *LiveHandler) GetLiveMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		WriteError(w, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	matches, err := h.liveService.GetLiveMatches(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get live matches", http.StatusInternalServerError)
		return
	}

	response := &GetLiveMatchesResponse{
		Success: true,
		Data:    convertToLiveMatches(matches),
	}

	WriteJSON(w, response, http.StatusOK)
}

// GetLiveMatch returns a specific live match
func (h *LiveHandler) GetLiveMatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		WriteError(w, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract match ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/live/matches/")
	if path == "" {
		WriteError(w, nil, "Match ID is required", http.StatusBadRequest)
		return
	}

	match, err := h.liveService.GetLiveMatch(ctx, path)
	if err != nil {
		WriteError(w, err, "Failed to get live match", http.StatusNotFound)
		return
	}

	response := &GetLiveMatchResponse{
		Success: true,
		Data:    convertToLiveMatch(match),
	}

	WriteJSON(w, response, http.StatusOK)
}

// PlaceLiveBet places a live bet
func (h *LiveHandler) PlaceLiveBet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		WriteError(w, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LiveBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateBetRequest(&req); err != nil {
		WriteError(w, err, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get user ID
	userID := getUserID(ctx)
	if userID == "" {
		WriteError(w, nil, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Place bet
	response, err := h.liveService.PlaceLiveBet(ctx, &live.LiveBetRequest{
		MatchID:   req.MatchID,
		MarketID:  req.MarketID,
		OutcomeID: req.OutcomeID,
		UserID:    userID,
		Amount:    req.Amount,
		Odds:      req.Odds,
	})
	if err != nil {
		WriteError(w, err, "Failed to place bet", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, response, http.StatusOK)
}

// UpdateOdds updates odds for a live match
func (h *LiveHandler) UpdateOdds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		WriteError(w, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OddsUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateOddsUpdateRequest(&req); err != nil {
		WriteError(w, err, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update odds
	err := h.liveService.UpdateLiveOdds(ctx, &live.OddsUpdateRequest{
		MatchID:   req.MatchID,
		MarketID:  req.MarketID,
		OutcomeID: req.OutcomeID,
		NewOdds:   req.NewOdds,
	})
	if err != nil {
		WriteError(w, err, "Failed to update odds", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]bool{"success": true}, http.StatusOK)
}

// SuspendMatch suspends a live match
func (h *LiveHandler) SuspendMatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		WriteError(w, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SuspendMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.MatchID == "" {
		WriteError(w, nil, "Match ID is required", http.StatusBadRequest)
		return
	}

	// Suspend match
	err := h.liveService.SuspendMatch(ctx, req.MatchID, req.Reason)
	if err != nil {
		WriteError(w, err, "Failed to suspend match", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]bool{"success": true}, http.StatusOK)
}

// GetMetrics returns live betting metrics
func (h *LiveHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		WriteError(w, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics, err := h.liveService.GetMetrics(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get metrics", http.StatusInternalServerError)
		return
	}

	response := &LiveMetricsResponse{
		Success: true,
		Data: &LiveBettingMetrics{
			TotalMatches:       metrics.TotalMatches,
			ActiveMatches:      metrics.ActiveMatches,
			SuspendedMatches:   metrics.SuspendedMatches,
			OddsUpdateInterval: metrics.OddsUpdateInterval,
			LastActivity:       metrics.LastActivity,
		},
	}

	WriteJSON(w, response, http.StatusOK)
}

// convertToLiveMatches converts live service matches to HTTP response format
func convertToLiveMatches(matches interface{}) []LiveMatch {
	// Implementation stub - return empty slice for now
	return []LiveMatch{}
}

// convertToLiveMatch converts a live service match to HTTP response format
func convertToLiveMatch(match interface{}) *LiveMatch {
	// Implementation stub - return nil for now
	return nil
}

// validateBetRequest validates a live bet request
func (h *LiveHandler) validateBetRequest(req interface{}) error {
	// Implementation stub
	return nil
}

// validateOddsUpdateRequest validates an odds update request
func (h *LiveHandler) validateOddsUpdateRequest(req interface{}) error {
	// Implementation stub
	return nil
}
