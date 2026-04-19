package http

import (
	"time"

	"github.com/shopspring/decimal"
)

// SecurityAuditRequest represents a security audit request
type SecurityAuditRequest struct {
	Scope       []string `json:"scope"`        // "authentication", "authorization", "data", "infrastructure"
	DeepScan    bool     `json:"deep_scan"`    // Perform deep scan
	IncludeLogs bool     `json:"include_logs"` // Include logs in audit
}

// SecurityAuditResponse represents security audit response
type SecurityAuditResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// SecurityAudit represents comprehensive security audit results
type SecurityAudit struct {
	ID              string                    `json:"id"`
	ExecutedAt      time.Time                 `json:"executed_at"`
	Scope           []string                  `json:"scope"`
	OverallScore    int                       `json:"overall_score"` // 0-100
	RiskLevel       string                    `json:"risk_level"`    // "LOW", "MEDIUM", "HIGH", "CRITICAL"
	Findings        []*SecurityFinding        `json:"findings"`
	Recommendations []*SecurityRecommendation `json:"recommendations"`
	ComplianceScore *ComplianceScore          `json:"compliance_score"`
	Vulnerabilities []*Vulnerability          `json:"vulnerabilities"`
	Summary         *AuditSummary             `json:"summary"`
}

// SecurityFinding represents a security finding
type SecurityFinding struct {
	ID          string     `json:"id"`
	Category    string     `json:"category"` // "AUTH", "DATA", "INFRA", "COMPLIANCE"
	Severity    string     `json:"severity"` // "LOW", "MEDIUM", "HIGH", "CRITICAL"
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Impact      string     `json:"impact"`
	Affected    []string   `json:"affected"`
	FoundAt     time.Time  `json:"found_at"`
	Status      string     `json:"status"` // "OPEN", "IN_PROGRESS", "RESOLVED"
	AssignedTo  string     `json:"assigned_to,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// SecurityRecommendation represents a security recommendation
type SecurityRecommendation struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"` // "LOW", "MEDIUM", "HIGH", "CRITICAL"
	Category    string     `json:"category"`
	Effort      string     `json:"effort"` // "LOW", "MEDIUM", "HIGH"
	Impact      string     `json:"impact"`
	Status      string     `json:"status"` // "PENDING", "IN_PROGRESS", "COMPLETED"
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// ComplianceScore represents compliance scores
type ComplianceScore struct {
	GDPR        int       `json:"gdpr"`    // 0-100
	PCI         int       `json:"pci"`     // 0-100
	SOX         int       `json:"sox"`     // 0-100
	Overall     int       `json:"overall"` // 0-100
	LastChecked time.Time `json:"last_checked"`
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Severity     string     `json:"severity"`
	CVSSScore    float64    `json:"cvss_score"`
	CVE          string     `json:"cve,omitempty"`
	Component    string     `json:"component"`
	Version      string     `json:"version"`
	DiscoveredAt time.Time  `json:"discovered_at"`
	PublishedAt  time.Time  `json:"published_at"`
	Status       string     `json:"status"`
	PatchedAt    *time.Time `json:"patched_at,omitempty"`
	Affected     []string   `json:"affected"`
	References   []string   `json:"references"`
}

// AuditSummary represents audit summary
type AuditSummary struct {
	TotalFindings    int                  `json:"total_findings"`
	CriticalFindings int                  `json:"critical_findings"`
	HighFindings     int                  `json:"high_findings"`
	MediumFindings   int                  `json:"medium_findings"`
	LowFindings      int                  `json:"low_findings"`
	OpenFindings     int                  `json:"open_findings"`
	ResolvedFindings int                  `json:"resolved_findings"`
	Categories       map[string]int       `json:"categories"`
	Timeline         []*AuditTimelineItem `json:"timeline"`
	RiskDistribution map[string]int       `json:"risk_distribution"`
}

// AuditTimelineItem represents an audit timeline item
type AuditTimelineItem struct {
	Timestamp time.Time `json:"timestamp"`
	Event     string    `json:"event"`
	Details   string    `json:"details"`
}

// PenetrationTestRequest represents penetration test request
type PenetrationTestRequest struct {
	TestType     string    `json:"test_type"` // "BLACK_BOX", "WHITE_BOX", "GRAY_BOX"
	Scope        []string  `json:"scope"`     // "WEB", "API", "MOBILE", "NETWORK"
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Objectives   []string  `json:"objectives"`
	Constraints  []string  `json:"constraints"`
	ReportFormat string    `json:"report_format"` // "DETAILED", "EXECUTIVE", "TECHNICAL"
}

// PenetrationTestResponse represents penetration test response
type PenetrationTestResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// PenetrationTest represents penetration test results
type PenetrationTest struct {
	ID              string                   `json:"id"`
	TestType        string                   `json:"test_type"`
	Status          string                   `json:"status"` // "PLANNED", "IN_PROGRESS", "COMPLETED", "CANCELLED"
	StartDate       time.Time                `json:"start_date"`
	EndDate         time.Time                `json:"end_date"`
	Testers         []*PenetrationTester     `json:"testers"`
	Scope           []string                 `json:"scope"`
	Objectives      []string                 `json:"objectives"`
	Findings        []*PentestFinding        `json:"findings"`
	RiskScore       int                      `json:"risk_score"` // 0-100
	Exploits        []*Exploit               `json:"exploits"`
	Report          *PenetrationTestReport   `json:"report"`
	Recommendations []*PentestRecommendation `json:"recommendations"`
}

// PenetrationTester represents a penetration tester
type PenetrationTester struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	Role           string   `json:"role"`
	Certifications []string `json:"certifications"`
}

// PentestFinding represents a penetration test finding
type PentestFinding struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Category    string    `json:"category"`
	Component   string    `json:"component"`
	Steps       []string  `json:"steps"`
	Evidence    []string  `json:"evidence"`
	Impact      string    `json:"impact"`
	Remediation string    `json:"remediation"`
	Discovered  time.Time `json:"discovered"`
}

// Exploit represents a discovered exploit
type Exploit struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // "REMOTE", "LOCAL", "WEB", "NETWORK"`
	Severity    string    `json:"severity"`
	Component   string    `json:"component"`
	Payload     string    `json:"payload,omitempty"`
	Steps       []string  `json:"steps"`
	Proof       string    `json:"proof"`
	Discovered  time.Time `json:"discovered"`
}

// PenetrationTestReport represents penetration test report
type PenetrationTestReport struct {
	ExecutiveSummary string                 `json:"executive_summary"`
	TechnicalDetails string                 `json:"technical_details"`
	RiskAssessment   *RiskAssessment        `json:"risk_assessment"`
	Methodology      string                 `json:"methodology"`
	Limitations      []string               `json:"limitations"`
	Appendices       map[string]interface{} `json:"appendices"`
}

// PentestRecommendation represents penetration test recommendation
type PentestRecommendation struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Category    string `json:"category"`
	Cost        string `json:"cost"`     // "LOW", "MEDIUM", "HIGH"
	Timeline    string `json:"timeline"` // "SHORT", "MEDIUM", "LONG"
}

// RiskAssessment represents risk assessment
type RiskAssessment struct {
	OverallRisk    int                   `json:"overall_risk"` // 0-100
	RiskFactors    []*RiskFactor         `json:"risk_factors"`
	MitigationPlan []*MitigationStrategy `json:"mitigation_plan"`
	ResidualRisk   int                   `json:"residual_risk"`
}

// RiskFactor represents a risk factor
type RiskFactor struct {
	Factor     string `json:"factor"`
	Impact     string `json:"impact"`
	Likelihood string `json:"likelihood"`
	Score      int    `json:"score"`
}

// MitigationStrategy represents a mitigation strategy
type MitigationStrategy struct {
	Strategy    string `json:"strategy"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Timeline    string `json:"timeline"`
	Owner       string `json:"owner"`
}

// GDPRRequest represents GDPR request
type GDPRRequest struct {
	Type        string    `json:"type"` // "DATA_ACCESS", "DATA_DELETION", "DATA_PORTABILITY", "RECTIFICATION"
	UserID      string    `json:"user_id"`
	Reason      string    `json:"reason"`
	ContactInfo string    `json:"contact_info"`
	RequestedAt time.Time `json:"requested_at"`
}

// GDPRResponse represents GDPR response
type GDPRResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// GDPRResponseData represents GDPR response data
type GDPRResponseData struct {
	RequestID   string                 `json:"request_id"`
	Status      string                 `json:"status"` // "PENDING", "PROCESSING", "COMPLETED", "REJECTED"
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Reason      string                 `json:"reason,omitempty"`
}

// ResponsibleGamingRequest represents responsible gaming request
type ResponsibleGamingRequest struct {
	UserID      string                 `json:"user_id"`
	Action      string                 `json:"action"` // "SET_LIMIT", "EXCLUDE", "SELF_ASSESS", "TIMEOUT"
	Parameters  map[string]interface{} `json:"parameters"`
	Reason      string                 `json:"reason"`
	RequestedAt time.Time              `json:"requested_at"`
}

// ResponsibleGamingResponse represents responsible gaming response
type ResponsibleGamingResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ResponsibleGamingData represents responsible gaming data
type ResponsibleGamingData struct {
	UserID          string                  `json:"user_id"`
	CurrentLimits   []*GamingLimit          `json:"current_limits"`
	ExclusionStatus *ExclusionStatus        `json:"exclusion_status"`
	SelfAssessment  *SelfAssessmentResult   `json:"self_assessment"`
	ActivityHistory []*GamingActivity       `json:"activity_history"`
	RiskProfile     *RiskProfile            `json:"risk_profile"`
	Recommendations []*GamingRecommendation `json:"recommendations"`
}

// GamingLimit represents gaming limits
type GamingLimit struct {
	Type      string          `json:"type"` // "DAILY", "WEEKLY", "MONTHLY"
	Amount    decimal.Decimal `json:"amount"`
	Duration  time.Duration   `json:"duration"`
	Active    bool            `json:"active"`
	SetAt     time.Time       `json:"set_at"`
	ExpiresAt *time.Time      `json:"expires_at,omitempty"`
}

// ExclusionStatus represents exclusion status
type ExclusionStatus struct {
	Status    string        `json:"status"` // "ACTIVE", "INACTIVE", "EXPIRED"
	Type      string        `json:"type"`   // "TEMPORARY", "PERMANENT"
	Duration  time.Duration `json:"duration"`
	StartedAt time.Time     `json:"started_at"`
	EndsAt    *time.Time    `json:"ends_at,omitempty"`
	Reason    string        `json:"reason"`
}

// SelfAssessmentResult represents self-assessment result
type SelfAssessmentResult struct {
	Score           int                 `json:"score"`      // 0-100
	RiskLevel       string              `json:"risk_level"` // "LOW", "MEDIUM", "HIGH"
	Answers         []*AssessmentAnswer `json:"answers"`
	Recommendations []string            `json:"recommendations"`
	CompletedAt     time.Time           `json:"completed_at"`
	NextAssessment  time.Time           `json:"next_assessment"`
}

// AssessmentAnswer represents assessment answer
type AssessmentAnswer struct {
	QuestionID string `json:"question_id"`
	Answer     string `json:"answer"`
	Score      int    `json:"score"`
}

// GamingActivity represents gaming activity
type GamingActivity struct {
	Date        time.Time       `json:"date"`
	TotalBets   int64           `json:"total_bets"`
	TotalStake  decimal.Decimal `json:"total_stake"`
	TotalPayout decimal.Decimal `json:"total_payout"`
	NetResult   decimal.Decimal `json:"net_result"`
	SessionTime time.Duration   `json:"session_time"`
	GamesPlayed int             `json:"games_played"`
}

// RiskProfile represents risk profile
type RiskProfile struct {
	OverallRisk    string    `json:"overall_risk"`
	BehavioralRisk string    `json:"behavioral_risk"`
	FinancialRisk  string    `json:"financial_risk"`
	TimeRisk       string    `json:"time_risk"`
	Factors        []string  `json:"factors"`
	LastUpdated    time.Time `json:"last_updated"`
}

// GamingRecommendation represents gaming recommendation
type GamingRecommendation struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Actionable  bool   `json:"actionable"`
}
