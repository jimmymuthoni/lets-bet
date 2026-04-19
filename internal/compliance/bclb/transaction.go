package compliance

import (
	"context"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

// isSuspiciousTransaction checks for suspicious transaction patterns
func (s *BCLBService) isSuspiciousTransaction(userID string, amount decimal.Decimal, transactionType string) bool {
	// Check for suspicious transaction patterns using database
	isSuspicious, err := s.transactionRepo.IsSuspiciousTransaction(context.Background(), userID, amount, transactionType)
	if err != nil {
		log.Printf("Error checking suspicious transaction for user %s: %v", userID, err)
		return false
	}

	return isSuspicious
}

// exceedsTransactionFrequency checks transaction frequency limits
func (s *BCLBService) exceedsTransactionFrequency(userID string) bool {
	// Check transaction frequency limits using database
	// Define frequency limit: max 100 transactions per hour
	const maxTransactionsPerHour = 100
	const oneHour = time.Hour

	exceeds, err := s.transactionRepo.ExceedsTransactionFrequency(context.Background(), userID, maxTransactionsPerHour, oneHour)
	if err != nil {
		log.Printf("Error checking transaction frequency for user %s: %v", userID, err)
		return false
	}

	return exceeds
}

// ValidateLargeTransaction validates large transactions for AML compliance
func (s *BCLBService) ValidateLargeTransaction(ctx context.Context, userID string, amount decimal.Decimal, transactionType string) (*ComplianceCheck, error) {
	check := &ComplianceCheck{
		CheckTime:     time.Now(),
		CheckType:     "LARGE_TRANSACTION",
		UserID:        userID,
		TransactionID: generateTransactionID(),
	}

	// Define large transaction threshold (e.g., $10,000)
	largeTransactionThreshold := decimal.NewFromInt(10000)

	if amount.GreaterThanOrEqual(largeTransactionThreshold) {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "LARGE_TRANSACTION",
			Description: "Transaction amount exceeds reporting threshold",
			Severity:    ViolationSeverityHigh,
			Amount:      amount,
			Limit:       largeTransactionThreshold,
			Action:      "FLAG_FOR_MANUAL_REVIEW",
		})

		// Additional checks for large transactions
		if !s.isKYCCompliant(userID) {
			check.Violations = append(check.Violations, ComplianceViolation{
				Type:        "LARGE_TRANSACTION_NO_KYC",
				Description: "Large transaction from user without KYC verification",
				Severity:    ViolationSeverityCritical,
				Amount:      amount,
				Action:      "BLOCK_TRANSACTION",
			})
		}

		// Check for rapid large transactions
		if s.hasRecentLargeTransactions(userID, amount) {
			check.Violations = append(check.Violations, ComplianceViolation{
				Type:        "RAPID_LARGE_TRANSACTIONS",
				Description: "Multiple large transactions in short period",
				Severity:    ViolationSeverityCritical,
				Amount:      amount,
				Action:      "FLAG_FOR_INVESTIGATION",
			})
		}
	}

	check.Passed = len(check.Violations) == 0

	// Log compliance check
	s.logComplianceCheck(check)

	// Publish compliance event
	s.publishComplianceEvent("compliance.large_transaction_validated", check)

	return check, nil
}

// hasRecentLargeTransactions checks for recent large transactions from the same user
func (s *BCLBService) hasRecentLargeTransactions(userID string, amount decimal.Decimal) bool {
	// In a real implementation, this would query the database for recent transactions
	// For now, return false as a placeholder
	_ = userID // Suppress unused parameter warning
	_ = amount // Suppress unused parameter warning
	return false
}

// ValidateHighRiskTransaction validates high-risk transaction types
func (s *BCLBService) ValidateHighRiskTransaction(ctx context.Context, userID string, amount decimal.Decimal, transactionType string) (*ComplianceCheck, error) {
	check := &ComplianceCheck{
		CheckTime:     time.Now(),
		CheckType:     "HIGH_RISK_TRANSACTION",
		UserID:        userID,
		TransactionID: generateTransactionID(),
	}

	// Define high-risk transaction types
	highRiskTypes := map[string]bool{
		"international_transfer": true,
		"third_party_payment":    true,
		"crypto_exchange":        true,
		"anonymous_payment":      true,
	}

	if highRiskTypes[transactionType] {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "HIGH_RISK_TRANSACTION_TYPE",
			Description: "Transaction type identified as high risk",
			Severity:    ViolationSeverityHigh,
			Amount:      amount,
			Action:      "ENHANCED_DUE_DILIGENCE",
		})

		// Require enhanced verification for high-risk transactions
		if !s.isKYCCompliant(userID) {
			check.Violations = append(check.Violations, ComplianceViolation{
				Type:        "HIGH_RISK_NO_KYC",
				Description: "High-risk transaction from user without KYC verification",
				Severity:    ViolationSeverityCritical,
				Amount:      amount,
				Action:      "BLOCK_TRANSACTION",
			})
		}
	}

	check.Passed = len(check.Violations) == 0

	// Log compliance check
	s.logComplianceCheck(check)

	// Publish compliance event
	s.publishComplianceEvent("compliance.high_risk_transaction_validated", check)

	return check, nil
}
