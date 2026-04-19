package security

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

// generateID generates a unique ID for responsible gaming records
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// _calculateEndDate calculates end date based on duration string
func _calculateEndDate(duration string) time.Time {
	now := time.Now()

	switch duration {
	case "6 months":
		return now.AddDate(0, 6, 0)
	case "1 year":
		return now.AddDate(1, 0, 0)
	case "2 years":
		return now.AddDate(2, 0, 0)
	case "5 years":
		return now.AddDate(5, 0, 0)
	case "permanent":
		return now.AddDate(100, 0, 0) // Far future date
	default:
		return now.AddDate(0, 6, 0) // Default to 6 months
	}
}

// _calculateComplianceScore calculates overall compliance score
func _calculateComplianceScore(assessment *ResponsibleGaming) float64 {
	score := 100.0

	// Deduct points for violations
	for _, violation := range assessment.Violations {
		switch violation.Severity {
		case SeverityCritical:
			score -= 25
		case SeverityHigh:
			score -= 15
		case SeverityMedium:
			score -= 10
		case SeverityLow:
			score -= 5
		case SeverityInfo:
			score -= 2
		}
	}

	// Deduct points for missing interventions
	if len(assessment.Interventions) == 0 {
		score -= 10
	}

	// Add points for educational materials
	if len(assessment.Education) > 0 {
		score += 5
	}

	if score < 0 {
		score = 0
	}

	return score
}

// _calculateUserComplianceScore calculates compliance score for a specific user
func _calculateUserComplianceScore(assessment *ResponsibleGaming) float64 {
	score := 100.0

	// Check if user has active self-exclusion
	for _, se := range assessment.SelfExclusion {
		if se.Status == "Active" {
			score += 20 // Bonus for responsible self-exclusion
		}
	}

	// Check if user has set limits
	if len(assessment.DepositLimits) > 0 {
		score += 10
	}
	if len(assessment.BettingLimits) > 0 {
		score += 10
	}
	if len(assessment.TimeLimits) > 0 {
		score += 10
	}

	// Deduct points for violations
	for _, violation := range assessment.Violations {
		switch violation.Severity {
		case SeverityCritical:
			score -= 30
		case SeverityHigh:
			score -= 20
		case SeverityMedium:
			score -= 15
		case SeverityLow:
			score -= 10
		case SeverityInfo:
			score -= 5
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

// _generateRGRecommendations generates responsible gaming recommendations
func _generateRGRecommendations(assessment *ResponsibleGaming) []string {
	var recommendations []string

	// Add recommendations based on violations
	for _, violation := range assessment.Violations {
		switch violation.Type {
		case "Underage Betting":
			recommendations = append(recommendations, "Strengthen age verification processes")
		case "Excessive Betting":
			recommendations = append(recommendations, "Implement more aggressive betting limits")
		case "Problem Gambling Indicators":
			recommendations = append(recommendations, "Enhance automated detection of problem gambling")
		}
	}

	// Add general recommendations
	recommendations = append(recommendations, "Increase awareness of responsible gaming tools")
	recommendations = append(recommendations, "Improve staff training on responsible gaming")
	recommendations = append(recommendations, "Enhance educational materials for users")

	return recommendations
}

// _generateUserRecommendations generates user-specific recommendations
func _generateUserRecommendations(assessment *ResponsibleGaming, userID string) []string {
	var recommendations []string

	// Check if user needs to set limits
	if len(assessment.DepositLimits) == 0 {
		recommendations = append(recommendations, "Consider setting deposit limits")
	}
	if len(assessment.BettingLimits) == 0 {
		recommendations = append(recommendations, "Consider setting betting limits")
	}
	if len(assessment.TimeLimits) == 0 {
		recommendations = append(recommendations, "Consider setting time limits")
	}

	// Check for recent violations
	for _, violation := range assessment.Violations {
		if violation.Date.After(time.Now().AddDate(0, -1, 0)) { // Last month
			switch violation.Type {
			case "Excessive Betting":
				recommendations = append(recommendations, "Take a break from betting")
				recommendations = append(recommendations, "Consider self-exclusion options")
			case "Problem Gambling Indicators":
				recommendations = append(recommendations, "Seek help from gambling support services")
				recommendations = append(recommendations, "Contact our responsible gaming team")
			}
		}
	}

	return recommendations
}

// _getUserSelfExclusion gets user's self-exclusion records
func _getUserSelfExclusion(ctx context.Context, userID string) []SelfExclusionRecord {
	// In a real implementation, this would query the database
	// For now, return empty slice
	return []SelfExclusionRecord{}
}

// _getUserDepositLimits gets user's deposit limits
func _getUserDepositLimits(ctx context.Context, userID string) []DepositLimit {
	// In a real implementation, this would query the database
	// For now, return empty slice
	return []DepositLimit{}
}

// _getUserBettingLimits gets user's betting limits
func _getUserBettingLimits(ctx context.Context, userID string) []BettingLimit {
	// In a real implementation, this would query the database
	// For now, return empty slice
	return []BettingLimit{}
}

// _getUserTimeLimits gets user's time limits
func _getUserTimeLimits(ctx context.Context, userID string) []TimeLimit {
	// In a real implementation, this would query the database
	// For now, return empty slice
	return []TimeLimit{}
}

// _getUserCoolingOffPeriods gets user's cooling off periods
func _getUserCoolingOffPeriods(ctx context.Context, userID string) []CoolingOffRecord {
	// In a real implementation, this would query the database
	// For now, return empty slice
	return []CoolingOffRecord{}
}

// _getUserInterventions gets user's intervention records
func _getUserInterventions(ctx context.Context, userID string) []Intervention {
	// In a real implementation, this would query the database
	// For now, return empty slice
	return []Intervention{}
}

// _getUserViolations gets user's violation records
func _getUserViolations(ctx context.Context, userID string) []RGViolation {
	// In a real implementation, this would query the database
	// For now, return empty slice
	return []RGViolation{}
}
