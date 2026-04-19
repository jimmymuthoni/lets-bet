package jackpots

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// Repository interfaces to avoid import cycle
type JackpotRepository interface {
	Create(ctx context.Context, jackpot *Jackpot) error
	GetByID(ctx context.Context, id string) (*Jackpot, error)
	Update(ctx context.Context, jackpot *Jackpot) error
	GetActive(ctx context.Context) ([]*Jackpot, error)
	CreateTicket(ctx context.Context, ticket *JackpotTicket) error
	DeleteTicket(ctx context.Context, ticketID string) error
	GetActiveTickets(ctx context.Context, jackpotID string) ([]*JackpotTicket, error)
	GetUserTickets(ctx context.Context, userID string) ([]*JackpotTicket, error)
	UpdateTicketStatus(ctx context.Context, ticketID string, status TicketStatus, prize decimal.Decimal) error
}

type SportBetRepository interface {
	Create(ctx context.Context, bet interface{}) error
}

// WalletService interface for wallet operations
type WalletService interface {
	Credit(ctx context.Context, userID string, amount decimal.Decimal, movement Movement) (*Transaction, error)
	Debit(ctx context.Context, userID string, amount decimal.Decimal, movement Movement) (*Transaction, error)
}

// EventBus interface for publishing events
type EventBus interface {
	Publish(topic string, data interface{}) error
}

// Movement represents a wallet movement
type Movement struct {
	ID          string          `json:"id"`
	UserID      string          `json:"user_id"`
	Amount      decimal.Decimal `json:"amount"`
	Type        string          `json:"type"`
	Status      string          `json:"status"`
	Reference   string          `json:"reference"`
	CreatedAt   time.Time       `json:"created_at"`
	CompletedAt time.Time       `json:"completed_at,omitempty"`
}

// Transaction represents a wallet transaction
type Transaction struct {
	ID          string          `json:"id"`
	UserID      string          `json:"user_id"`
	Amount      decimal.Decimal `json:"amount"`
	Type        string          `json:"type"`
	Status      string          `json:"status"`
	Reference   string          `json:"reference"`
	MovementID  string          `json:"movement_id"`
	CreatedAt   time.Time       `json:"created_at"`
	CompletedAt time.Time       `json:"completed_at,omitempty"`
}

// Jackpot represents a jackpot game
type Jackpot struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Type             JackpotType     `json:"type"`
	CurrentAmount    decimal.Decimal `json:"current_amount"`
	SeedAmount       decimal.Decimal `json:"seed_amount"`
	MinBet           decimal.Decimal `json:"min_bet"`
	MaxBet           decimal.Decimal `json:"max_bet"`
	ContributionRate decimal.Decimal `json:"contribution_rate"`
	Status           JackpotStatus   `json:"status"`
	StartTime        time.Time       `json:"start_time"`
	EndTime          time.Time       `json:"end_time,omitempty"`
	NextDrawTime     time.Time       `json:"next_draw_time"`
	WinningTicketID  string          `json:"winning_ticket_id,omitempty"`
	WinningUserID    string          `json:"winning_user_id,omitempty"`
	WinningAmount    decimal.Decimal `json:"winning_amount,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// JackpotType represents different types of jackpots
type JackpotType string

const (
	JackpotTypeProgressive JackpotType = "PROGRESSIVE"
	JackpotTypeFixed       JackpotType = "FIXED"
	JackpotTypeDaily       JackpotType = "DAILY"
	JackpotTypeWeekly      JackpotType = "WEEKLY"
	JackpotTypeMystery     JackpotType = "MYSTERY"
)

// JackpotStatus represents jackpot status
type JackpotStatus string

const (
	JackpotStatusActive    JackpotStatus = "ACTIVE"
	JackpotStatusPaused    JackpotStatus = "PAUSED"
	JackpotStatusCompleted JackpotStatus = "COMPLETED"
	JackpotStatusCancelled JackpotStatus = "CANCELLED"
)

// JackpotTicket represents a jackpot ticket
type JackpotTicket struct {
	ID        string          `json:"id"`
	JackpotID string          `json:"jackpot_id"`
	UserID    string          `json:"user_id"`
	BetAmount decimal.Decimal `json:"bet_amount"`
	Numbers   []int           `json:"numbers"`
	Status    TicketStatus    `json:"status"`
	Prize     decimal.Decimal `json:"prize"`
	DrawTime  time.Time       `json:"draw_time"`
	Won       bool            `json:"won"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// TicketStatus represents ticket status
type TicketStatus string

const (
	TicketStatusActive    TicketStatus = "ACTIVE"
	TicketStatusWinner    TicketStatus = "WINNER"
	TicketStatusLoser     TicketStatus = "LOSER"
	TicketStatusExpired   TicketStatus = "EXPIRED"
	TicketStatusCancelled TicketStatus = "CANCELLED"
)

// JackpotConfig represents jackpot configuration
type JackpotConfig struct {
	MaxActiveJackpots       int             `json:"max_active_jackpots"`
	DefaultSeedAmount       decimal.Decimal `json:"default_seed_amount"`
	DefaultContributionRate decimal.Decimal `json:"default_contribution_rate"`
	MinContributionAmount   decimal.Decimal `json:"min_contribution_amount"`
	MaxTicketsPerUser       int             `json:"max_tickets_per_user"`
	DrawInterval            time.Duration   `json:"draw_interval"`
	AutoDraw                bool            `json:"auto_draw"`
	WinNotificationDelay    time.Duration   `json:"win_notification_delay"`
}

// JackpotMetrics represents jackpot metrics
type JackpotMetrics struct {
	TotalJackpots      int64           `json:"total_jackpots"`
	ActiveJackpots     int64           `json:"active_jackpots"`
	TotalTickets       int64           `json:"total_tickets"`
	ActiveTickets      int64           `json:"active_tickets"`
	TotalPayouts       decimal.Decimal `json:"total_payouts"`
	TotalContributions decimal.Decimal `json:"total_contributions"`
	AverageJackpotSize decimal.Decimal `json:"average_jackpot_size"`
	LargestJackpot     decimal.Decimal `json:"largest_jackpot"`
	LastDrawTime       time.Time       `json:"last_draw_time"`
	NextDrawTime       time.Time       `json:"next_draw_time"`
}

// JackpotDraw represents a jackpot draw result
type JackpotDraw struct {
	ID             string          `json:"id"`
	JackpotID      string          `json:"jackpot_id"`
	DrawTime       time.Time       `json:"draw_time"`
	WinningNumbers []int           `json:"winning_numbers"`
	TotalTickets   int64           `json:"total_tickets"`
	WinningTicket  *JackpotTicket  `json:"winning_ticket,omitempty"`
	PrizeAmount    decimal.Decimal `json:"prize_amount"`
	Status         DrawStatus      `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
}

// DrawStatus represents draw status
type DrawStatus string

const (
	DrawStatusPending   DrawStatus = "PENDING"
	DrawStatusCompleted DrawStatus = "COMPLETED"
	DrawStatusFailed    DrawStatus = "FAILED"
)
