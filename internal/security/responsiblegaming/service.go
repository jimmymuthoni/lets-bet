package security

import (
	"context"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

// PerformResponsibleGamingAssessment performs a comprehensive responsible gaming assessment
func (s *ResponsibleGamingService) PerformResponsibleGamingAssessment(ctx context.Context) (*ResponsibleGaming, error) {
	assessment := &ResponsibleGaming{
		ID:              generateID(),
		LastAssessment:  time.Now(),
		NextAssessment:  time.Now().AddDate(0, 3, 0), // Next assessment in 3 months
		ComplianceScore: 0.0,
	}

	// Perform comprehensive assessment
	selfExclusion := s.assessSelfExclusion(ctx)
	depositLimits := s.assessDepositLimits(ctx)
	bettingLimits := s.assessBettingLimits(ctx)
	timeLimits := s.assessTimeLimits(ctx)
	coolingOffPeriods := s.assessCoolingOffPeriods(ctx)
	interventions := s.assessInterventions(ctx)
	education := s.assessEducation(ctx)
	violations := s.assessRGViolations(ctx)

	// Set assessment data
	assessment.SelfExclusion = selfExclusion
	assessment.DepositLimits = depositLimits
	assessment.BettingLimits = bettingLimits
	assessment.TimeLimits = timeLimits
	assessment.CoolingOffPeriods = coolingOffPeriods
	assessment.Interventions = interventions
	assessment.Education = education
	assessment.Violations = violations

	// Calculate compliance score
	assessment.ComplianceScore = _calculateComplianceScore(assessment)

	// Generate recommendations
	assessment.Recommendations = _generateRGRecommendations(assessment)

	// Publish assessment completion event
	s.publishRGEvent("responsible_gaming.assessment.completed", assessment)

	return assessment, nil
}

// GetUserResponsibleGamingStatus gets responsible gaming status for a specific user
func (s *ResponsibleGamingService) GetUserResponsibleGamingStatus(ctx context.Context, userID string) (*ResponsibleGaming, error) {
	assessment := &ResponsibleGaming{
		ID:              generateID(),
		LastAssessment:  time.Now(),
		NextAssessment:  time.Now().AddDate(0, 1, 0), // Next assessment in 1 month
		ComplianceScore: 0.0,
	}

	// Get user-specific data
	selfExclusion := _getUserSelfExclusion(ctx, userID)
	depositLimits := _getUserDepositLimits(ctx, userID)
	bettingLimits := _getUserBettingLimits(ctx, userID)
	timeLimits := _getUserTimeLimits(ctx, userID)
	coolingOffPeriods := _getUserCoolingOffPeriods(ctx, userID)
	interventions := _getUserInterventions(ctx, userID)
	violations := _getUserViolations(ctx, userID)

	// Set assessment data
	assessment.SelfExclusion = selfExclusion
	assessment.DepositLimits = depositLimits
	assessment.BettingLimits = bettingLimits
	assessment.TimeLimits = timeLimits
	assessment.CoolingOffPeriods = coolingOffPeriods
	assessment.Interventions = interventions
	assessment.Violations = violations

	// Calculate compliance score
	assessment.ComplianceScore = _calculateUserComplianceScore(assessment)

	// Generate user-specific recommendations
	assessment.Recommendations = _generateUserRecommendations(assessment, userID)

	// Publish assessment completion event
	s.publishRGEvent("responsible_gaming.user_assessment.completed", assessment)

	return assessment, nil
}

// SetSelfExclusion sets self-exclusion for a user
func (s *ResponsibleGamingService) SetSelfExclusion(ctx context.Context, userID string, duration string, reason string) error {
	// In a real implementation, this would update the database
	endDate := _calculateEndDate(duration)

	record := SelfExclusionRecord{
		ID:        generateID(),
		UserID:    userID,
		Duration:  duration,
		StartDate: time.Now(),
		EndDate:   endDate,
		Status:    "Active",
		Reason:    reason,
	}

	// Publish self-exclusion event
	s.publishRGEvent("responsible_gaming.self_exclusion.set", record)

	log.Printf("Self-exclusion set for user %s: %s until %v", userID, duration, endDate)
	return nil
}

// RemoveSelfExclusion removes self-exclusion for a user
func (s *ResponsibleGamingService) RemoveSelfExclusion(ctx context.Context, userID string) error {
	// In a real implementation, this would update the database
	record := SelfExclusionRecord{
		ID:      generateID(),
		UserID:  userID,
		Status:  "Removed",
		EndDate: time.Now(),
	}

	// Publish self-exclusion removal event
	s.publishRGEvent("responsible_gaming.self_exclusion.removed", record)

	log.Printf("Self-exclusion removed for user %s", userID)
	return nil
}

// SetDepositLimit sets deposit limit for a user
func (s *ResponsibleGamingService) SetDepositLimit(ctx context.Context, userID string, limitType string, amount string, period string) error {
	// In a real implementation, this would update the database
	amountDecimal, _ := decimal.NewFromString(amount)

	limit := DepositLimit{
		ID:         generateID(),
		UserID:     userID,
		Type:       limitType,
		Amount:     amountDecimal,
		Period:     period,
		Status:     "Active",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	// Publish deposit limit event
	s.publishRGEvent("responsible_gaming.deposit_limit.set", limit)

	log.Printf("Deposit limit set for user %s: %s %s per %s", userID, amount, limitType, period)
	return nil
}

// SetBettingLimit sets betting limit for a user
func (s *ResponsibleGamingService) SetBettingLimit(ctx context.Context, userID string, limitType string, amount string, period string) error {
	// In a real implementation, this would update the database
	amountDecimal, _ := decimal.NewFromString(amount)

	limit := BettingLimit{
		ID:         generateID(),
		UserID:     userID,
		Type:       limitType,
		Amount:     amountDecimal,
		Period:     period,
		Status:     "Active",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	// Publish betting limit event
	s.publishRGEvent("responsible_gaming.betting_limit.set", limit)

	log.Printf("Betting limit set for user %s: %s %s per %s", userID, amount, limitType, period)
	return nil
}

// publishRGEvent publishes responsible gaming events
func (s *ResponsibleGamingService) publishRGEvent(topic string, data any) {
	if s.eventBus != nil {
		err := s.eventBus.Publish(topic, data)
		if err != nil {
			log.Printf("Error publishing responsible gaming event %s: %v", topic, err)
		}
	}
}
