package compliance

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// UserRepository interface for user data access
type UserRepository interface {
	GetUser(ctx context.Context, userID string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	GetUserAgeVerification(ctx context.Context, userID string) (*AgeVerification, error)
	GetUserKYCStatus(ctx context.Context, userID string) (*KYCStatus, error)
	GetUserSelfExclusion(ctx context.Context, userID string) (*SelfExclusion, error)
	UpdateSelfExclusion(ctx context.Context, userID string, endDate time.Time, reason string) error
}

// BetRepository interface for betting data access
type BetRepository interface {
	GetUserDailyStake(ctx context.Context, userID string, date time.Time) (decimal.Decimal, error)
	GetUserWeeklyStake(ctx context.Context, userID string, weekStart time.Time) (decimal.Decimal, error)
	GetUserMonthlyStake(ctx context.Context, userID string, monthStart time.Time) (decimal.Decimal, error)
	GetUserBettingHistory(ctx context.Context, userID string, limit int) ([]*Bet, error)
}

// TransactionRepository interface for transaction data access
type TransactionRepository interface {
	GetUserTransactionCount(ctx context.Context, userID string, period time.Duration) (int64, error)
	GetUserTransactionHistory(ctx context.Context, userID string, limit int) ([]*Transaction, error)
	IsSuspiciousTransaction(ctx context.Context, userID string, amount decimal.Decimal, transactionType string) (bool, error)
	ExceedsTransactionFrequency(ctx context.Context, userID string, limit int, period time.Duration) (bool, error)
}

// CoolingOffRepository interface for cooling off data access
type CoolingOffRepository interface {
	IsUserInCoolingOffPeriod(ctx context.Context, userID string) (bool, error)
	GetUserCoolingOff(ctx context.Context, userID string) (*CoolingOff, error)
	SetCoolingOff(ctx context.Context, userID string, duration time.Duration, reason string) error
}

// User represents a user entity
type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AgeVerification represents age verification status
type AgeVerification struct {
	UserID     string    `json:"user_id"`
	Verified   bool      `json:"verified"`
	Method     string    `json:"method"`
	VerifiedAt time.Time `json:"verified_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	DocumentID string    `json:"document_id"`
}

// KYCStatus represents KYC verification status
type KYCStatus struct {
	UserID      string    `json:"user_id"`
	Status      string    `json:"status"` // PENDING, VERIFIED, REJECTED
	Level       string    `json:"level"`  // BASIC, STANDARD, ENHANCED
	VerifiedAt  time.Time `json:"verified_at"`
	Documents   []string  `json:"documents"`
	LastUpdated time.Time `json:"last_updated"`
}

// SelfExclusion represents self-exclusion status
type SelfExclusion struct {
	UserID     string    `json:"user_id"`
	Active     bool      `json:"active"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Duration   string    `json:"duration"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

// Bet represents a betting transaction
type Bet struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Amount    decimal.Decimal `json:"amount"`
	Status    string          `json:"status"`
	PlacedAt  time.Time       `json:"placed_at"`
	EventType string          `json:"event_type"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Amount    decimal.Decimal `json:"amount"`
	Type      string          `json:"type"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	IPAddress string          `json:"ip_address"`
	UserAgent string          `json:"user_agent"`
	RiskScore int             `json:"risk_score"`
}

// CoolingOff represents cooling off period
type CoolingOff struct {
	UserID     string        `json:"user_id"`
	Active     bool          `json:"active"`
	StartDate  time.Time     `json:"start_date"`
	EndDate    time.Time     `json:"end_date"`
	Duration   time.Duration `json:"duration"`
	Reason     string        `json:"reason"`
	CreatedAt  time.Time     `json:"created_at"`
	ModifiedAt time.Time     `json:"modified_at"`
}
