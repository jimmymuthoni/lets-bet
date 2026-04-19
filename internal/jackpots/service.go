package jackpots

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// JackpotService manages jackpot games and payouts
type JackpotService struct {
	jackpotRepo   JackpotRepository
	betRepo       SportBetRepository
	walletService WalletService
	eventBus      EventBus
	rng           *rand.Rand
	mu            sync.RWMutex
}

// NewJackpotService creates a new jackpot service
func NewJackpotService(
	jackpotRepo JackpotRepository,
	betRepo SportBetRepository,
	walletService WalletService,
	eventBus EventBus,
) *JackpotService {
	return &JackpotService{
		jackpotRepo:   jackpotRepo,
		betRepo:       betRepo,
		walletService: walletService,
		eventBus:      eventBus,
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CreateJackpot creates a new jackpot game
func (s *JackpotService) CreateJackpot(ctx context.Context, jackpot *Jackpot) error {
	// Validate jackpot configuration
	if err := validateJackpot(jackpot); err != nil {
		return fmt.Errorf("invalid jackpot configuration: %w", err)
	}

	// Set initial values
	jackpot.ID = generateID()
	jackpot.Status = JackpotStatusActive
	jackpot.CreatedAt = time.Now()
	jackpot.UpdatedAt = time.Now()

	// Set next draw time based on jackpot type
	switch jackpot.Type {
	case JackpotTypeDaily:
		jackpot.NextDrawTime = time.Now().Add(24 * time.Hour)
	case JackpotTypeWeekly:
		jackpot.NextDrawTime = time.Now().Add(7 * 24 * time.Hour)
	default:
		jackpot.NextDrawTime = time.Now().Add(24 * time.Hour)
	}

	// Create jackpot
	if err := s.jackpotRepo.Create(ctx, jackpot); err != nil {
		return fmt.Errorf("failed to create jackpot: %w", err)
	}

	// Publish jackpot creation event
	s.publishJackpotEvent("jackpot.created", jackpot)

	log.Printf("Created jackpot: %s", jackpot.ID)
	return nil
}

// GetActiveJackpots returns all active jackpots
func (s *JackpotService) GetActiveJackpots(ctx context.Context) ([]*Jackpot, error) {
	jackpots, err := s.jackpotRepo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active jackpots: %w", err)
	}

	// Publish jackpots access event
	s.publishJackpotEvent("jackpot.active_accessed", jackpots)

	return jackpots, nil
}

// PurchaseTicket purchases a jackpot ticket
func (s *JackpotService) PurchaseTicket(ctx context.Context, jackpotID, userID string, betAmount decimal.Decimal, numbers []int) (*JackpotTicket, error) {
	// Get jackpot
	jackpot, err := s.jackpotRepo.GetByID(ctx, jackpotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get jackpot: %w", err)
	}

	// Validate ticket purchase
	if err := validateTicketPurchase(jackpot, betAmount, numbers); err != nil {
		return nil, fmt.Errorf("invalid ticket purchase: %w", err)
	}

	// Create ticket
	ticket := &JackpotTicket{
		ID:        generateID(),
		JackpotID: jackpotID,
		UserID:    userID,
		BetAmount: betAmount,
		Numbers:   numbers,
		Status:    TicketStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save ticket
	if err := s.jackpotRepo.CreateTicket(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	// Update jackpot amount
	jackpot.CurrentAmount = jackpot.CurrentAmount.Add(betAmount.Mul(jackpot.ContributionRate))
	jackpot.UpdatedAt = time.Now()
	if err := s.jackpotRepo.Update(ctx, jackpot); err != nil {
		return nil, fmt.Errorf("failed to update jackpot: %w", err)
	}

	// Publish ticket purchase event
	s.publishJackpotEvent("jackpot.ticket.purchased", ticket)

	log.Printf("Purchased ticket: %s for jackpot: %s", ticket.ID, jackpotID)
	return ticket, nil
}

// DrawJackpot draws a jackpot and selects a winner
func (s *JackpotService) DrawJackpot(ctx context.Context, jackpotID string) (*JackpotDraw, error) {
	// Get jackpot
	jackpot, err := s.jackpotRepo.GetByID(ctx, jackpotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get jackpot: %w", err)
	}

	// Get active tickets
	tickets, err := s.jackpotRepo.GetActiveTickets(ctx, jackpotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active tickets: %w", err)
	}

	if len(tickets) == 0 {
		return nil, fmt.Errorf("no active tickets for jackpot")
	}

	// Draw winning numbers
	winningNumbers := generateWinningNumbers(s.rng, 6)

	// Find winner
	winner := findWinner(tickets, winningNumbers)

	// Create draw result
	draw := &JackpotDraw{
		ID:             generateID(),
		JackpotID:      jackpotID,
		DrawTime:       time.Now(),
		WinningNumbers: winningNumbers,
		TotalTickets:   int64(len(tickets)),
		PrizeAmount:    jackpot.CurrentAmount,
		Status:         DrawStatusCompleted,
		CreatedAt:      time.Now(),
	}

	if winner != nil {
		draw.WinningTicket = winner
		
		// Update ticket status
		winner.Status = TicketStatusWinner
		winner.Prize = jackpot.CurrentAmount
		winner.DrawTime = time.Now()
		winner.UpdatedAt = time.Now()
		
		if err := s.jackpotRepo.UpdateTicketStatus(ctx, winner.ID, TicketStatusWinner, jackpot.CurrentAmount); err != nil {
			return nil, fmt.Errorf("failed to update winning ticket: %w", err)
		}

		// Credit winner's wallet
		movement := Movement{
			ID:          generateID(),
			UserID:      winner.UserID,
			Amount:      jackpot.CurrentAmount,
			Type:        "WIN",
			Status:      "COMPLETED",
			Reference:   jackpotID,
			CreatedAt:   time.Now(),
			CompletedAt: time.Now(),
		}

		_, err = s.walletService.Credit(ctx, winner.UserID, jackpot.CurrentAmount, movement)
		if err != nil {
			return nil, fmt.Errorf("failed to credit winner: %w", err)
		}

		// Update jackpot
		jackpot.Status = JackpotStatusCompleted
		jackpot.WinningTicketID = winner.ID
		jackpot.WinningUserID = winner.UserID
		jackpot.WinningAmount = jackpot.CurrentAmount
		jackpot.EndTime = time.Now()
		jackpot.UpdatedAt = time.Now()
	} else {
		// No winner, roll over
		jackpot.NextDrawTime = time.Now().Add(24 * time.Hour)
		jackpot.UpdatedAt = time.Now()
	}

	if err := s.jackpotRepo.Update(ctx, jackpot); err != nil {
		return nil, fmt.Errorf("failed to update jackpot: %w", err)
	}

	// Publish draw event
	s.publishJackpotEvent("jackpot.draw.completed", draw)

	log.Printf("Drew jackpot: %s with %d tickets", jackpotID, len(tickets))
	return draw, nil
}

// GetJackpotMetrics returns jackpot metrics
func (s *JackpotService) GetJackpotMetrics(ctx context.Context) (*JackpotMetrics, error) {
	// In a real implementation, these would be calculated from actual data
	return &JackpotMetrics{
		TotalJackpots:      15,
		ActiveJackpots:     8,
		TotalTickets:        50000,
		ActiveTickets:       12000,
		TotalPayouts:        decimal.NewFromInt(2500000),
		TotalContributions:  decimal.NewFromInt(3000000),
		AverageJackpotSize: decimal.NewFromInt(125000),
		LargestJackpot:      decimal.NewFromInt(500000),
		LastDrawTime:        time.Now().Add(-2 * time.Hour),
		NextDrawTime:        time.Now().Add(4 * time.Hour),
	}, nil
}

// publishJackpotEvent publishes jackpot events
func (s *JackpotService) publishJackpotEvent(topic string, data interface{}) {
	if s.eventBus != nil {
		err := s.eventBus.Publish(topic, data)
		if err != nil {
			log.Printf("Error publishing jackpot event %s: %v", topic, err)
		}
	}
}
