package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type NATSConfig struct {
	URL string
}

type TenantConfig struct {
	CountryCode      string // primary launch country: KE, NG, GH
	Currency         string // KES, NGN, GHS
	Timezone         string
	AllowedCountries []string // ISO codes allowed by the geolocation fence (empty = allow all)
}

type JWTConfig struct {
	Secret       string
	Issuer       string
	ExpiryHours  int
	RefreshHours int
}

type SecurityConfig struct {
	BcryptCost         int
	RateLimitRequests  int
	RateLimitWindow    time.Duration
	CORSAllowedOrigins []string
	CORSAllowedMethods []string
	CORSAllowedHeaders []string
}

type MPesaConfig struct {
	Environment        string
	ConsumerKey        string
	ConsumerSecret     string
	ShortCode          string
	PassKey            string
	InitiatorName      string
	SecurityCredential string
	CallbackBaseURL    string
}

type FlutterwaveConfig struct {
	SecretKey string
	PublicKey string
}

type SportradarConfig struct {
	APIKey string
	APIURL string
}

type SmileIDConfig struct {
	APIKey    string
	PartnerID string
	Env       string // sandbox or production
}

type TaxConfig struct {
	GGRRate   float64
	WHTRate   float64
	Threshold float64
}

type CrashConfig struct {
	TickInterval  time.Duration
	MinBet        float64
	MaxBet        float64
	MaxMultiplier float64
	HouseEdge     float64
}

type BetConfig struct {
	MinStake           float64
	MaxStake           float64
	MaxSelectionsMulti int
}

type LoggingConfig struct {
	Level  string
	Format string // json or text
}

type FeatureFlags struct {
	LiveBetting   bool
	CrashGames    bool
	VirtualSports bool
	Casino        bool
	DebugMode     bool
	EnableSwagger bool
	EnablePprof   bool
}

// Load reads configuration from environment variables.
// A service name can be provided to scope per-service port env lookup.
func Load(serviceName string) (*Config, error) {
	cfg := &Config{
		Service: ServiceConfig{
			Name:        getString("SERVICE_NAME", serviceName),
			Environment: getString("ENVIRONMENT", "development"),
			Port:        getInt("PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:            getString("DATABASE_HOST", "localhost"),
			Port:            getInt("DATABASE_PORT", 5432),
			Name:            getString("DATABASE_NAME", "betting_db"),
			User:            getString("DATABASE_USER", "postgres"),
			Password:        getString("DATABASE_PASSWORD", "postgres"),
			SSLMode:         getString("DATABASE_SSL_MODE", "disable"),
			MaxOpenConns:    getInt("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt("DATABASE_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDuration("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getString("REDIS_HOST", "localhost"),
			Port:     getInt("REDIS_PORT", 6379),
			Password: getString("REDIS_PASSWORD", ""),
			DB:       getInt("REDIS_DB", 0),
		},
		NATS: NATSConfig{
			URL: getString("NATS_URL", "nats://localhost:4222"),
		},
		Tenant: TenantConfig{
			CountryCode:      getString("COUNTRY_CODE", "KE"),
			Currency:         getString("CURRENCY", "KES"),
			Timezone:         getString("TIMEZONE", "Africa/Nairobi"),
			AllowedCountries: getStringSlice("ALLOWED_COUNTRIES", nil),
		},
		JWT: JWTConfig{
			Secret:       getString("JWT_SECRET", "change-me-in-production"),
			Issuer:       getString("JWT_ISSUER", "betting-platform"),
			ExpiryHours:  getInt("JWT_EXPIRY_HOURS", 24),
			RefreshHours: getInt("JWT_REFRESH_EXPIRY_HOURS", 168),
		},
		Security: SecurityConfig{
			BcryptCost:         getInt("BCRYPT_COST", 12),
			RateLimitRequests:  getInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow:    getDuration("RATE_LIMIT_WINDOW", time.Minute),
			CORSAllowedOrigins: getStringSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			CORSAllowedMethods: getStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			CORSAllowedHeaders: getStringSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
		},
		MPesa: MPesaConfig{
			Environment:        getString("MPESA_ENVIRONMENT", "sandbox"),
			ConsumerKey:        getString("MPESA_CONSUMER_KEY", ""),
			ConsumerSecret:     getString("MPESA_CONSUMER_SECRET", ""),
			ShortCode:          getString("MPESA_SHORTCODE", ""),
			PassKey:            getString("MPESA_PASSKEY", ""),
			InitiatorName:      getString("MPESA_INITIATOR_NAME", ""),
			SecurityCredential: getString("MPESA_SECURITY_CREDENTIAL", ""),
			CallbackBaseURL:    getString("MPESA_CALLBACK_BASE_URL", "http://localhost:8080"),
		},
		Flutterwave: FlutterwaveConfig{
			SecretKey: getString("FLUTTERWAVE_SECRET_KEY", ""),
			PublicKey: getString("FLUTTERWAVE_PUBLIC_KEY", ""),
		},
		Sportradar: SportradarConfig{
			APIKey: getString("SPORTRADAR_API_KEY", ""),
			APIURL: getString("SPORTRADAR_API_URL", "https://api.sportradar.com"),
		},
		SmileID: SmileIDConfig{
			APIKey:    getString("SMILE_ID_API_KEY", ""),
			PartnerID: getString("SMILE_ID_PARTNER_ID", ""),
			Env:       getString("SMILE_ID_ENV", "sandbox"),
		},
		Tax: TaxConfig{
			GGRRate:   getFloat("TAX_GGR_RATE", 0.15),
			WHTRate:   getFloat("TAX_WHT_RATE", 0.20),
			Threshold: getFloat("TAX_THRESHOLD", 500),
		},
		Crash: CrashConfig{
			TickInterval:  getDuration("CRASH_TICK_INTERVAL", 100*time.Millisecond),
			MinBet:        getFloat("CRASH_MIN_BET", 10),
			MaxBet:        getFloat("CRASH_MAX_BET", 10000),
			MaxMultiplier: getFloat("CRASH_MAX_MULTIPLIER", 100),
			HouseEdge:     getFloat("CRASH_HOUSE_EDGE", 0.01),
		},
		Bet: BetConfig{
			MinStake:           getFloat("MIN_BET_STAKE", 10),
			MaxStake:           getFloat("MAX_BET_STAKE", 100000),
			MaxSelectionsMulti: getInt("MAX_SELECTIONS_MULTI", 20),
		},
		Logging: LoggingConfig{
			Level:  getString("LOG_LEVEL", "info"),
			Format: getString("LOG_FORMAT", "json"),
		},
		Features: FeatureFlags{
			LiveBetting:   getBool("FEATURE_LIVE_BETTING", true),
			CrashGames:    getBool("FEATURE_CRASH_GAMES", true),
			VirtualSports: getBool("FEATURE_VIRTUAL_SPORTS", false),
			Casino:        getBool("FEATURE_CASINO", false),
			DebugMode:     getBool("DEBUG_MODE", false),
			EnableSwagger: getBool("ENABLE_SWAGGER", true),
			EnablePprof:   getBool("ENABLE_PPROF", false),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate ensures required configuration values are set for the chosen environment.
func (c *Config) Validate() error {
	if c.Service.Environment == "production" {
		if c.JWT.Secret == "" || c.JWT.Secret == "change-me-in-production" {
			return fmt.Errorf("JWT_SECRET must be set in production")
		}
		if c.Database.Password == "" {
			return fmt.Errorf("DATABASE_PASSWORD must be set in production")
		}
	}
	if c.Service.Port <= 0 || c.Service.Port > 65535 {
		return fmt.Errorf("invalid PORT: %d", c.Service.Port)
	}
	if c.Tenant.CountryCode == "" {
		return fmt.Errorf("COUNTRY_CODE is required")
	}
	return nil
}

// IsProduction returns true when environment is production.
func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.Service.Environment, "production")
}

// --- helpers ---

func getString(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getFloat(key string, def float64) float64 {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getStringSlice(key string, def []string) []string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		parts := strings.Split(v, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	return def
}
