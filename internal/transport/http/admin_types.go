package http

import (
	"time"

	"github.com/betting-platform/internal/admin"
)

// AdminHandler handles admin dashboard HTTP requests
type AdminHandler struct {
	adminService *admin.AdminService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminService *admin.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// DashboardRequest represents dashboard data request
type DashboardRequest struct {
	TimeRange string     `json:"time_range"` // "today", "week", "month", "year"
	FromDate  *time.Time `json:"from_date,omitempty"`
	ToDate    *time.Time `json:"to_date,omitempty"`
}

// DashboardResponse represents dashboard data response
type DashboardResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message,omitempty"`
	Data    *admin.DashboardData `json:"data,omitempty"`
}

// DashboardData represents comprehensive dashboard data
type DashboardData struct {
	Overview       *OverviewStats    `json:"overview"`
	UserMetrics    *UserMetrics      `json:"user_metrics"`
	BettingMetrics *BettingMetrics   `json:"betting_metrics"`
	FinancialStats *FinancialStats   `json:"financial_stats"`
	RecentActivity []*RecentActivity `json:"recent_activity"`
	TopUsers       []*TopUser        `json:"top_users"`
	SystemHealth   *SystemHealth     `json:"system_health"`
}

// OverviewStats represents overview statistics
type OverviewStats struct {
	TotalUsers     int64     `json:"total_users"`
	ActiveUsers    int64     `json:"active_users"`
	TotalBets      int64     `json:"total_bets"`
	TotalRevenue   float64   `json:"total_revenue"`
	TotalPayouts   float64   `json:"total_payouts"`
	GrossProfit    float64   `json:"gross_profit"`
	ConversionRate float64   `json:"conversion_rate"`
	AverageBetSize float64   `json:"average_bet_size"`
	NewUsersToday  int64     `json:"new_users_today"`
	BetsToday      int64     `json:"bets_today"`
	RevenueToday   float64   `json:"revenue_today"`
	LastUpdated    time.Time `json:"last_updated"`
}

// UserMetrics represents user-related metrics
type UserMetrics struct {
	TotalUsers         int64               `json:"total_users"`
	ActiveUsers        int64               `json:"active_users"`
	NewUsers           int64               `json:"new_users"`
	VerifiedUsers      int64               `json:"verified_users"`
	BannedUsers        int64               `json:"banned_users"`
	UserGrowthRate     float64             `json:"user_growth_rate"`
	ActiveUserRate     float64             `json:"active_user_rate"`
	VerificationRate   float64             `json:"verification_rate"`
	UserActivityByTime []*UserActivityTime `json:"user_activity_by_time"`
	UserDemographics   *UserDemographics   `json:"user_demographics"`
}

// UserActivityTime represents user activity over time
type UserActivityTime struct {
	TimePeriod  string `json:"time_period"`
	ActiveUsers int64  `json:"active_users"`
	NewUsers    int64  `json:"new_users"`
}

// UserDemographics represents user demographic data
type UserDemographics struct {
	ByCountry map[string]int64 `json:"by_country"`
	ByAge     map[string]int64 `json:"by_age"`
	ByGender  map[string]int64 `json:"by_gender"`
	ByDevice  map[string]int64 `json:"by_device"`
}

// BettingMetrics represents betting-related metrics
type BettingMetrics struct {
	TotalBets     int64            `json:"total_bets"`
	TotalStake    float64          `json:"total_stake"`
	TotalPayouts  float64          `json:"total_payouts"`
	GrossRevenue  float64          `json:"gross_revenue"`
	NetRevenue    float64          `json:"net_revenue"`
	WinRate       float64          `json:"win_rate"`
	AverageOdds   float64          `json:"average_odds"`
	AverageStake  float64          `json:"average_stake"`
	BetsBySport   []*SportBetting  `json:"bets_by_sport"`
	BetsByMarket  []*MarketBetting `json:"bets_by_market"`
	BettingTrends []*BettingTrend  `json:"betting_trends"`
	PopularEvents []*PopularEvent  `json:"popular_events"`
}

// SportBetting represents betting data by sport
type SportBetting struct {
	SportName  string  `json:"sport_name"`
	BetCount   int64   `json:"bet_count"`
	TotalStake float64 `json:"total_stake"`
	Revenue    float64 `json:"revenue"`
}

// MarketBetting represents betting data by market
type MarketBetting struct {
	MarketName string  `json:"market_name"`
	BetCount   int64   `json:"bet_count"`
	TotalStake float64 `json:"total_stake"`
	Revenue    float64 `json:"revenue"`
}

// BettingTrend represents betting trends over time
type BettingTrend struct {
	Date       string  `json:"date"`
	BetCount   int64   `json:"bet_count"`
	TotalStake float64 `json:"total_stake"`
	Revenue    float64 `json:"revenue"`
}

// PopularEvent represents popular betting events
type PopularEvent struct {
	EventID    string    `json:"event_id"`
	EventName  string    `json:"event_name"`
	SportName  string    `json:"sport_name"`
	BetCount   int64     `json:"bet_count"`
	TotalStake float64   `json:"total_stake"`
	StartTime  time.Time `json:"start_time"`
}

// FinancialStats represents financial statistics
type FinancialStats struct {
	TotalRevenue     float64           `json:"total_revenue"`
	TotalPayouts     float64           `json:"total_payouts"`
	GrossProfit      float64           `json:"gross_profit"`
	NetProfit        float64           `json:"net_profit"`
	ProfitMargin     float64           `json:"profit_margin"`
	RevenueByDay     []*DailyRevenue   `json:"revenue_by_day"`
	RevenueBySport   []*SportRevenue   `json:"revenue_by_sport"`
	RevenueByPayment []*PaymentRevenue `json:"revenue_by_payment"`
	FinancialHealth  *FinancialHealth  `json:"financial_health"`
}

// DailyRevenue represents daily revenue data
type DailyRevenue struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Payouts float64 `json:"payouts"`
	Profit  float64 `json:"profit"`
}

// SportRevenue represents revenue by sport
type SportRevenue struct {
	SportName string  `json:"sport_name"`
	Revenue   float64 `json:"revenue"`
	Payouts   float64 `json:"payouts"`
	Profit    float64 `json:"profit"`
}

// PaymentRevenue represents revenue by payment method
type PaymentRevenue struct {
	PaymentMethod    string  `json:"payment_method"`
	Revenue          float64 `json:"revenue"`
	TransactionCount int64   `json:"transaction_count"`
}

// FinancialHealth represents financial health indicators
type FinancialHealth struct {
	LiquidityRatio float64 `json:"liquidity_ratio"`
	Profitability  float64 `json:"profitability"`
	RevenueGrowth  float64 `json:"revenue_growth"`
	CostEfficiency float64 `json:"cost_efficiency"`
	RiskLevel      string  `json:"risk_level"`
	HealthScore    int     `json:"health_score"`
}

// RecentActivity represents recent system activity
type RecentActivity struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	UserID      string                 `json:"user_id,omitempty"`
	Amount      float64                `json:"amount,omitempty"`
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TopUser represents top betting users
type TopUser struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	TotalBets   int64     `json:"total_bets"`
	TotalStake  float64   `json:"total_stake"`
	TotalPayout float64   `json:"total_payout"`
	NetProfit   float64   `json:"net_profit"`
	WinRate     float64   `json:"win_rate"`
	LastActive  time.Time `json:"last_active"`
	Status      string    `json:"status"`
}

// SystemHealth represents system health status
type SystemHealth struct {
	OverallStatus    string              `json:"overall_status"`
	DatabaseStatus   *ComponentHealth    `json:"database_status"`
	RedisStatus      *ComponentHealth    `json:"redis_status"`
	APIServerStatus  *ComponentHealth    `json:"api_server_status"`
	PaymentGateways  []*GatewayHealth    `json:"payment_gateways"`
	ExternalServices []*ServiceHealth    `json:"external_services"`
	Performance      *PerformanceMetrics `json:"performance"`
	LastCheck        time.Time           `json:"last_check"`
}

// ComponentHealth represents health of a system component
type ComponentHealth struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Response  string    `json:"response"`
	Latency   int64     `json:"latency"`
	LastCheck time.Time `json:"last_check"`
	Uptime    float64   `json:"uptime"`
}

// GatewayHealth represents payment gateway health
type GatewayHealth struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Available bool      `json:"available"`
	ErrorRate float64   `json:"error_rate"`
	LastCheck time.Time `json:"last_check"`
}

// ServiceHealth represents external service health
type ServiceHealth struct {
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	ResponseTime int64     `json:"response_time"`
	LastCheck    time.Time `json:"last_check"`
	Uptime       float64   `json:"uptime"`
}

// PerformanceMetrics represents system performance metrics
type PerformanceMetrics struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	DiskUsage    float64 `json:"disk_usage"`
	NetworkIO    float64 `json:"network_io"`
	ResponseTime int64   `json:"response_time"`
	RequestRate  int64   `json:"request_rate"`
	ErrorRate    float64 `json:"error_rate"`
}

// UserManagementRequest represents user management request
type UserManagementRequest struct {
	Limit    int        `json:"limit"`
	Offset   int        `json:"offset"`
	Search   string     `json:"search,omitempty"`
	Status   string     `json:"status,omitempty"`
	SortBy   string     `json:"sort_by,omitempty"`
	SortDir  string     `json:"sort_dir,omitempty"`
	FromDate *time.Time `json:"from_date,omitempty"`
	ToDate   *time.Time `json:"to_date,omitempty"`
}

// UserManagementResponse represents user management response
type UserManagementResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// UserData represents user data for management
type UserData struct {
	Users      []*UserInfo `json:"users"`
	TotalCount int64       `json:"total_count"`
	PageInfo   *PageInfo   `json:"page_info"`
}

// UserInfo represents user information
type UserInfo struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Phone        string    `json:"phone,omitempty"`
	Country      string    `json:"country"`
	Status       string    `json:"status"`
	IsVerified   bool      `json:"is_verified"`
	IsBanned     bool      `json:"is_banned"`
	Balance      float64   `json:"balance"`
	TotalBets    int64     `json:"total_bets"`
	TotalStake   float64   `json:"total_stake"`
	TotalPayout  float64   `json:"total_payout"`
	NetProfit    float64   `json:"net_profit"`
	WinRate      float64   `json:"win_rate"`
	LastLogin    time.Time `json:"last_login"`
	RegisteredAt time.Time `json:"registered_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PageInfo represents pagination information
type PageInfo struct {
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	TotalItems  int64 `json:"total_items"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

// UserActionRequest represents user action request
type UserActionRequest struct {
	UserID string `json:"user_id"`
	Action string `json:"action"` // "ban", "unban", "verify", "reset_password", "delete"
	Reason string `json:"reason,omitempty"`
}

// UserActionResponse represents user action response
type UserActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

// SystemConfigRequest represents system configuration request
type SystemConfigRequest struct {
	Section string                 `json:"section"`
	Config  map[string]interface{} `json:"config"`
}

// SystemConfigResponse represents system configuration response
type SystemConfigResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Config  interface{} `json:"config,omitempty"`
}
