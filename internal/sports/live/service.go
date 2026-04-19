// Package live provides live sports betting functionality
package live

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/infrastructure/repository/postgres"
	"github.com/betting-platform/internal/odds/genius"
	"github.com/betting-platform/internal/odds/sportradar"
)

// LiveBettingService provides live sports betting functionality
type LiveBettingService struct {
	matchRepo        postgres.MatchRepository
	betRepo          postgres.SportBetRepository
	marketRepo       postgres.BettingMarketRepository
	outcomeRepo      postgres.MarketOutcomeRepository
	sportradarClient *sportradar.SportradarClient
	geniusClient     *genius.GeniusClient
	eventBus         EventBus

	// Live data cache
	liveMatches      map[string]*LiveMatch
	liveMatchesMutex sync.RWMutex

	// Odds update cache
	oddsUpdates      map[string]*OddsUpdate
	oddsUpdatesMutex sync.RWMutex

	// Settlement queue
	settlementQueue chan *SettlementRequest

	// Configuration
	oddsUpdateInterval time.Duration
	settlementInterval time.Duration
	maxOddsDelay       time.Duration
}

// EventBus interface for event publishing
type EventBus interface {
	Publish(topic string, event any) error
	Subscribe(topic string, handler func(any)) error
}

// LiveMatch represents a live match with real-time data
type LiveMatch struct {
	Match *domain.Match
	// Live-specific data
	CurrentMinute   int
	HomeScore       int
	AwayScore       int
	HomePossession  float64
	AwayPossession  float64
	HomeCorners     int
	AwayCorners     int
	HomeYellowCards int
	AwayYellowCards int
	HomeRedCards    int
	AwayRedCards    int
	// Live markets
	LiveMarkets []*LiveMarket
	// Timing
	LastUpdated    time.Time
	NextOddsUpdate time.Time
	// Status
	IsSuspended      bool
	SuspensionReason string
}

// LiveMarket represents a live betting market
type LiveMarket struct {
	Market *domain.Market
	// Live-specific data
	IsLive          bool
	LiveOdds        []*LiveOutcome
	LastOddsUpdate  time.Time
	OddsUpdateCount int
	// Market status
	IsSuspended      bool
	SuspensionReason string
}

// LiveOutcome represents a live betting outcome
type LiveOutcome struct {
	*domain.Outcome

	// Live-specific data
	CurrentOdds     decimal.Decimal
	PreviousOdds    decimal.Decimal
	OddsChangeTime  time.Time
	OddsChangeCount int

	// Volume data
	TotalVolume decimal.Decimal
	LiveVolume  decimal.Decimal
}

// OddsUpdate represents an odds update event
type OddsUpdate struct {
	MatchID    string
	MarketID   string
	OutcomeID  string
	OldOdds    decimal.Decimal
	NewOdds    decimal.Decimal
	UpdateTime time.Time
	UpdateType string // "automatic", "manual", "suspension"
}

// SettlementRequest represents a settlement request
type SettlementRequest struct {
	MatchID   string
	MarketID  string
	OutcomeID string
	Status    domain.OutcomeStatus
	SettledAt time.Time
	Reason    string
}

// NewLiveBettingService creates a new live betting service
func NewLiveBettingService(
	matchRepo postgres.MatchRepository,
	betRepo postgres.SportBetRepository,
	marketRepo postgres.BettingMarketRepository,
	outcomeRepo postgres.MarketOutcomeRepository,
	sportradarClient *sportradar.SportradarClient,
	geniusClient *genius.GeniusClient,
	eventBus EventBus,
) *LiveBettingService {
	service := &LiveBettingService{
		matchRepo:          matchRepo,
		betRepo:            betRepo,
		marketRepo:         marketRepo,
		outcomeRepo:        outcomeRepo,
		sportradarClient:   sportradarClient,
		geniusClient:       geniusClient,
		eventBus:           eventBus,
		liveMatches:        make(map[string]*LiveMatch),
		oddsUpdates:        make(map[string]*OddsUpdate),
		settlementQueue:    make(chan *SettlementRequest, 1000),
		oddsUpdateInterval: 10 * time.Second,
		settlementInterval: 30 * time.Second,
		maxOddsDelay:       5 * time.Second,
	}

	return service
}

// Start starts the live betting service
func (s *LiveBettingService) Start(ctx context.Context) error {
	log.Println("Starting live betting service...")

	// Start live data fetcher
	go s.liveDataFetcher(ctx)

	// Start odds updater
	go s.oddsUpdater(ctx)

	// Start settlement processor
	go s.settlementProcessor(ctx)

	// Start cleanup routine
	go s.cleanupRoutine(ctx)

	log.Println("Live betting service started")
	return nil
}

// Stop stops the live betting service
func (s *LiveBettingService) Stop() error {
	log.Println("Stopping live betting service...")

	// Close channels
	close(s.settlementQueue)

	log.Println("Live betting service stopped")
	return nil
}

// liveDataFetcher fetches live data from odds providers
func (s *LiveBettingService) liveDataFetcher(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.fetchLiveData(ctx)
		}
	}
}

// fetchLiveData fetches live data from all providers
func (s *LiveBettingService) fetchLiveData(ctx context.Context) {
	// Fetch from Sportradar
	if s.sportradarClient != nil {
		s.fetchSportradarLiveData(ctx)
	}

	// Fetch from Genius Sports
	if s.geniusClient != nil {
		s.fetchGeniusLiveData(ctx)
	}
}

// fetchSportradarLiveData fetches live data from Sportradar
func (s *LiveBettingService) fetchSportradarLiveData(ctx context.Context) {
	liveMatches, err := s.sportradarClient.GetLiveMatches(ctx)
	if err != nil {
		log.Printf("Error fetching Sportradar live matches: %v", err)
		return
	}

	for _, match := range liveMatches {
		domainMatch := s.sportradarClient.ConvertToDomainMatch(match)
		s.processLiveMatch(ctx, *domainMatch, "sportradar")
	}
}

// fetchGeniusLiveData fetches live data from Genius Sports
func (s *LiveBettingService) fetchGeniusLiveData(ctx context.Context) {
	response, err := s.geniusClient.GetLiveMatches(ctx)
	if err != nil {
		log.Printf("Error fetching Genius live matches: %v", err)
		return
	}

	for _, match := range response.Data {
		domainMatch := s.geniusClient.ConvertToDomainMatch(match)
		s.processLiveMatch(ctx, *domainMatch, "genius")
	}
}

// processLiveMatch processes a live match from an odds provider
func (s *LiveBettingService) processLiveMatch(ctx context.Context, match domain.Match, provider string) {
	s.liveMatchesMutex.Lock()
	defer s.liveMatchesMutex.Unlock()

	// Check if we already have this match
	liveMatch, exists := s.liveMatches[match.ID]
	if !exists {
		// Create new live match
		liveMatch = &LiveMatch{
			Match:          &match,
			CurrentMinute:  s.calculateCurrentMinute(&match),
			HomeScore:      0,
			AwayScore:      0,
			LiveMarkets:    make([]*LiveMarket, 0),
			LastUpdated:    time.Now(),
			NextOddsUpdate: time.Now().Add(s.oddsUpdateInterval),
		}
		s.liveMatches[match.ID] = liveMatch

		// Publish live match started event
		s.publishEvent("live.match.started", map[string]any{
			"match_id": match.ID,
			"provider": provider,
		})
	}

	// Update match data
	s.updateLiveMatchData(liveMatch, &match)

	// Update markets
	s.updateLiveMarkets(ctx, liveMatch, &match)
}

// updateLiveMatchData updates live match data
func (s *LiveBettingService) updateLiveMatchData(liveMatch *LiveMatch, match *domain.Match) {
	// Update score if available
	if match.Score != nil {
		liveMatch.HomeScore = match.Score.HomeScore
		liveMatch.AwayScore = match.Score.AwayScore
	}

	// Update status
	liveMatch.Match.Status = match.Status
	liveMatch.LastUpdated = time.Now()

	// Check if match should be suspended
	shouldSuspend := s.shouldSuspendMatch(liveMatch)
	if shouldSuspend && !liveMatch.IsSuspended {
		liveMatch.IsSuspended = true
		liveMatch.SuspensionReason = "Automatic suspension due to score change"

		// Publish suspension event
		s.publishEvent("live.match.suspended", map[string]any{
			"match_id": liveMatch.Match.ID,
			"reason":   liveMatch.SuspensionReason,
		})
	} else if !shouldSuspend && liveMatch.IsSuspended {
		liveMatch.IsSuspended = false
		liveMatch.SuspensionReason = ""

		// Publish resumption event
		s.publishEvent("live.match.resumed", map[string]any{
			"match_id": liveMatch.Match.ID,
		})
	}
}

// updateLiveMarkets updates live markets for a match
func (s *LiveBettingService) updateLiveMarkets(ctx context.Context, liveMatch *LiveMatch, match *domain.Match) {
	for _, market := range match.Markets {
		liveMarket := s.findOrCreateLiveMarket(liveMatch, &market)
		s.updateLiveMarket(ctx, liveMarket, &market)
	}
}

// findOrCreateLiveMarket finds or creates a live market
func (s *LiveBettingService) findOrCreateLiveMarket(liveMatch *LiveMatch, market *domain.Market) *LiveMarket {
	for _, lm := range liveMatch.LiveMarkets {
		if lm.Market.ID == market.ID {
			return lm
		}
	}

	// Create new live market
	liveMarket := &LiveMarket{
		Market:         market,
		IsLive:         true,
		LiveOdds:       make([]*LiveOutcome, 0),
		LastOddsUpdate: time.Now(),
	}

	liveMatch.LiveMarkets = append(liveMatch.LiveMarkets, liveMarket)
	return liveMarket
}

// updateLiveMarket updates live market data
func (s *LiveBettingService) updateLiveMarket(ctx context.Context, liveMarket *LiveMarket, market *domain.Market) {
	// Update market status
	liveMarket.Market.Status = market.Status

	// Update outcomes
	for _, outcome := range market.Outcomes {
		liveOutcome := s.findOrCreateLiveOutcome(liveMarket, &outcome)
		s.updateLiveOutcome(ctx, liveOutcome, &outcome)
	}

	liveMarket.LastOddsUpdate = time.Now()
	liveMarket.OddsUpdateCount++
}

func (s *LiveBettingService) findOrCreateLiveOutcome(liveMarket *LiveMarket, outcome *domain.Outcome) *LiveOutcome {
	for _, lo := range liveMarket.LiveOdds {
		if lo.Outcome.ID == outcome.ID {
			return lo
		}
	}

	// Create new live outcome
	liveOutcome := &LiveOutcome{
		Outcome:         outcome,
		CurrentOdds:     outcome.Odds,
		PreviousOdds:    outcome.Odds,
		OddsChangeTime:  time.Now(),
		OddsChangeCount: 0,
		TotalVolume:     decimal.Zero,
		LiveVolume:      decimal.Zero,
	}

	liveMarket.LiveOdds = append(liveMarket.LiveOdds, liveOutcome)
	return liveOutcome
}

// updateLiveOutcome updates live outcome data
func (s *LiveBettingService) updateLiveOutcome(_ context.Context, liveOutcome *LiveOutcome, outcome *domain.Outcome) {
	// Check if odds changed
	if !outcome.Odds.Equal(liveOutcome.CurrentOdds) {
		liveOutcome.PreviousOdds = liveOutcome.CurrentOdds
		liveOutcome.CurrentOdds = outcome.Odds
		liveOutcome.OddsChangeTime = time.Now()
		liveOutcome.OddsChangeCount++

		// Create odds update
		oddsUpdate := &OddsUpdate{
			MatchID:    liveOutcome.MarketID, // This should be match ID, need to fix
			MarketID:   liveOutcome.MarketID,
			OutcomeID:  liveOutcome.ID,
			OldOdds:    liveOutcome.PreviousOdds,
			NewOdds:    liveOutcome.CurrentOdds,
			UpdateTime: time.Now(),
			UpdateType: "automatic",
		}

		s.recordOddsUpdate(oddsUpdate)

		// Publish odds update event
		s.publishEvent("live.odds.updated", map[string]any{
			"match_id":   oddsUpdate.MatchID,
			"market_id":  oddsUpdate.MarketID,
			"outcome_id": oddsUpdate.OutcomeID,
			"old_odds":   oddsUpdate.OldOdds,
			"new_odds":   oddsUpdate.NewOdds,
		})
	}

	// Update outcome status
	liveOutcome.Status = outcome.Status
}

// oddsUpdater processes odds updates
func (s *LiveBettingService) oddsUpdater(ctx context.Context) {
	ticker := time.NewTicker(s.oddsUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processOddsUpdates(ctx)
		}
	}
}

// processOddsUpdates processes pending odds updates
func (s *LiveBettingService) processOddsUpdates(ctx context.Context) {
	s.oddsUpdatesMutex.Lock()
	defer s.oddsUpdatesMutex.Unlock()

	for id, update := range s.oddsUpdates {
		// Check if update is still valid
		if time.Since(update.UpdateTime) > s.maxOddsDelay {
			delete(s.oddsUpdates, id)
			continue
		}

		// Process the update
		err := s.applyOddsUpdate(ctx, update)
		if err != nil {
			log.Printf("Error applying odds update: %v", err)
		}

		// Remove processed update
		delete(s.oddsUpdates, id)
	}
}

// applyOddsUpdate applies an odds update to the database
func (s *LiveBettingService) applyOddsUpdate(ctx context.Context, update *OddsUpdate) error {
	// Update outcome odds in database
	return s.outcomeRepo.UpdateOdds(ctx, update.OutcomeID, update.NewOdds)
}

// recordOddsUpdate records an odds update
func (s *LiveBettingService) recordOddsUpdate(update *OddsUpdate) {
	s.oddsUpdatesMutex.Lock()
	defer s.oddsUpdatesMutex.Unlock()

	s.oddsUpdates[update.OutcomeID] = update
}

// settlementProcessor processes settlement requests
func (s *LiveBettingService) settlementProcessor(ctx context.Context) {
	ticker := time.NewTicker(s.settlementInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processSettlements(ctx)
		case request := <-s.settlementQueue:
			s.processSettlementRequest(ctx, request)
		}
	}
}

// processSettlements processes pending settlements
func (s *LiveBettingService) processSettlements(_ context.Context) {
	s.liveMatchesMutex.RLock()
	defer s.liveMatchesMutex.RUnlock()

	for _, liveMatch := range s.liveMatches {
		if liveMatch.Match.Status == domain.MatchStatusFinished || liveMatch.Match.Status == domain.MatchStatusCompleted {
			s.settleMatch(context.Background(), liveMatch)
		}
	}
}

// settleMatch settles a finished match
func (s *LiveBettingService) settleMatch(ctx context.Context, liveMatch *LiveMatch) {
	for _, liveMarket := range liveMatch.LiveMarkets {
		s.settleMarket(ctx, liveMatch, liveMarket)
	}
}

// settleMarket settles a market
func (s *LiveBettingService) settleMarket(ctx context.Context, liveMatch *LiveMatch, liveMarket *LiveMarket) {
	for _, liveOutcome := range liveMarket.LiveOdds {
		// Determine outcome status based on match result
		status := s.determineOutcomeStatus(ctx, liveMatch, liveMarket, liveOutcome)

		if status != liveOutcome.Outcome.Status {
			// Create settlement request
			request := &SettlementRequest{
				MatchID:   liveMatch.Match.ID,
				MarketID:  liveMarket.Market.ID,
				OutcomeID: liveOutcome.Outcome.ID,
				Status:    status,
				SettledAt: time.Now(),
				Reason:    "Automatic settlement",
			}

			s.settlementQueue <- request
		}
	}
}

// ... (rest of the code remains the same)
// determineOutcomeStatus determines the final status of an outcome
func (s *LiveBettingService) determineOutcomeStatus(_ context.Context, liveMatch *LiveMatch, liveMarket *LiveMarket, liveOutcome *LiveOutcome) domain.OutcomeStatus {
	// This is a simplified implementation
	// In reality, this would be much more complex based on market type and match result

	if liveMatch.Match.Status == domain.MatchStatusCancelled || liveMatch.Match.Status == domain.MatchStatusAbandoned {
		return domain.OutcomeStatusVoid
	}

	// For match winner markets
	if liveMarket.Market.Type == domain.MarketTypeMatchWinner {
		if liveOutcome.Outcome.Name == liveMatch.Match.HomeTeam && liveMatch.HomeScore > liveMatch.AwayScore {
			return domain.OutcomeStatusWon
		} else if liveOutcome.Outcome.Name == liveMatch.Match.AwayTeam && liveMatch.AwayScore > liveMatch.HomeScore {
			return domain.OutcomeStatusWon
		} else if liveOutcome.Outcome.Name == "Draw" && liveMatch.HomeScore == liveMatch.AwayScore {
			return domain.OutcomeStatusWon
		} else {
			return domain.OutcomeStatusLost
		}
	}

	// For other market types, implement specific logic
	return domain.OutcomeStatusPending
}

// processSettlementRequest processes a settlement request
func (s *LiveBettingService) processSettlementRequest(ctx context.Context, request *SettlementRequest) {
	// Update outcome status
	err := s.outcomeRepo.UpdateStatus(ctx, request.OutcomeID, request.Status)
	if err != nil {
		log.Printf("Error updating outcome status: %v", err)
		return
	}

	// Update market status
	err = s.marketRepo.UpdateStatus(ctx, request.MarketID, domain.MarketStatusSettled)
	if err != nil {
		log.Printf("Error updating market status: %v", err)
		return
	}

	// Settle bets
	err = s.settleBets(ctx, request)
	if err != nil {
		log.Printf("Error settling bets: %v", err)
		return
	}

	// Publish settlement event
	s.publishEvent("live.bet.settled", map[string]any{
		"match_id":   request.MatchID,
		"market_id":  request.MarketID,
		"outcome_id": request.OutcomeID,
		"status":     request.Status,
		"settled_at": request.SettledAt,
	})
}

// settleBets settles bets for an outcome
func (s *LiveBettingService) settleBets(ctx context.Context, request *SettlementRequest) error {
	// Get all bets for this outcome
	bets, err := s.betRepo.GetByOutcome(ctx, request.OutcomeID)
	if err != nil {
		return fmt.Errorf("get bets by outcome: %w", err)
	}

	for _, bet := range bets {
		// Determine bet status
		betStatus := s.determineBetStatus(bet, request.Status)

		// Update bet
		bet.Status = betStatus
		bet.SettledAt = &request.SettledAt

		err := s.betRepo.Update(ctx, bet)
		if err != nil {
			log.Printf("Error updating bet %s: %v", bet.ID, err)
			continue
		}

		// Process payout if bet won
		if betStatus == domain.BetStatusWon {
			err := s.processPayout(ctx, bet)
			if err != nil {
				log.Printf("Error processing payout for bet %s: %v", bet.ID, err)
			}
		}
	}

	return nil
}

// determineBetStatus determines the status of a bet
func (s *LiveBettingService) determineBetStatus(_ *domain.SportBet, outcomeStatus domain.OutcomeStatus) domain.BetStatus {
	switch outcomeStatus {
	case domain.OutcomeStatusWon:
		return domain.BetStatusWon
	case domain.OutcomeStatusLost:
		return domain.BetStatusLost
	case domain.OutcomeStatusVoid:
		return domain.BetStatusVoid
	default:
		return domain.BetStatusPending
	}
}

// processPayout processes a winning bet payout
func (s *LiveBettingService) processPayout(ctx context.Context, bet *domain.SportBet) error {
	// Calculate payout
	payout := bet.Amount.Mul(bet.Odds)

	// Update bet payout
	bet.Payout = payout

	err := s.betRepo.Update(ctx, bet)
	if err != nil {
		return fmt.Errorf("update bet payout: %w", err)
	}

	// Publish payout event
	s.publishEvent("live.payout.processed", map[string]any{
		"bet_id":  bet.ID,
		"user_id": bet.UserID,
		"payout":  payout,
	})

	return nil
}

// cleanupRoutine performs periodic cleanup
func (s *LiveBettingService) cleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanup(ctx)
		}
	}
}

// cleanup removes old live matches
func (s *LiveBettingService) cleanup(_ context.Context) {
	s.liveMatchesMutex.Lock()
	defer s.liveMatchesMutex.Unlock()

	// Remove finished matches older than 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	for id, liveMatch := range s.liveMatches {
		if liveMatch.Match.Status == domain.MatchStatusFinished || liveMatch.Match.Status == domain.MatchStatusCompleted {
			if liveMatch.LastUpdated.Before(cutoff) {
				delete(s.liveMatches, id)
			}
		}
	}

	// Clean up old odds updates
	s.oddsUpdatesMutex.Lock()
	defer s.oddsUpdatesMutex.Unlock()

	for id, update := range s.oddsUpdates {
		if time.Since(update.UpdateTime) > 1*time.Hour {
			delete(s.oddsUpdates, id)
		}
	}
}

// calculateCurrentMinute calculates the current minute of a match
func (s *LiveBettingService) calculateCurrentMinute(match *domain.Match) int {
	elapsed := time.Since(match.StartTime)
	return int(elapsed.Minutes())
}

// shouldSuspendMatch determines if a match should be suspended
func (s *LiveBettingService) shouldSuspendMatch(liveMatch *LiveMatch) bool {
	// Suspend if score changed recently
	if time.Since(liveMatch.LastUpdated) < 30*time.Second {
		return true
	}

	// Suspend if red card was given
	if liveMatch.HomeRedCards > 0 || liveMatch.AwayRedCards > 0 {
		return true
	}

	// Suspend if penalty
	// This would need more sophisticated logic

	return false
}

// publishEvent publishes an event to the event bus
func (s *LiveBettingService) publishEvent(topic string, data any) {
	if s.eventBus != nil {
		err := s.eventBus.Publish(topic, data)
		if err != nil {
			log.Printf("Error publishing event %s: %v", topic, err)
		}
	}
}

// GetLiveMatches returns all live matches
func (s *LiveBettingService) GetLiveMatches(ctx context.Context) ([]*LiveMatch, error) {
	s.liveMatchesMutex.RLock()
	defer s.liveMatchesMutex.RUnlock()

	matches := make([]*LiveMatch, 0, len(s.liveMatches))
	for _, liveMatch := range s.liveMatches {
		matches = append(matches, liveMatch)
	}

	return matches, nil
}

// GetLiveMatch returns a specific live match by ID
func (s *LiveBettingService) GetLiveMatch(ctx context.Context, matchID string) (*LiveMatch, error) {
	s.liveMatchesMutex.RLock()
	defer s.liveMatchesMutex.RUnlock()

	liveMatch, exists := s.liveMatches[matchID]
	if !exists {
		return nil, fmt.Errorf("live match not found: %s", matchID)
	}

	return liveMatch, nil
}

// PlaceLiveBet places a live bet
func (s *LiveBettingService) PlaceLiveBet(ctx context.Context, bet *domain.SportBet) error {
	// Validate that the match is live and accepting bets
	liveMatch, err := s.GetLiveMatch(ctx, bet.EventID)
	if err != nil {
		return fmt.Errorf("match not found or not live: %w", err)
	}

	// Check if match is still accepting bets
	if liveMatch.Match.Status != domain.MatchStatusLive {
		return fmt.Errorf("match is not accepting live bets")
	}

	// TODO: Implement actual bet placement logic
	// This would integrate with the existing betting system
	log.Printf("Live bet placed: %+v", bet)
	return nil
}

// findLiveMarket finds a live market by ID
func (s *LiveBettingService) findLiveMarket(matchID, marketID string) *LiveMarket {
	s.liveMatchesMutex.RLock()
	defer s.liveMatchesMutex.RUnlock()

	liveMatch, exists := s.liveMatches[matchID]
	if !exists {
		return nil
	}

	for _, market := range liveMatch.LiveMarkets {
		if market.Market.ID == marketID {
			return market
		}
	}

	return nil
}
