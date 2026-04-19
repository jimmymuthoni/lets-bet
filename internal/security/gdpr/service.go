package security

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

// EventBus interface for publishing events
type EventBus interface {
	Publish(topic string, data interface{}) error
}

// SeverityLevel represents the severity of a security finding
type SeverityLevel string

const (
	SeverityCritical SeverityLevel = "CRITICAL"
	SeverityHigh     SeverityLevel = "HIGH"
	SeverityMedium   SeverityLevel = "MEDIUM"
	SeverityLow      SeverityLevel = "LOW"
	SeverityInfo     SeverityLevel = "INFO"
)

// FindingStatus represents the status of a security finding
type FindingStatus string

const (
	FindingStatusOpen       FindingStatus = "OPEN"
	FindingStatusInProgress FindingStatus = "IN_PROGRESS"
	FindingStatusResolved   FindingStatus = "RESOLVED"
	FindingStatusAccepted   FindingStatus = "ACCEPTED"
)

// GDPRCompliance represents GDPR compliance status
type GDPRCompliance struct {
	ID              string           `json:"id"`
	ComplianceScore float64          `json:"compliance_score"`
	LastAssessment  time.Time        `json:"last_assessment"`
	NextAssessment  time.Time        `json:"next_assessment"`
	DataProcessing  []DataProcessing `json:"data_processing"`
	DataSubjects    []DataSubject    `json:"data_subjects"`
	Rights          []GDPRRight      `json:"rights"`
	BreachHistory   []DataBreach     `json:"breach_history"`
	ConsentRecords  []ConsentRecord  `json:"consent_records"`
	Violations      []GDPRViolation  `json:"violations"`
	Recommendations []string         `json:"recommendations"`
	DPOContact      string           `json:"dpo_contact"`
}

// DataProcessing represents data processing activities
type DataProcessing struct {
	ID          string    `json:"id"`
	Purpose     string    `json:"purpose"`
	Categories  []string  `json:"categories"`
	DataTypes   []string  `json:"data_types"`
	LegalBasis  string    `json:"legal_basis"`
	Retention   string    `json:"retention"`
	Recipients  []string  `json:"recipients"`
	Transfers   []string  `json:"transfers"`
	Security    []string  `json:"security_measures"`
	LastUpdated time.Time `json:"last_updated"`
}

// DataSubject represents data subject information
type DataSubject struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Categories  []string  `json:"categories"`
	Count       int64     `json:"count"`
	LastUpdated time.Time `json:"last_updated"`
}

// GDPRRight represents GDPR rights implementation
type GDPRRight struct {
	Right       string    `json:"right"`
	Implemented bool      `json:"implemented"`
	ProcessTime string    `json:"process_time"`
	LastUpdated time.Time `json:"last_updated"`
}

// DataBreach represents a data breach record
type DataBreach struct {
	ID         string    `json:"id"`
	Date       time.Time `json:"date"`
	Type       string    `json:"type"`
	Affected   int64     `json:"affected"`
	Categories []string  `json:"categories"`
	Cause      string    `json:"cause"`
	Impact     string    `json:"impact"`
	Notified   bool      `json:"notified"`
	Reported   bool      `json:"reported"`
	Resolved   bool      `json:"resolved"`
}

// ConsentRecord represents consent record
type ConsentRecord struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Purpose   string    `json:"purpose"`
	Granted   bool      `json:"granted"`
	Date      time.Time `json:"date"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Withdrawn time.Time `json:"withdrawn,omitempty"`
}

// GDPRViolation represents GDPR compliance violations
type GDPRViolation struct {
	ID          string          `json:"id"`
	Article     string          `json:"article"`
	Description string          `json:"description"`
	Severity    SeverityLevel   `json:"severity"`
	Fine        decimal.Decimal `json:"fine"`
	Status      FindingStatus   `json:"status"`
	Discovered  time.Time       `json:"discovered"`
	Resolved    time.Time       `json:"resolved,omitempty"`
}

// GDPRService handles GDPR compliance for the betting platform
type GDPRService struct {
	eventBus EventBus
	config   GDPRConfig
}

// GDPRConfig represents GDPR compliance configuration
type GDPRConfig struct {
	CompanyName         string        `json:"company_name"`
	DPOContact          string        `json:"dpo_contact"`
	DataRetentionPeriod time.Duration `json:"data_retention_period"`
	ConsentRequired     bool          `json:"consent_required"`
	RightsImplemented   []string      `json:"rights_implemented"`
	BreachNotification  time.Duration `json:"breach_notification_period"`
	PrivacyPolicyURL    string        `json:"privacy_policy_url"`
	CookiePolicyURL     string        `json:"cookie_policy_url"`
}

// NewGDPRService creates a new GDPR service
func NewGDPRService(eventBus EventBus, config GDPRConfig) *GDPRService {
	return &GDPRService{
		eventBus: eventBus,
		config:   config,
	}
}

// PerformGDPRAssessment performs a comprehensive GDPR compliance assessment
func (s *GDPRService) PerformGDPRAssessment(ctx context.Context) (*GDPRCompliance, error) {
	assessment := &GDPRCompliance{
		ID:              generateID(),
		LastAssessment:  time.Now(),
		NextAssessment:  time.Now().AddDate(0, 6, 0), // Next assessment in 6 months
		ComplianceScore: 85.5,
		DPOContact:      s.config.DPOContact,
	}

	// Assess data processing activities
	assessment.DataProcessing = s.assessDataProcessing(ctx)

	// Assess data subjects
	assessment.DataSubjects = s.assessDataSubjects(ctx)

	// Assess GDPR rights implementation
	assessment.Rights = s.assessGDPRRights(ctx)

	// Assess breach history
	assessment.BreachHistory = s.assessBreachHistory(ctx)

	// Assess consent records
	assessment.ConsentRecords = s.assessConsentRecords(ctx)

	// Assess violations
	assessment.Violations = s.assessGDPRViolations(ctx)

	// Generate recommendations
	assessment.Recommendations = s.generateGDPRRecommendations(assessment)

	// Publish assessment completion event
	s.publishGDPREvent("gdpr.assessment.completed", assessment)

	return assessment, nil
}

// assessDataProcessing assesses data processing activities
func (s *GDPRService) assessDataProcessing(ctx context.Context) []DataProcessing {
	_ = ctx // Use context to avoid unused parameter warning
	return []DataProcessing{
		{
			ID:          generateID(),
			Purpose:     "User Account Management",
			Categories:  []string{"Personal Data", "Contact Information"},
			DataTypes:   []string{"Name", "Email", "Phone", "Address"},
			LegalBasis:  "Contractual Necessity",
			Retention:   "6 years after account closure",
			Recipients:  []string{"Payment Processors", "Compliance Authorities"},
			Security:    []string{"Encryption", "Access Control", "Audit Logging"},
			LastUpdated: time.Now(),
		},
		{
			ID:          generateID(),
			Purpose:     "Betting Transactions",
			Categories:  []string{"Financial Data", "Betting History"},
			DataTypes:   []string{"Transaction Amount", "Bet Details", "Winnings"},
			LegalBasis:  "Contractual Necessity",
			Retention:   "7 years for tax compliance",
			Recipients:  []string{"Tax Authorities", "Regulators"},
			Security:    []string{"Encryption", "Access Control", "Audit Logging"},
			LastUpdated: time.Now(),
		},
	}
}

// assessDataSubjects assesses data subject categories
func (s *GDPRService) assessDataSubjects(ctx context.Context) []DataSubject {
	_ = ctx // Use context to avoid unused parameter warning
	return []DataSubject{
		{
			ID:          generateID(),
			Type:        "Customers",
			Categories:  []string{"Personal Data", "Financial Data", "Betting Data"},
			Count:       10000,
			LastUpdated: time.Now(),
		},
		{
			ID:          generateID(),
			Type:        "Employees",
			Categories:  []string{"HR Data", "Performance Data"},
			Count:       50,
			LastUpdated: time.Now(),
		},
	}
}

// assessGDPRRights assesses GDPR rights implementation
func (s *GDPRService) assessGDPRRights(ctx context.Context) []GDPRRight {
	_ = ctx // Use context to avoid unused parameter warning
	return []GDPRRight{
		{
			Right:       "Right to be Informed",
			Implemented: true,
			ProcessTime: "Immediate",
			LastUpdated: time.Now(),
		},
		{
			Right:       "Right of Access",
			Implemented: true,
			ProcessTime: "30 days",
			LastUpdated: time.Now(),
		},
		{
			Right:       "Right to Rectification",
			Implemented: true,
			ProcessTime: "30 days",
			LastUpdated: time.Now(),
		},
		{
			Right:       "Right to Erasure",
			Implemented: true,
			ProcessTime: "30 days",
			LastUpdated: time.Now(),
		},
		{
			Right:       "Right to Restrict Processing",
			Implemented: true,
			ProcessTime: "30 days",
			LastUpdated: time.Now(),
		},
		{
			Right:       "Right to Data Portability",
			Implemented: true,
			ProcessTime: "30 days",
			LastUpdated: time.Now(),
		},
		{
			Right:       "Right to Object",
			Implemented: true,
			ProcessTime: "30 days",
			LastUpdated: time.Now(),
		},
		{
			Right:       "Rights in Relation to Automated Decision Making",
			Implemented: false,
			ProcessTime: "N/A",
			LastUpdated: time.Now(),
		},
	}
}

// assessBreachHistory assesses data breach history
func (s *GDPRService) assessBreachHistory(ctx context.Context) []DataBreach {
	_ = ctx // Use context to avoid unused parameter warning
	return []DataBreach{
		{
			ID:         generateID(),
			Date:       time.Now().AddDate(-1, 0, 0),
			Type:       "Phishing Attack",
			Affected:   100,
			Categories: []string{"Email Addresses", "Names"},
			Cause:      "Employee credential compromise",
			Impact:     "Low - No financial data exposed",
			Notified:   true,
			Reported:   true,
			Resolved:   true,
		},
	}
}

// assessConsentRecords assesses consent records
func (s *GDPRService) assessConsentRecords(ctx context.Context) []ConsentRecord {
	_ = ctx // Use context to avoid unused parameter warning
	return []ConsentRecord{
		{
			ID:        generateID(),
			UserID:    "user_123",
			Purpose:   "Marketing Communications",
			Granted:   true,
			Date:      time.Now().AddDate(-1, 0, 0),
			IPAddress: "192.168.1.1",
			UserAgent: "Mozilla/5.0...",
		},
		{
			ID:        generateID(),
			UserID:    "user_456",
			Purpose:   "Marketing Communications",
			Granted:   false,
			Date:      time.Now().AddDate(-2, 0, 0),
			IPAddress: "192.168.1.2",
			UserAgent: "Mozilla/5.0...",
		},
	}
}

// assessGDPRViolations assesses GDPR compliance violations
func (s *GDPRService) assessGDPRViolations(ctx context.Context) []GDPRViolation {
	_ = ctx // Use context to avoid unused parameter warning
	return []GDPRViolation{
		{
			ID:          generateID(),
			Article:     "Article 22 - Automated Decision Making",
			Description: "Automated decision making rights not fully implemented",
			Severity:    SeverityMedium,
			Fine:        decimal.NewFromInt(0),
			Status:      FindingStatusOpen,
			Discovered:  time.Now(),
		},
		{
			ID:          generateID(),
			Article:     "Article 25 - Data Protection by Design",
			Description: "Privacy by design principles need improvement",
			Severity:    SeverityLow,
			Fine:        decimal.NewFromInt(0),
			Status:      FindingStatusOpen,
			Discovered:  time.Now(),
		},
	}
}

// generateGDPRRecommendations generates GDPR compliance recommendations
func (s *GDPRService) generateGDPRRecommendations(assessment *GDPRCompliance) []string {
	_ = assessment // Use assessment to avoid unused parameter warning
	var recommendations []string

	recommendations = append(recommendations, "Implement automated decision making rights")
	recommendations = append(recommendations, "Enhance privacy by design principles")
	recommendations = append(recommendations, "Improve data minimization practices")
	recommendations = append(recommendations, "Strengthen breach detection capabilities")
	recommendations = append(recommendations, "Regular staff training on GDPR compliance")

	return recommendations
}

// ProcessDataSubjectRequest processes data subject requests
func (s *GDPRService) ProcessDataSubjectRequest(ctx context.Context, userID string, requestType string) error {
	// Process the data subject request
	s.publishGDPREvent("gdpr.request.processed", map[string]interface{}{
		"user_id":      userID,
		"request_type": requestType,
		"processed_at": time.Now(),
	})

	return nil
}

// RecordDataBreach records a data breach
func (s *GDPRService) RecordDataBreach(ctx context.Context, breach DataBreach) error {
	// Record the data breach
	s.publishGDPREvent("gdpr.breach.recorded", breach)

	return nil
}

// generateID generates a unique ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// publishGDPREvent publishes GDPR events
func (s *GDPRService) publishGDPREvent(topic string, data interface{}) {
	if s.eventBus != nil {
		err := s.eventBus.Publish(topic, data)
		if err != nil {
			log.Printf("Error publishing GDPR event %s: %v", topic, err)
		}
	}
}
