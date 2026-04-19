package virtualsports

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/infrastructure/repository/postgres"
)

// VirtualSportsService manages virtual sports games and betting
type VirtualSportsService struct {
	matchRepo     postgres.MatchRepository
	marketRepo    postgres.BettingMarketRepository
	outcomeRepo   postgres.MarketOutcomeRepository
	betRepo       postgres.SportBetRepository
	walletService WalletService
	eventBus      EventBus
	rng           *rand.Rand
	mu            sync.RWMutex

	// Virtual game state
	games     map[string]*VirtualGame
	schedules map[string]*GameSchedule
}

// WalletService interface for wallet operations
type WalletService interface {
	Credit(ctx context.Context, userID string, amount decimal.Decimal, movement Movement) (*Transaction, error)
	Debit(ctx context.Context, userID string, amount decimal.Decimal, movement Movement) (*Transaction, error)
}

// Movement represents a wallet movement
type Movement struct {
	UserID        string                 `json:"user_id"`
	Amount        decimal.Decimal        `json:"amount"`
	Type          domain.TransactionType `json:"type"`
	ReferenceID   *string                `json:"reference_id,omitempty"`
	ReferenceType string                 `json:"reference_type"`
	Description   string                 `json:"description"`
	ProviderName  string                 `json:"provider_name"`
	ProviderTxnID string                 `json:"provider_txn_id"`
	CountryCode   string                 `json:"country_code"`
}

// Transaction represents a wallet transaction
type Transaction struct {
	ID string `json:"id"`
}

// EventBus interface for publishing events
type EventBus interface {
	Publish(topic string, data interface{}) error
}

// NewVirtualSportsService creates a new virtual sports service
func NewVirtualSportsService(
	matchRepo postgres.MatchRepository,
	marketRepo postgres.BettingMarketRepository,
	outcomeRepo postgres.MarketOutcomeRepository,
	betRepo postgres.SportBetRepository,
	walletService WalletService,
	eventBus EventBus,
) *VirtualSportsService {
	return &VirtualSportsService{
		matchRepo:     matchRepo,
		marketRepo:    marketRepo,
		outcomeRepo:   outcomeRepo,
		betRepo:       betRepo,
		walletService: walletService,
		eventBus:      eventBus,
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
		games:         make(map[string]*VirtualGame),
		schedules:     make(map[string]*GameSchedule),
	}
}

// VirtualGame represents a virtual sports game
type VirtualGame struct {
	ID           string        `json:"id"`
	Sport        domain.Sport  `json:"sport"`
	Name         string        `json:"name"`
	Status       GameStatus    `json:"status"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      *time.Time    `json:"end_time,omitempty"`
	Duration     time.Duration `json:"duration"`
	HomeTeam     *VirtualTeam  `json:"home_team"`
	AwayTeam     *VirtualTeam  `json:"away_team"`
	CurrentScore *GameScore    `json:"current_score,omitempty"`
	FinalScore   *GameScore    `json:"final_score,omitempty"`
	Events       []GameEvent   `json:"events"`
	Odds         []VirtualOdds `json:"odds"`
	CreatedAt    time.Time     `json:"created_at"`
}

type GameStatus string

const (
	GameStatusScheduled GameStatus = "SCHEDULED"
	GameStatusLive      GameStatus = "LIVE"
	GameStatusFinished  GameStatus = "FINISHED"
	GameStatusCancelled GameStatus = "CANCELLED"
)

// VirtualTeam represents a virtual sports team
type VirtualTeam struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Strength float64 `json:"strength"` // 0.0 to 1.0, affects game outcomes
	Form     float64 `json:"form"`     // Recent form, affects performance
}

// GameScore represents the current game score
type GameScore struct {
	HomeScore int `json:"home_score"`
	AwayScore int `json:"away_score"`
}

// GameEvent represents an event during the game
type GameEvent struct {
	ID      string     `json:"id"`
	Type    EventType  `json:"type"`
	Time    time.Time  `json:"time"`
	Team    string     `json:"team"`
	Player  string     `json:"player"`
	Details string     `json:"details"`
	Score   *GameScore `json:"score,omitempty"`
}

type EventType string

const (
	EventTypeKickoff      EventType = "KICKOFF"
	EventTypeGoal         EventType = "GOAL"
	EventTypeYellowCard   EventType = "YELLOW_CARD"
	EventTypeRedCard      EventType = "RED_CARD"
	EventTypeSubstitution EventType = "SUBSTITUTION"
	EventTypeHalftime     EventType = "HALFTIME"
	EventTypeFulltime     EventType = "FULLTIME"
)

// VirtualOdds represents betting odds for the game
type VirtualOdds struct {
	MarketID    string            `json:"market_id"`
	MarketName  string            `json:"market_name"`
	MarketType  domain.MarketType `json:"market_type"`
	Outcomes    []VirtualOutcome  `json:"outcomes"`
	LastUpdated time.Time         `json:"last_updated"`
}

// VirtualOutcome represents a betting outcome
type VirtualOutcome struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Odds        decimal.Decimal      `json:"odds"`
	Price       decimal.Decimal      `json:"price"`
	Status      domain.OutcomeStatus `json:"status"`
	Probability float64              `json:"probability"`
}

// GameSchedule represents when games should be generated
type GameSchedule struct {
	Sport     domain.Sport  `json:"sport"`
	Interval  time.Duration `json:"interval"`   // How often to generate games
	GameCount int           `json:"game_count"` // Number of games per interval
	NextRun   time.Time     `json:"next_run"`
	Active    bool          `json:"active"`
}

// CreateVirtualGame creates a new virtual sports game
func (s *VirtualSportsService) CreateVirtualGame(ctx context.Context, gameType domain.Sport) (*VirtualGame, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate teams
	homeTeam := s.generateTeam(gameType)
	awayTeam := s.generateTeam(gameType)

	// Calculate game duration based on sport
	duration := s.getGameDuration(gameType)

	game := &VirtualGame{
		ID:        s.generateID(),
		Sport:     gameType,
		Name:      fmt.Sprintf("%s vs %s", homeTeam.Name, awayTeam.Name),
		Status:    GameStatusScheduled,
		StartTime: time.Now().Add(5 * time.Minute), // Start in 5 minutes
		Duration:  duration,
		HomeTeam:  homeTeam,
		AwayTeam:  awayTeam,
		Events:    []GameEvent{},
		CreatedAt: time.Now(),
	}

	// Generate initial odds
	game.Odds = s.generateInitialOdds(game)

	// Save to repository
	err := s.saveGame(ctx, game)
	if err != nil {
		return nil, fmt.Errorf("failed to save virtual game: %w", err)
	}

	// Store in memory
	s.games[game.ID] = game

	// Publish event
	s.publishEvent("virtual.game.created", map[string]interface{}{
		"game_id":    game.ID,
		"sport":      game.Sport,
		"home_team":  homeTeam.Name,
		"away_team":  awayTeam.Name,
		"start_time": game.StartTime,
	})

	return game, nil
}

// generateTeam creates a virtual team with random attributes
func (s *VirtualSportsService) generateTeam(sport domain.Sport) *VirtualTeam {
	teamNames := s.getTeamNames(sport)
	name := teamNames[s.rng.Intn(len(teamNames))]

	return &VirtualTeam{
		ID:       s.generateID(),
		Name:     name,
		Strength: 0.3 + s.rng.Float64()*0.7, // 0.3 to 1.0
		Form:     0.2 + s.rng.Float64()*0.8, // 0.2 to 1.0
	}
}

// getTeamNames returns team names for a given sport
func (s *VirtualSportsService) getTeamNames(sport domain.Sport) []string {
	switch sport {
	case domain.SportFootball:
		return []string{
			"Thunder FC", "Lightning United", "Storm City", "Hurricane Athletic",
			"Tornado Rangers", "Blaze Warriors", "Cyclone FC", "Volcano United",
			"Earthquake FC", "Avalanche Athletic", "Monsoon City", "Typhoon Rangers",
		}
	case domain.SportBasketball:
		return []string{
			"Thunder Hawks", "Lightning Eagles", "Storm Wolves", "Hurricane Bears",
			"Tornado Lions", "Blaze Tigers", "Cyclone Sharks", "Volcano Panthers",
			"Earthquake Cobras", "Avalanche Vipers", "Monsoon Pythons", "Typhoon Falcons",
		}
	case domain.SportTennis:
		return []string{
			"Thunder Serve", "Lightning Smash", "Storm Volley", "Hurricane Ace",
			"Tornado Net", "Blaze Court", "Cyclone Rally", "Volcano Match",
			"Earthquake Set", "Avalanche Game", "Monsoon Point", "Typhoon Break",
		}
	default:
		return []string{"Team A", "Team B", "Team C", "Team D"}
	}
}

// getGameDuration returns the typical duration for a sport
func (s *VirtualSportsService) getGameDuration(sport domain.Sport) time.Duration {
	switch sport {
	case domain.SportFootball:
		return 90 * time.Minute // 90 minutes
	case domain.SportBasketball:
		return 48 * time.Minute // 48 minutes (4 quarters)
	case domain.SportTennis:
		return 120 * time.Minute // 2 hours average
	default:
		return 60 * time.Minute
	}
}

// generateInitialOdds creates initial betting odds for the game
func (s *VirtualSportsService) generateInitialOdds(game *VirtualGame) []VirtualOdds {
	// Calculate win probabilities based on team strengths
	homeProb := s.calculateWinProbability(game.HomeTeam, game.AwayTeam)
	awayProb := 1.0 - homeProb
	drawProb := 0.2 // Base draw probability

	// Adjust for sport
	if game.Sport == domain.SportBasketball || game.Sport == domain.SportTennis {
		drawProb = 0.0 // No draws in these sports
		homeProb = homeProb * (1.0 + drawProb)
		awayProb = awayProb * (1.0 + drawProb)
	}

	// Normalize probabilities
	total := homeProb + awayProb + drawProb
	homeProb /= total
	awayProb /= total
	drawProb /= total

	// Convert to odds (with margin)
	margin := 0.05 // 5% margin
	homeOdds := s.probabilityToOdds(homeProb * (1 - margin))
	awayOdds := s.probabilityToOdds(awayProb * (1 - margin))
	drawOdds := s.probabilityToOdds(drawProb * (1 - margin))

	odds := []VirtualOdds{
		{
			MarketID:   s.generateID(),
			MarketName: "Match Winner",
			MarketType: domain.MarketTypeMatchWinner,
			Outcomes: []VirtualOutcome{
				{
					ID:          s.generateID(),
					Name:        game.HomeTeam.Name,
					Odds:        homeOdds,
					Price:       decimal.NewFromInt(10), // $10 stake
					Status:      domain.OutcomeStatusPending,
					Probability: homeProb,
				},
				{
					ID:          s.generateID(),
					Name:        "Draw",
					Odds:        drawOdds,
					Price:       decimal.NewFromInt(10),
					Status:      domain.OutcomeStatusPending,
					Probability: drawProb,
				},
				{
					ID:          s.generateID(),
					Name:        game.AwayTeam.Name,
					Odds:        awayOdds,
					Price:       decimal.NewFromInt(10),
					Status:      domain.OutcomeStatusPending,
					Probability: awayProb,
				},
			},
			LastUpdated: time.Now(),
		},
	}

	// Add over/under markets for football
	if game.Sport == domain.SportFootball {
		overUnderOdds := s.generateOverUnderOdds(game)
		odds = append(odds, overUnderOdds)
	}

	return odds
}

// calculateWinProbability calculates win probability based on team attributes
func (s *VirtualSportsService) calculateWinProbability(home, away *VirtualTeam) float64 {
	// Base calculation using strength and form
	homePower := home.Strength*0.7 + home.Form*0.3
	awayPower := away.Strength*0.7 + away.Form*0.3

	// Add home advantage
	homePower += 0.1

	// Calculate probability
	totalPower := homePower + awayPower
	homeProb := homePower / totalPower

	// Add some randomness
	homeProb += (s.rng.Float64() - 0.5) * 0.1

	// Ensure within bounds
	if homeProb < 0.1 {
		homeProb = 0.1
	} else if homeProb > 0.9 {
		homeProb = 0.9
	}

	return homeProb
}

// probabilityToOdds converts probability to decimal odds
func (s *VirtualSportsService) probabilityToOdds(probability float64) decimal.Decimal {
	if probability <= 0 {
		return decimal.NewFromInt(1000) // Very high odds
	}

	odds := 1.0 / probability
	return decimal.NewFromFloat(odds).Round(2)
}

// generateOverUnderOdds creates over/under betting odds
func (s *VirtualSportsService) generateOverUnderOdds(game *VirtualGame) VirtualOdds {
	// Expected goals based on team strengths
	expectedGoals := (game.HomeTeam.Strength + game.AwayTeam.Strength) * 3.0
	line := int(expectedGoals + 0.5) // Round to nearest 0.5

	// Calculate probabilities (simplified)
	overProb := 0.45 + (expectedGoals-float64(line))*0.1
	underProb := 1.0 - overProb

	// Convert to odds
	margin := 0.05
	overOdds := s.probabilityToOdds(overProb * (1 - margin))
	underOdds := s.probabilityToOdds(underProb * (1 - margin))

	return VirtualOdds{
		MarketID:   s.generateID(),
		MarketName: fmt.Sprintf("Over/Under %.1f Goals", float64(line)),
		MarketType: domain.MarketTypeTotalGoals,
		Outcomes: []VirtualOutcome{
			{
				ID:          s.generateID(),
				Name:        fmt.Sprintf("Over %.1f", float64(line)),
				Odds:        overOdds,
				Price:       decimal.NewFromInt(10),
				Status:      domain.OutcomeStatusPending,
				Probability: overProb,
			},
			{
				ID:          s.generateID(),
				Name:        fmt.Sprintf("Under %.1f", float64(line)),
				Odds:        underOdds,
				Price:       decimal.NewFromInt(10),
				Status:      domain.OutcomeStatusPending,
				Probability: underProb,
			},
		},
		LastUpdated: time.Now(),
	}
}

// SimulateGame simulates a virtual sports game
func (s *VirtualSportsService) SimulateGame(ctx context.Context, gameID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.games[gameID]
	if !exists {
		return fmt.Errorf("game not found: %s", gameID)
	}

	if game.Status != GameStatusScheduled {
		return fmt.Errorf("game is not scheduled: %s", gameID)
	}

	// Start the game
	game.Status = GameStatusLive
	game.StartTime = time.Now()
	game.CurrentScore = &GameScore{HomeScore: 0, AwayScore: 0}

	// Simulate the game based on sport
	switch game.Sport {
	case domain.SportFootball:
		s.simulateFootballGame(game)
	case domain.SportBasketball:
		s.simulateBasketballGame(game)
	case domain.SportTennis:
		s.simulateTennisGame(game)
	default:
		return fmt.Errorf("unsupported sport: %s", game.Sport)
	}

	// Finish the game
	game.Status = GameStatusFinished
	now := time.Now()
	game.EndTime = &now
	game.FinalScore = &GameScore{
		HomeScore: game.CurrentScore.HomeScore,
		AwayScore: game.CurrentScore.AwayScore,
	}

	// Update outcomes
	s.updateGameOutcomes(game)

	// Publish events
	s.publishGameEvents(game)

	// Save final state
	err := s.saveGame(ctx, game)
	if err != nil {
		log.Printf("Error saving simulated game: %v", err)
	}

	return nil
}

// simulateFootballGame simulates a football match
func (s *VirtualSportsService) simulateFootballGame(game *VirtualGame) {
	// Game events timeline
	eventTimes := []int{15, 23, 38, 45, 52, 67, 78, 85} // Minutes

	for _, minute := range eventTimes {
		// Determine if an event happens
		if s.rng.Float64() < 0.3 { // 30% chance of event
			event := s.generateFootballEvent(game, minute)
			game.Events = append(game.Events, *event)

			// Update score if it's a goal
			if event.Type == EventTypeGoal {
				if event.Team == "home" {
					game.CurrentScore.HomeScore++
				} else {
					game.CurrentScore.AwayScore++
				}
			}
		}

		// Add some time between events
		time.Sleep(100 * time.Millisecond)
	}

	// Add halftime and fulltime events
	game.Events = append(game.Events, GameEvent{
		ID:      s.generateID(),
		Type:    EventTypeHalftime,
		Time:    time.Now().Add(45 * time.Minute),
		Details: "Halftime",
		Score:   &GameScore{HomeScore: game.CurrentScore.HomeScore, AwayScore: game.CurrentScore.AwayScore},
	})

	game.Events = append(game.Events, GameEvent{
		ID:      s.generateID(),
		Type:    EventTypeFulltime,
		Time:    time.Now().Add(90 * time.Minute),
		Details: "Fulltime",
		Score:   game.FinalScore,
	})
}

// generateFootballEvent generates a random football event
func (s *VirtualSportsService) generateFootballEvent(game *VirtualGame, minute int) *GameEvent {
	eventTypes := []EventType{EventTypeGoal, EventTypeYellowCard, EventTypeRedCard, EventTypeSubstitution}
	eventType := eventTypes[s.rng.Intn(len(eventTypes))]

	team := "home"
	if s.rng.Float64() < 0.5 {
		team = "away"
	}

	teamName := game.HomeTeam.Name
	if team == "away" {
		teamName = game.AwayTeam.Name
	}

	playerName := s.generatePlayerName()

	details := ""
	switch eventType {
	case EventTypeGoal:
		details = fmt.Sprintf("GOAL! %s scores for %s", playerName, teamName)
	case EventTypeYellowCard:
		details = fmt.Sprintf("Yellow card for %s (%s)", playerName, teamName)
	case EventTypeRedCard:
		details = fmt.Sprintf("RED CARD! %s sent off (%s)", playerName, teamName)
	case EventTypeSubstitution:
		details = fmt.Sprintf("Substitution: %s off, %s on (%s)", playerName, s.generatePlayerName(), teamName)
	}

	return &GameEvent{
		ID:      s.generateID(),
		Type:    eventType,
		Time:    time.Now().Add(time.Duration(minute) * time.Minute),
		Team:    team,
		Player:  playerName,
		Details: details,
		Score:   game.CurrentScore,
	}
}

// simulateBasketballGame simulates a basketball game
func (s *VirtualSportsService) simulateBasketballGame(game *VirtualGame) {
	// Basketball has more frequent scoring
	quarters := 4
	pointsPerQuarter := 25

	for quarter := 0; quarter < quarters; quarter++ {
		for i := 0; i < pointsPerQuarter; i++ {
			// Determine who scores
			homeScores := s.rng.Float64() < s.calculateWinProbability(game.HomeTeam, game.AwayTeam)

			points := 2
			if s.rng.Float64() < 0.2 { // 20% chance of 3-pointer
				points = 3
			}

			if homeScores {
				game.CurrentScore.HomeScore += points
			} else {
				game.CurrentScore.AwayScore += points
			}

			// Add some time between scores
			time.Sleep(50 * time.Millisecond)
		}

		// Add quarter break event
		game.Events = append(game.Events, GameEvent{
			ID:      s.generateID(),
			Type:    EventTypeHalftime,
			Time:    time.Now().Add(time.Duration(quarter+1) * 12 * time.Minute),
			Details: fmt.Sprintf("End of Quarter %d", quarter+1),
			Score:   &GameScore{HomeScore: game.CurrentScore.HomeScore, AwayScore: game.CurrentScore.AwayScore},
		})
	}

	// Add final event
	game.Events = append(game.Events, GameEvent{
		ID:      s.generateID(),
		Type:    EventTypeFulltime,
		Time:    time.Now().Add(48 * time.Minute),
		Details: "Fulltime",
		Score:   game.FinalScore,
	})
}

// simulateTennisGame simulates a tennis match
func (s *VirtualSportsService) simulateTennisGame(game *VirtualGame) {
	// Simplified tennis simulation - best of 3 sets
	sets := []int{0, 0}
	maxSets := 3

	for len(sets) < maxSets && (sets[0] < 2 && sets[1] < 2) {
		// Simulate a set
		setWinner := s.rng.Float64() < s.calculateWinProbability(game.HomeTeam, game.AwayTeam)

		if setWinner {
			sets[0]++
		} else {
			sets[1]++
		}

		// Generate games in the set (simplified)
		games := 6
		if setWinner {
			game.CurrentScore.HomeScore += games
		} else {
			game.CurrentScore.AwayScore += games
		}

		// Add set event
		game.Events = append(game.Events, GameEvent{
			ID:      s.generateID(),
			Type:    EventTypeHalftime,
			Time:    time.Now().Add(time.Duration(len(sets)) * 30 * time.Minute),
			Details: fmt.Sprintf("Set %d completed", len(sets)),
			Score:   &GameScore{HomeScore: sets[0], AwayScore: sets[1]},
		})

		time.Sleep(200 * time.Millisecond)
	}

	// Add final event
	game.Events = append(game.Events, GameEvent{
		ID:      s.generateID(),
		Type:    EventTypeFulltime,
		Time:    time.Now().Add(120 * time.Minute),
		Details: "Match completed",
		Score:   game.FinalScore,
	})
}

// generatePlayerName generates a random player name
func (s *VirtualSportsService) generatePlayerName() string {
	firstNames := []string{"John", "Mike", "David", "James", "Robert", "Michael", "William", "Richard"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis"}

	return fmt.Sprintf("%s %s", firstNames[s.rng.Intn(len(firstNames))], lastNames[s.rng.Intn(len(lastNames))])
}

// updateGameOutcomes updates betting outcomes based on game result
func (s *VirtualSportsService) updateGameOutcomes(game *VirtualGame) {
	for i, odds := range game.Odds {
		for j, outcome := range odds.Outcomes {
			// Determine if outcome won
			won := false

			switch odds.MarketType {
			case domain.MarketTypeMatchWinner:
				if outcome.Name == game.HomeTeam.Name && game.FinalScore.HomeScore > game.FinalScore.AwayScore {
					won = true
				} else if outcome.Name == game.AwayTeam.Name && game.FinalScore.AwayScore > game.FinalScore.HomeScore {
					won = true
				} else if outcome.Name == "Draw" && game.FinalScore.HomeScore == game.FinalScore.AwayScore {
					won = true
				}
			case domain.MarketTypeTotalGoals:
				totalGoals := game.FinalScore.HomeScore + game.FinalScore.AwayScore
				if outcome.Name[:4] == "Over" && totalGoals > 2 {
					won = true
				} else if outcome.Name[:4] == "Under" && totalGoals <= 2 {
					won = true
				}
			}

			if won {
				game.Odds[i].Outcomes[j].Status = domain.OutcomeStatusWon
			} else {
				game.Odds[i].Outcomes[j].Status = domain.OutcomeStatusLost
			}
		}
	}
}

// publishGameEvents publishes game events to the event bus
func (s *VirtualSportsService) publishGameEvents(game *VirtualGame) {
	for _, event := range game.Events {
		s.publishEvent("virtual.game.event", map[string]interface{}{
			"game_id":    game.ID,
			"event_id":   event.ID,
			"event_type": event.Type,
			"time":       event.Time,
			"details":    event.Details,
			"score":      event.Score,
		})
	}

	// Publish game completion
	s.publishEvent("virtual.game.completed", map[string]interface{}{
		"game_id":      game.ID,
		"sport":        game.Sport,
		"final_score":  game.FinalScore,
		"events_count": len(game.Events),
	})
}

// GetActiveGames returns all currently active virtual games
func (s *VirtualSportsService) GetActiveGames(ctx context.Context) ([]*VirtualGame, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var activeGames []*VirtualGame
	for _, game := range s.games {
		if game.Status == GameStatusLive || game.Status == GameStatusScheduled {
			activeGames = append(activeGames, game)
		}
	}

	return activeGames, nil
}

// GetGameByID returns a specific virtual game
func (s *VirtualSportsService) GetGameByID(ctx context.Context, gameID string) (*VirtualGame, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	game, exists := s.games[gameID]
	if !exists {
		return nil, fmt.Errorf("game not found: %s", gameID)
	}

	return game, nil
}

// saveGame saves a virtual game to the repository
func (s *VirtualSportsService) saveGame(ctx context.Context, game *VirtualGame) error {
	// Convert to domain match for storage
	var score *domain.MatchScore
	if game.CurrentScore != nil {
		score = &domain.MatchScore{
			HomeScore: game.CurrentScore.HomeScore,
			AwayScore: game.CurrentScore.AwayScore,
		}
	}

	match := &domain.Match{
		ID:          game.ID,
		Sport:       game.Sport,
		League:      "Virtual Sports",
		HomeTeam:    game.HomeTeam.Name,
		AwayTeam:    game.AwayTeam.Name,
		StartTime:   game.StartTime,
		Status:      domain.MatchStatusLive,
		Score:       score,
		CountryCode: "VS", // Virtual Sports
	}

	if game.Status == GameStatusFinished {
		match.Status = domain.MatchStatusCompleted
		if game.FinalScore != nil {
			match.Score = &domain.MatchScore{
				HomeScore: game.FinalScore.HomeScore,
				AwayScore: game.FinalScore.AwayScore,
			}
		}
	}

	// Save to match repository
	err := s.matchRepo.Create(ctx, match)
	if err != nil {
		return fmt.Errorf("failed to save match: %w", err)
	}

	return nil
}

// generateID generates a unique ID
func (s *VirtualSportsService) generateID() string {
	return fmt.Sprintf("vs_%d_%d", time.Now().Unix(), s.rng.Intn(10000))
}

// publishEvent publishes an event to the event bus
func (s *VirtualSportsService) publishEvent(topic string, data interface{}) {
	if s.eventBus != nil {
		err := s.eventBus.Publish(topic, data)
		if err != nil {
			log.Printf("Error publishing virtual sports event %s: %v", topic, err)
		}
	}
}
