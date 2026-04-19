package compliance

import (
	"time"

	"github.com/shopspring/decimal"
)

// EventBus interface for publishing events
type EventBus interface {
	Publish(topic string, data interface{}) error
}

// BCLBConfig represents BCLB compliance configuration
type BCLBConfig struct {
	LicenseNumber            string          `json:"license_number"`
	LicenseExpiry            time.Time       `json:"license_expiry"`
	MaxDailyStake            decimal.Decimal `json:"max_daily_stake"`
	MaxWeeklyStake           decimal.Decimal `json:"max_weekly_stake"`
	MaxMonthlyStake          decimal.Decimal `json:"max_monthly_stake"`
	MinAge                   int             `json:"min_age"`
	MaxBetPerEvent           decimal.Decimal `json:"max_bet_per_event"`
	MaxAccumulatorBets       int             `json:"max_accumulator_bets"`
	RequiredKYCLevel         string          `json:"required_kyc_level"`
	SelfExclusionMinDays     int             `json:"self_exclusion_min_days"`
	ResponsibleGamingEnabled bool            `json:"responsible_gaming_enabled"`
}

// ViolationSeverity represents the severity of a compliance violation
type ViolationSeverity string

const (
	ViolationSeverityLow      ViolationSeverity = "LOW"
	ViolationSeverityMedium   ViolationSeverity = "MEDIUM"
	ViolationSeverityHigh     ViolationSeverity = "HIGH"
	ViolationSeverityCritical ViolationSeverity = "CRITICAL"
)

// ComplianceCheck represents a compliance validation result
type ComplianceCheck struct {
	Passed        bool                  `json:"passed"`
	Violations    []ComplianceViolation `json:"violations"`
	Warnings      []string              `json:"warnings"`
	CheckTime     time.Time             `json:"check_time"`
	CheckType     string                `json:"check_type"`
	UserID        string                `json:"user_id,omitempty"`
	TransactionID string                `json:"transaction_id,omitempty"`
	BetID         string                `json:"bet_id,omitempty"`
}

// ComplianceViolation represents a specific compliance violation
type ComplianceViolation struct {
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Severity    ViolationSeverity `json:"severity"`
	Amount      decimal.Decimal   `json:"amount,omitzero"`
	Limit       decimal.Decimal   `json:"limit,omitzero"`
	Action      string            `json:"action"`
}

// UserLimits represents user-specific betting limits
type UserLimits struct {
	UserID           string          `json:"user_id"`
	DailyLimit       decimal.Decimal `json:"daily_limit"`
	WeeklyLimit      decimal.Decimal `json:"weekly_limit"`
	MonthlyLimit     decimal.Decimal `json:"monthly_limit"`
	MaxBetSize       decimal.Decimal `json:"max_bet_size"`
	AccumulatorLimit int             `json:"accumulator_limit"`
	CoolingOffPeriod time.Duration   `json:"cooling_off_period"`
	SelfExcluded     bool            `json:"self_excluded"`
	SelfExclusionEnd *time.Time      `json:"self_exclusion_end,omitempty"`
	LastUpdated      time.Time       `json:"last_updated"`
}

// UserComplianceStats represents user compliance statistics
type UserComplianceStats struct {
	TotalUsers         int64   `json:"total_users"`
	VerifiedUsers      int64   `json:"verified_users"`
	AgeVerifiedUsers   int64   `json:"age_verified_users"`
	SelfExcludedUsers  int64   `json:"self_excluded_users"`
	LimitBreachedUsers int64   `json:"limit_breached_users"`
	SuspiciousUsers    int64   `json:"suspicious_users"`
	ComplianceRate     float64 `json:"compliance_rate"`
}

// FinancialComplianceStats represents financial compliance statistics
type FinancialComplianceStats struct {
	TotalStake             decimal.Decimal `json:"total_stake"`
	TotalWinnings          decimal.Decimal `json:"total_winnings"`
	TotalDeposits          decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals       decimal.Decimal `json:"total_withdrawals"`
	TaxCollected           decimal.Decimal `json:"tax_collected"`
	AMLFlaggedTransactions int64           `json:"aml_flagged_transactions"`
	LargeTransactions      int64           `json:"large_transactions"`
}

// ComplianceReport represents a comprehensive compliance report
type ComplianceReport struct {
	Period          string                   `json:"period"`
	TotalChecks     int64                    `json:"total_checks"`
	PassedChecks    int64                    `json:"passed_checks"`
	FailedChecks    int64                    `json:"failed_checks"`
	Violations      []ComplianceViolation    `json:"violations"`
	UserStatistics  UserComplianceStats      `json:"user_statistics"`
	FinancialStats  FinancialComplianceStats `json:"financial_stats"`
	Recommendations []string                 `json:"recommendations"`
	GeneratedAt     time.Time                `json:"generated_at"`
}
