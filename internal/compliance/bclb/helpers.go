package compliance

import (
	"context"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

// getUserLimits retrieves user-specific betting limits
func (s *BCLBService) getUserLimits(userID string) *UserLimits {
	// Get user-specific limits from database
	user, err := s.userRepo.GetUser(context.Background(), userID)
	if err != nil {
		log.Printf("Error getting user %s: %v", userID, err)
		// Return default limits if user not found
		return &UserLimits{
			UserID:       userID,
			DailyLimit:   s.config.MaxDailyStake,
			WeeklyLimit:  s.config.MaxWeeklyStake,
			MonthlyLimit: s.config.MaxMonthlyStake,
			MaxBetSize:   s.config.MaxBetPerEvent,
			LastUpdated:  time.Now(),
		}
	}

	// In a real implementation, user limits would be stored in a separate table
	// For now, use config defaults but could be overridden by user-specific settings
	return &UserLimits{
		UserID:       userID,
		DailyLimit:   s.config.MaxDailyStake,
		WeeklyLimit:  s.config.MaxWeeklyStake,
		MonthlyLimit: s.config.MaxMonthlyStake,
		MaxBetSize:   s.config.MaxBetPerEvent,
		LastUpdated:  user.UpdatedAt,
	}
}

// isAgeVerified checks if user's age is verified
func (s *BCLBService) isAgeVerified(userID string) bool {
	// Check user's age verification status from database
	ageVerification, err := s.userRepo.GetUserAgeVerification(context.Background(), userID)
	if err != nil {
		log.Printf("Error getting age verification for user %s: %v", userID, err)
		return false
	}

	// Check if verification is still valid
	if !ageVerification.Verified {
		return false
	}

	// Check if verification has expired
	if time.Now().After(ageVerification.ExpiresAt) {
		return false
	}

	// Additional check: verify user is actually old enough
	user, err := s.userRepo.GetUser(context.Background(), userID)
	if err != nil {
		log.Printf("Error getting user %s for age check: %v", userID, err)
		return false
	}

	// Calculate user's age
	now := time.Now()
	age := now.Year() - user.DateOfBirth.Year()

	// Adjust for birthday not yet occurred this year
	if now.Month() < user.DateOfBirth.Month() ||
		(now.Month() == user.DateOfBirth.Month() && now.Day() < user.DateOfBirth.Day()) {
		age--
	}

	return age >= s.config.MinAge
}

// isKYCCompliant checks if user's KYC is compliant
func (s *BCLBService) isKYCCompliant(userID string) bool {
	// Check user's KYC status from database
	kycStatus, err := s.userRepo.GetUserKYCStatus(context.Background(), userID)
	if err != nil {
		log.Printf("Error getting KYC status for user %s: %v", userID, err)
		return false
	}

	// Check if KYC is verified and meets required level
	if kycStatus.Status != "VERIFIED" {
		return false
	}

	// Check if KYC level meets requirements
	switch s.config.RequiredKYCLevel {
	case "ENHANCED":
		return kycStatus.Level == "ENHANCED"
	case "STANDARD":
		return kycStatus.Level == "STANDARD" || kycStatus.Level == "ENHANCED"
	case "BASIC":
		return kycStatus.Level == "BASIC" || kycStatus.Level == "STANDARD" || kycStatus.Level == "ENHANCED"
	default:
		return kycStatus.Status == "VERIFIED"
	}
}

// isSelfExcluded checks if user is self-excluded
func (s *BCLBService) isSelfExcluded(userID string) bool {
	// Check user's self-exclusion status from database
	selfExclusion, err := s.userRepo.GetUserSelfExclusion(context.Background(), userID)
	if err != nil {
		log.Printf("Error getting self-exclusion status for user %s: %v", userID, err)
		return false
	}

	// Check if self-exclusion is active
	if !selfExclusion.Active {
		return false
	}

	// Check if self-exclusion period has expired
	if time.Now().After(selfExclusion.EndDate) {
		// Self-exclusion has expired, update database
		err = s.userRepo.UpdateSelfExclusion(context.Background(), userID, time.Time{}, "expired")
		if err != nil {
			log.Printf("Error updating expired self-exclusion for user %s: %v", userID, err)
		}
		return false
	}

	return true
}

// getDailyStake calculates user's daily stake from database
func (s *BCLBService) getDailyStake(userID string) decimal.Decimal {
	// Calculate user's daily stake from database
	today := time.Now().Truncate(24 * time.Hour) // Start of today
	dailyStake, err := s.betRepo.GetUserDailyStake(context.Background(), userID, today)
	if err != nil {
		log.Printf("Error getting daily stake for user %s: %v", userID, err)
		return decimal.Zero
	}

	return dailyStake
}

// getWeeklyStake calculates user's weekly stake from database
func (s *BCLBService) getWeeklyStake(userID string) decimal.Decimal {
	// Calculate user's weekly stake from database
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday())) // Start of week (Sunday)
	weeklyStake, err := s.betRepo.GetUserWeeklyStake(context.Background(), userID, weekStart)
	if err != nil {
		log.Printf("Error getting weekly stake for user %s: %v", userID, err)
		return decimal.Zero
	}

	return weeklyStake
}

// getMonthlyStake calculates user's monthly stake from database
func (s *BCLBService) getMonthlyStake(userID string) decimal.Decimal {
	// Calculate user's monthly stake from database
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthlyStake, err := s.betRepo.GetUserMonthlyStake(context.Background(), userID, monthStart)
	if err != nil {
		log.Printf("Error getting monthly stake for user %s: %v", userID, err)
		return decimal.Zero
	}

	return monthlyStake
}

// isInCoolingOffPeriod checks if user is in cooling off period
func (s *BCLBService) isInCoolingOffPeriod(userID string) bool {
	// Check if user is in cooling off period from database
	inCoolingOff, err := s.coolingOffRepo.IsUserInCoolingOffPeriod(context.Background(), userID)
	if err != nil {
		log.Printf("Error checking cooling off period for user %s: %v", userID, err)
		return false
	}

	return inCoolingOff
}
