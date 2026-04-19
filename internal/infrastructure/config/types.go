package config

import (
	"fmt"
	"time"
)

// Config is the centralized configuration for all services.
type Config struct {
	Service     ServiceConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	NATS        NATSConfig
	Tenant      TenantConfig
	JWT         JWTConfig
	Security    SecurityConfig
	MPesa       MPesaConfig
	Flutterwave FlutterwaveConfig
	Sportradar  SportradarConfig
	SmileID     SmileIDConfig
	Tax         TaxConfig
	Crash       CrashConfig
	Bet         BetConfig
	Logging     LoggingConfig
	Features    FeatureFlags
}

type ServiceConfig struct {
	Name        string
	Environment string // development, staging, production
	Port        int
}

type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		d.Host, d.Port, d.Name, d.User, d.Password, d.SSLMode,
	)
}

type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type NATSConfig struct {
	URL           string
	MaxReconnects int
	ReconnectWait time.Duration
	Timeout       time.Duration
	PingInterval  time.Duration
	MaxPingsOut   int
}

type TenantConfig struct {
	DefaultCountry      string
	DefaultCurrency     string
	SupportedCountries  []string
	SupportedCurrencies []string
	AllowedCountries    []string
}

type JWTConfig struct {
	Secret         string
	ExpirationTime time.Duration
	Issuer         string
	RefreshTime    time.Duration
}

type SecurityConfig struct {
	CORSOrigins           []string
	CORSAllowedOrigins    []string
	CORSAllowedMethods    []string
	CORSAllowedHeaders    []string
	RateLimitPerMinute    int
	RateLimitRequests     int
	RateLimitWindow       time.Duration
	MaxBodySize           int64
	EnableHTTPS           bool
	SessionTimeout        time.Duration
	PasswordMinLength     int
	PasswordRequireUpper  bool
	PasswordRequireLower  bool
	PasswordRequireNumber bool
	PasswordRequireSymbol bool
}

type MPesaConfig struct {
	ConsumerKey        string
	ConsumerSecret     string
	ShortCode          string
	PassKey            string
	InitiatorName      string
	SecurityCredential string
	Environment        string
	CallbackURL        string
	Timeout            time.Duration
}

type FlutterwaveConfig struct {
	PublicKey     string
	SecretKey     string
	EncryptionKey string
	BaseURL       string
	WebhookSecret string
	Timeout       time.Duration
}

type SportradarConfig struct {
	APIKey           string
	BaseURL          string
	Timeout          time.Duration
	RateLimit        int
	EnableSoccer     bool
	EnableBasketball bool
	EnableTennis     bool
	EnableCricket    bool
}

type SmileIDConfig struct {
	PartnerID            string
	APIKey               string
	BaseURL              string
	Environment          string
	Timeout              time.Duration
	EnableKYC            bool
	EnableIDVerification bool
}

type TaxConfig struct {
	Enabled            bool
	DefaultRate        float64
	WHTRate            float64
	GamingTaxRate      float64
	TransactionTaxRate float64
	Currency           string
	AutoCalculate      bool
}

type CrashConfig struct {
	Enabled       bool
	BaseURL       string
	APIKey        string
	SecretKey     string
	Timeout       time.Duration
	MaxBetAmount  float64
	MinBetAmount  float64
	MaxMultiplier float64
	MinMultiplier float64
	HouseEdge     float64
}

type BetConfig struct {
	MinBetAmount    float64
	MaxBetAmount    float64
	MaxParlaySize   int
	MaxOdds         float64
	MinOdds         float64
	StakeTimeout    time.Duration
	SettlementDelay time.Duration
	CancelDelay     time.Duration
}

type LoggingConfig struct {
	Level      string
	Format     string
	Output     string
	File       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

type FeatureFlags struct {
	EnableLiveBetting       bool
	EnableVirtualSports     bool
	EnableJackpot           bool
	EnablePromotions        bool
	EnableNotifications     bool
	EnableAnalytics         bool
	EnableResponsibleGaming bool
	EnableMultiCurrency     bool
	EnableMobileApp         bool
	EnableAPIV2             bool
	EnableBetaFeatures      bool
}

// ConfigLoader interface for loading configuration
type ConfigLoader interface {
	Load() (*Config, error)
	Validate(*Config) error
}

// Environment variables
const (
	// Service
	EnvServiceName        = "SERVICE_NAME"
	EnvServiceEnvironment = "SERVICE_ENVIRONMENT"
	EnvServicePort        = "SERVICE_PORT"

	// Database
	EnvDBHost            = "DB_HOST"
	EnvDBPort            = "DB_PORT"
	EnvDBName            = "DB_NAME"
	EnvDBUser            = "DB_USER"
	EnvDBPassword        = "DB_PASSWORD"
	EnvDBSSLMode         = "DB_SSL_MODE"
	EnvDBMaxOpenConns    = "DB_MAX_OPEN_CONNS"
	EnvDBMaxIdleConns    = "DB_MAX_IDLE_CONNS"
	EnvDBConnMaxLifetime = "DB_CONN_MAX_LIFETIME"

	// Redis
	EnvRedisHost         = "REDIS_HOST"
	EnvRedisPort         = "REDIS_PORT"
	EnvRedisPassword     = "REDIS_PASSWORD"
	EnvRedisDB           = "REDIS_DB"
	EnvRedisPoolSize     = "REDIS_POOL_SIZE"
	EnvRedisMinIdleConns = "REDIS_MIN_IDLE_CONNS"
	EnvRedisDialTimeout  = "REDIS_DIAL_TIMEOUT"
	EnvRedisReadTimeout  = "REDIS_READ_TIMEOUT"
	EnvRedisWriteTimeout = "REDIS_WRITE_TIMEOUT"

	// NATS
	EnvNATSURL           = "NATS_URL"
	EnvNATSMaxReconnects = "NATS_MAX_RECONNECTS"
	EnvNATSReconnectWait = "NATS_RECONNECT_WAIT"
	EnvNATSTimeout       = "NATS_TIMEOUT"
	EnvNATSPingInterval  = "NATS_PING_INTERVAL"
	EnvNATSMaxPingsOut   = "NATS_MAX_PINGS_OUT"

	// Tenant
	EnvTenantDefaultCountry  = "TENANT_DEFAULT_COUNTRY"
	EnvTenantDefaultCurrency = "TENANT_DEFAULT_CURRENCY"

	// JWT
	EnvJWTSecret         = "JWT_SECRET"
	EnvJWTExpirationTime = "JWT_EXPIRATION_TIME"
	EnvJWTIssuer         = "JWT_ISSUER"
	EnvJWTRefreshTime    = "JWT_REFRESH_TIME"

	// Security
	EnvSecurityCORSOrigins           = "SECURITY_CORS_ORIGINS"
	EnvSecurityCORSAllowedOrigins    = "SECURITY_CORS_ALLOWED_ORIGINS"
	EnvSecurityCORSAllowedMethods    = "SECURITY_CORS_ALLOWED_METHODS"
	EnvSecurityCORSAllowedHeaders    = "SECURITY_CORS_ALLOWED_HEADERS"
	EnvSecurityRateLimitPerMinute    = "SECURITY_RATE_LIMIT_PER_MINUTE"
	EnvSecurityRateLimitRequests     = "SECURITY_RATE_LIMIT_REQUESTS"
	EnvSecurityRateLimitWindow       = "SECURITY_RATE_LIMIT_WINDOW"
	EnvSecurityMaxBodySize           = "SECURITY_MAX_BODY_SIZE"
	EnvSecurityEnableHTTPS           = "SECURITY_ENABLE_HTTPS"
	EnvSecuritySessionTimeout        = "SECURITY_SESSION_TIMEOUT"
	EnvSecurityPasswordMinLength     = "SECURITY_PASSWORD_MIN_LENGTH"
	EnvSecurityPasswordRequireUpper  = "SECURITY_PASSWORD_REQUIRE_UPPER"
	EnvSecurityPasswordRequireLower  = "SECURITY_PASSWORD_REQUIRE_LOWER"
	EnvSecurityPasswordRequireNumber = "SECURITY_PASSWORD_REQUIRE_NUMBER"
	EnvSecurityPasswordRequireSymbol = "SECURITY_PASSWORD_REQUIRE_SYMBOL"

	// M-Pesa
	EnvMPesaConsumerKey        = "MPESA_CONSUMER_KEY"
	EnvMPesaConsumerSecret     = "MPESA_CONSUMER_SECRET"
	EnvMPesaShortCode          = "MPESA_SHORT_CODE"
	EnvMPesaPassKey            = "MPESA_PASS_KEY"
	EnvMPesaInitiatorName      = "MPESA_INITIATOR_NAME"
	EnvMPesaSecurityCredential = "MPESA_SECURITY_CREDENTIAL"
	EnvMPesaEnvironment        = "MPESA_ENVIRONMENT"
	EnvMPesaCallbackURL        = "MPESA_CALLBACK_URL"
	EnvMPesaTimeout            = "MPESA_TIMEOUT"

	// Flutterwave
	EnvFlutterwavePublicKey     = "FLUTTERWAVE_PUBLIC_KEY"
	EnvFlutterwaveSecretKey     = "FLUTTERWAVE_SECRET_KEY"
	EnvFlutterwaveEncryptionKey = "FLUTTERWAVE_ENCRYPTION_KEY"
	EnvFlutterwaveBaseURL       = "FLUTTERWAVE_BASE_URL"
	EnvFlutterwaveWebhookSecret = "FLUTTERWAVE_WEBHOOK_SECRET"
	EnvFlutterwaveTimeout       = "FLUTTERWAVE_TIMEOUT"

	// Sportradar
	EnvSportradarAPIKey           = "SPORTRADAR_API_KEY"
	EnvSportradarBaseURL          = "SPORTRADAR_BASE_URL"
	EnvSportradarTimeout          = "SPORTRADAR_TIMEOUT"
	EnvSportradarRateLimit        = "SPORTRADAR_RATE_LIMIT"
	EnvSportradarEnableSoccer     = "SPORTRADAR_ENABLE_SOCCER"
	EnvSportradarEnableBasketball = "SPORTRADAR_ENABLE_BASKETBALL"
	EnvSportradarEnableTennis     = "SPORTRADAR_ENABLE_TENNIS"
	EnvSportradarEnableCricket    = "SPORTRADAR_ENABLE_CRICKET"

	// SmileID
	EnvSmileIDPartnerID            = "SMILE_ID_PARTNER_ID"
	EnvSmileIDAPIKey               = "SMILE_ID_API_KEY"
	EnvSmileIDBaseURL              = "SMILE_ID_BASE_URL"
	EnvSmileIDEnvironment          = "SMILE_ID_ENVIRONMENT"
	EnvSmileIDTimeout              = "SMILE_ID_TIMEOUT"
	EnvSmileIDEnableKYC            = "SMILE_ID_ENABLE_KYC"
	EnvSmileIDEnableIDVerification = "SMILE_ID_ENABLE_ID_VERIFICATION"

	// Tax
	EnvTaxEnabled            = "TAX_ENABLED"
	EnvTaxDefaultRate        = "TAX_DEFAULT_RATE"
	EnvTaxWHTRate            = "TAX_WHT_RATE"
	EnvTaxGamingTaxRate      = "TAX_GAMING_TAX_RATE"
	EnvTaxTransactionTaxRate = "TAX_TRANSACTION_TAX_RATE"
	EnvTaxCurrency           = "TAX_CURRENCY"
	EnvTaxAutoCalculate      = "TAX_AUTO_CALCULATE"

	// Crash
	EnvCrashEnabled       = "CRASH_ENABLED"
	EnvCrashBaseURL       = "CRASH_BASE_URL"
	EnvCrashAPIKey        = "CRASH_API_KEY"
	EnvCrashSecretKey     = "CRASH_SECRET_KEY"
	EnvCrashTimeout       = "CRASH_TIMEOUT"
	EnvCrashMaxBetAmount  = "CRASH_MAX_BET_AMOUNT"
	EnvCrashMinBetAmount  = "CRASH_MIN_BET_AMOUNT"
	EnvCrashMaxMultiplier = "CRASH_MAX_MULTIPLIER"
	EnvCrashMinMultiplier = "CRASH_MIN_MULTIPLIER"
	EnvCrashHouseEdge     = "CRASH_HOUSE_EDGE"

	// Bet
	EnvBetMinBetAmount    = "BET_MIN_BET_AMOUNT"
	EnvBetMaxBetAmount    = "BET_MAX_BET_AMOUNT"
	EnvBetMaxParlaySize   = "BET_MAX_PARLAY_SIZE"
	EnvBetMaxOdds         = "BET_MAX_ODDS"
	EnvBetMinOdds         = "BET_MIN_ODDS"
	EnvBetStakeTimeout    = "BET_STAKE_TIMEOUT"
	EnvBetSettlementDelay = "BET_SETTLEMENT_DELAY"
	EnvBetCancelDelay     = "BET_CANCEL_DELAY"

	// Logging
	EnvLoggingLevel      = "LOGGING_LEVEL"
	EnvLoggingFormat     = "LOGGING_FORMAT"
	EnvLoggingOutput     = "LOGGING_OUTPUT"
	EnvLoggingFile       = "LOGGING_FILE"
	EnvLoggingMaxSize    = "LOGGING_MAX_SIZE"
	EnvLoggingMaxBackups = "LOGGING_MAX_BACKUPS"
	EnvLoggingMaxAge     = "LOGGING_MAX_AGE"
	EnvLoggingCompress   = "LOGGING_COMPRESS"

	// Feature Flags
	EnvFeatureEnableLiveBetting       = "FEATURE_ENABLE_LIVE_BETTING"
	EnvFeatureEnableVirtualSports     = "FEATURE_ENABLE_VIRTUAL_SPORTS"
	EnvFeatureEnableJackpot           = "FEATURE_ENABLE_JACKPOT"
	EnvFeatureEnablePromotions        = "FEATURE_ENABLE_PROMOTIONS"
	EnvFeatureEnableNotifications     = "FEATURE_ENABLE_NOTIFICATIONS"
	EnvFeatureEnableAnalytics         = "FEATURE_ENABLE_ANALYTICS"
	EnvFeatureEnableResponsibleGaming = "FEATURE_ENABLE_RESPONSIBLE_GAMING"
	EnvFeatureEnableMultiCurrency     = "FEATURE_ENABLE_MULTI_CURRENCY"
	EnvFeatureEnableMobileApp         = "FEATURE_ENABLE_MOBILE_APP"
	EnvFeatureEnableAPIV2             = "FEATURE_ENABLE_API_V2"
	EnvFeatureEnableBetaFeatures      = "FEATURE_ENABLE_BETA_FEATURES"
)

// Default values
const (
	DefaultServiceName        = "betting-platform"
	DefaultServiceEnvironment = "development"
	DefaultServicePort        = 8080

	DefaultDBHost            = "localhost"
	DefaultDBPort            = 5432
	DefaultDBName            = "betting_platform"
	DefaultDBUser            = "postgres"
	DefaultDBSSLMode         = "disable"
	DefaultDBMaxOpenConns    = 25
	DefaultDBMaxIdleConns    = 5
	DefaultDBConnMaxLifetime = 5 * time.Minute

	DefaultRedisHost         = "localhost"
	DefaultRedisPort         = 6379
	DefaultRedisDB           = 0
	DefaultRedisPoolSize     = 10
	DefaultRedisMinIdleConns = 5
	DefaultRedisDialTimeout  = 5 * time.Second
	DefaultRedisReadTimeout  = 3 * time.Second
	DefaultRedisWriteTimeout = 3 * time.Second

	DefaultNATSURL           = "nats://localhost:4222"
	DefaultNATSMaxReconnects = 5
	DefaultNATSReconnectWait = 2 * time.Second
	DefaultNATSTimeout       = 5 * time.Second
	DefaultNATSPingInterval  = 2 * time.Minute
	DefaultNATSMaxPingsOut   = 3

	DefaultTenantDefaultCountry  = "KE"
	DefaultTenantDefaultCurrency = "KES"

	DefaultJWTExpirationTime = 24 * time.Hour
	DefaultJWTRefreshTime    = 168 * time.Hour // 7 days
	DefaultJWTIssuer         = "betting-platform"

	DefaultSecurityRateLimitPerMinute    = 100
	DefaultSecurityMaxBodySize           = 10 * 1024 * 1024 // 10MB
	DefaultSecurityEnableHTTPS           = true
	DefaultSecuritySessionTimeout        = 24 * time.Hour
	DefaultSecurityPasswordMinLength     = 8
	DefaultSecurityPasswordRequireUpper  = true
	DefaultSecurityPasswordRequireLower  = true
	DefaultSecurityPasswordRequireNumber = true
	DefaultSecurityPasswordRequireSymbol = false

	DefaultMPesaEnvironment = "sandbox"
	DefaultMPesaTimeout     = 30 * time.Second

	DefaultFlutterwaveBaseURL = "https://api.flutterwave.com/v3"
	DefaultFlutterwaveTimeout = 30 * time.Second

	DefaultSportradarBaseURL          = "https://api.sportradar.com"
	DefaultSportradarTimeout          = 30 * time.Second
	DefaultSportradarRateLimit        = 100
	DefaultSportradarEnableSoccer     = true
	DefaultSportradarEnableBasketball = true
	DefaultSportradarEnableTennis     = true
	DefaultSportradarEnableCricket    = true

	DefaultSmileIDEnvironment          = "sandbox"
	DefaultSmileIDTimeout              = 30 * time.Second
	DefaultSmileIDEnableKYC            = true
	DefaultSmileIDEnableIDVerification = true

	DefaultTaxEnabled            = true
	DefaultTaxDefaultRate        = 0.16 // 16%
	DefaultTaxWHTRate            = 0.15 // 15%
	DefaultTaxGamingTaxRate      = 0.20 // 20%
	DefaultTaxTransactionTaxRate = 0.00 // 0%
	DefaultTaxCurrency           = "KES"
	DefaultTaxAutoCalculate      = true

	DefaultCrashEnabled       = false
	DefaultCrashTimeout       = 30 * time.Second
	DefaultCrashMaxBetAmount  = 10000.0
	DefaultCrashMinBetAmount  = 10.0
	DefaultCrashMaxMultiplier = 1000.0
	DefaultCrashMinMultiplier = 1.1
	DefaultCrashHouseEdge     = 0.05 // 5%

	DefaultBetMinBetAmount    = 10.0
	DefaultBetMaxBetAmount    = 100000.0
	DefaultBetMaxParlaySize   = 10
	DefaultBetMaxOdds         = 1000.0
	DefaultBetMinOdds         = 1.1
	DefaultBetStakeTimeout    = 30 * time.Second
	DefaultBetSettlementDelay = 5 * time.Minute
	DefaultBetCancelDelay     = 1 * time.Minute

	DefaultLoggingLevel      = "info"
	DefaultLoggingFormat     = "json"
	DefaultLoggingOutput     = "stdout"
	DefaultLoggingFile       = "logs/app.log"
	DefaultLoggingMaxSize    = 100 // MB
	DefaultLoggingMaxBackups = 3
	DefaultLoggingMaxAge     = 28 // days
	DefaultLoggingCompress   = true

	DefaultFeatureEnableLiveBetting       = true
	DefaultFeatureEnableVirtualSports     = true
	DefaultFeatureEnableJackpot           = true
	DefaultFeatureEnablePromotions        = true
	DefaultFeatureEnableNotifications     = true
	DefaultFeatureEnableAnalytics         = true
	DefaultFeatureEnableResponsibleGaming = true
	DefaultFeatureEnableMultiCurrency     = false
	DefaultFeatureEnableMobileApp         = false
	DefaultFeatureEnableAPIV2             = false
	DefaultFeatureEnableBetaFeatures      = false
)
