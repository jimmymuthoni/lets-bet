package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		Service: ServiceConfig{
			Name:        getEnv(EnvServiceName, DefaultServiceName),
			Environment: getEnv(EnvServiceEnvironment, DefaultServiceEnvironment),
			Port:        getEnvInt(EnvServicePort, DefaultServicePort),
		},
		Database: DatabaseConfig{
			Host:            getEnv(EnvDBHost, DefaultDBHost),
			Port:            getEnvInt(EnvDBPort, DefaultDBPort),
			Name:            getEnv(EnvDBName, DefaultDBName),
			User:            getEnv(EnvDBUser, DefaultDBUser),
			Password:        getEnv(EnvDBPassword, ""),
			SSLMode:         getEnv(EnvDBSSLMode, DefaultDBSSLMode),
			MaxOpenConns:    getEnvInt(EnvDBMaxOpenConns, DefaultDBMaxOpenConns),
			MaxIdleConns:    getEnvInt(EnvDBMaxIdleConns, DefaultDBMaxIdleConns),
			ConnMaxLifetime: getEnvDuration(EnvDBConnMaxLifetime, DefaultDBConnMaxLifetime),
		},
		Redis: RedisConfig{
			Host:         getEnv(EnvRedisHost, DefaultRedisHost),
			Port:         getEnvInt(EnvRedisPort, DefaultRedisPort),
			Password:     getEnv(EnvRedisPassword, ""),
			DB:           getEnvInt(EnvRedisDB, DefaultRedisDB),
			PoolSize:     getEnvInt(EnvRedisPoolSize, DefaultRedisPoolSize),
			MinIdleConns: getEnvInt(EnvRedisMinIdleConns, DefaultRedisMinIdleConns),
			DialTimeout:  getEnvDuration(EnvRedisDialTimeout, DefaultRedisDialTimeout),
			ReadTimeout:  getEnvDuration(EnvRedisReadTimeout, DefaultRedisReadTimeout),
			WriteTimeout: getEnvDuration(EnvRedisWriteTimeout, DefaultRedisWriteTimeout),
		},
		NATS: NATSConfig{
			URL:           getEnv(EnvNATSURL, DefaultNATSURL),
			MaxReconnects: getEnvInt(EnvNATSMaxReconnects, DefaultNATSMaxReconnects),
			ReconnectWait: getEnvDuration(EnvNATSReconnectWait, DefaultNATSReconnectWait),
			Timeout:       getEnvDuration(EnvNATSTimeout, DefaultNATSTimeout),
			PingInterval:  getEnvDuration(EnvNATSPingInterval, DefaultNATSPingInterval),
			MaxPingsOut:   getEnvInt(EnvNATSMaxPingsOut, DefaultNATSMaxPingsOut),
		},
		Tenant: TenantConfig{
			DefaultCountry:      getEnv(EnvTenantDefaultCountry, DefaultTenantDefaultCountry),
			DefaultCurrency:     getEnv(EnvTenantDefaultCurrency, DefaultTenantDefaultCurrency),
			SupportedCountries:  strings.Split(getEnv("TENANT_SUPPORTED_COUNTRIES", "KE,UG,TZ,NG,ZA"), ","),
			SupportedCurrencies: strings.Split(getEnv("TENANT_SUPPORTED_CURRENCIES", "KES,UGX,TZS,NGN,ZAR"), ","),
			AllowedCountries:    strings.Split(getEnv("TENANT_ALLOWED_COUNTRIES", "KE,UG,TZ,NG,ZA"), ","),
		},
		JWT: JWTConfig{
			Secret:         getEnv(EnvJWTSecret, ""),
			ExpirationTime: getEnvDuration(EnvJWTExpirationTime, DefaultJWTExpirationTime),
			Issuer:         getEnv(EnvJWTIssuer, DefaultJWTIssuer),
			RefreshTime:    getEnvDuration(EnvJWTRefreshTime, DefaultJWTRefreshTime),
		},
		Security: SecurityConfig{
			CORSOrigins:           strings.Split(getEnv(EnvSecurityCORSOrigins, "*"), ","),
			CORSAllowedOrigins:    strings.Split(getEnv(EnvSecurityCORSAllowedOrigins, "*"), ","),
			CORSAllowedMethods:    strings.Split(getEnv(EnvSecurityCORSAllowedMethods, "GET,POST,PUT,DELETE,OPTIONS"), ","),
			CORSAllowedHeaders:    strings.Split(getEnv(EnvSecurityCORSAllowedHeaders, "Content-Type,Authorization"), ","),
			RateLimitPerMinute:    getEnvInt(EnvSecurityRateLimitPerMinute, DefaultSecurityRateLimitPerMinute),
			RateLimitRequests:     getEnvInt(EnvSecurityRateLimitRequests, 1000),
			RateLimitWindow:       getEnvDuration(EnvSecurityRateLimitWindow, time.Minute),
			MaxBodySize:           getEnvInt64(EnvSecurityMaxBodySize, DefaultSecurityMaxBodySize),
			EnableHTTPS:           getEnvBool(EnvSecurityEnableHTTPS, DefaultSecurityEnableHTTPS),
			SessionTimeout:        getEnvDuration(EnvSecuritySessionTimeout, DefaultSecuritySessionTimeout),
			PasswordMinLength:     getEnvInt(EnvSecurityPasswordMinLength, DefaultSecurityPasswordMinLength),
			PasswordRequireUpper:  getEnvBool(EnvSecurityPasswordRequireUpper, DefaultSecurityPasswordRequireUpper),
			PasswordRequireLower:  getEnvBool(EnvSecurityPasswordRequireLower, DefaultSecurityPasswordRequireLower),
			PasswordRequireNumber: getEnvBool(EnvSecurityPasswordRequireNumber, DefaultSecurityPasswordRequireNumber),
			PasswordRequireSymbol: getEnvBool(EnvSecurityPasswordRequireSymbol, DefaultSecurityPasswordRequireSymbol),
		},
		MPesa: MPesaConfig{
			ConsumerKey:        getEnv(EnvMPesaConsumerKey, ""),
			ConsumerSecret:     getEnv(EnvMPesaConsumerSecret, ""),
			ShortCode:          getEnv(EnvMPesaShortCode, ""),
			PassKey:            getEnv(EnvMPesaPassKey, ""),
			InitiatorName:      getEnv(EnvMPesaInitiatorName, ""),
			SecurityCredential: getEnv(EnvMPesaSecurityCredential, ""),
			Environment:        getEnv(EnvMPesaEnvironment, DefaultMPesaEnvironment),
			CallbackURL:        getEnv(EnvMPesaCallbackURL, ""),
			Timeout:            getEnvDuration(EnvMPesaTimeout, DefaultMPesaTimeout),
		},
		Flutterwave: FlutterwaveConfig{
			PublicKey:     getEnv(EnvFlutterwavePublicKey, ""),
			SecretKey:     getEnv(EnvFlutterwaveSecretKey, ""),
			EncryptionKey: getEnv(EnvFlutterwaveEncryptionKey, ""),
			BaseURL:       getEnv(EnvFlutterwaveBaseURL, DefaultFlutterwaveBaseURL),
			WebhookSecret: getEnv(EnvFlutterwaveWebhookSecret, ""),
			Timeout:       getEnvDuration(EnvFlutterwaveTimeout, DefaultFlutterwaveTimeout),
		},
		Sportradar: SportradarConfig{
			APIKey:           getEnv(EnvSportradarAPIKey, ""),
			BaseURL:          getEnv(EnvSportradarBaseURL, DefaultSportradarBaseURL),
			Timeout:          getEnvDuration(EnvSportradarTimeout, DefaultSportradarTimeout),
			RateLimit:        getEnvInt(EnvSportradarRateLimit, DefaultSportradarRateLimit),
			EnableSoccer:     getEnvBool(EnvSportradarEnableSoccer, DefaultSportradarEnableSoccer),
			EnableBasketball: getEnvBool(EnvSportradarEnableBasketball, DefaultSportradarEnableBasketball),
			EnableTennis:     getEnvBool(EnvSportradarEnableTennis, DefaultSportradarEnableTennis),
			EnableCricket:    getEnvBool(EnvSportradarEnableCricket, DefaultSportradarEnableCricket),
		},
		SmileID: SmileIDConfig{
			PartnerID:            getEnv(EnvSmileIDPartnerID, ""),
			APIKey:               getEnv(EnvSmileIDAPIKey, ""),
			BaseURL:              getEnv(EnvSmileIDBaseURL, ""),
			Environment:          getEnv(EnvSmileIDEnvironment, DefaultSmileIDEnvironment),
			Timeout:              getEnvDuration(EnvSmileIDTimeout, DefaultSmileIDTimeout),
			EnableKYC:            getEnvBool(EnvSmileIDEnableKYC, DefaultSmileIDEnableKYC),
			EnableIDVerification: getEnvBool(EnvSmileIDEnableIDVerification, DefaultSmileIDEnableIDVerification),
		},
		Tax: TaxConfig{
			Enabled:            getEnvBool(EnvTaxEnabled, DefaultTaxEnabled),
			DefaultRate:        getEnvFloat64(EnvTaxDefaultRate, DefaultTaxDefaultRate),
			WHTRate:            getEnvFloat64(EnvTaxWHTRate, DefaultTaxWHTRate),
			GamingTaxRate:      getEnvFloat64(EnvTaxGamingTaxRate, DefaultTaxGamingTaxRate),
			TransactionTaxRate: getEnvFloat64(EnvTaxTransactionTaxRate, DefaultTaxTransactionTaxRate),
			Currency:           getEnv(EnvTaxCurrency, DefaultTaxCurrency),
			AutoCalculate:      getEnvBool(EnvTaxAutoCalculate, DefaultTaxAutoCalculate),
		},
		Crash: CrashConfig{
			Enabled:       getEnvBool(EnvCrashEnabled, DefaultCrashEnabled),
			BaseURL:       getEnv(EnvCrashBaseURL, ""),
			APIKey:        getEnv(EnvCrashAPIKey, ""),
			SecretKey:     getEnv(EnvCrashSecretKey, ""),
			Timeout:       getEnvDuration(EnvCrashTimeout, DefaultCrashTimeout),
			MaxBetAmount:  getEnvFloat64(EnvCrashMaxBetAmount, DefaultCrashMaxBetAmount),
			MinBetAmount:  getEnvFloat64(EnvCrashMinBetAmount, DefaultCrashMinBetAmount),
			MaxMultiplier: getEnvFloat64(EnvCrashMaxMultiplier, DefaultCrashMaxMultiplier),
			MinMultiplier: getEnvFloat64(EnvCrashMinMultiplier, DefaultCrashMinMultiplier),
			HouseEdge:     getEnvFloat64(EnvCrashHouseEdge, DefaultCrashHouseEdge),
		},
		Bet: BetConfig{
			MinBetAmount:    getEnvFloat64(EnvBetMinBetAmount, DefaultBetMinBetAmount),
			MaxBetAmount:    getEnvFloat64(EnvBetMaxBetAmount, DefaultBetMaxBetAmount),
			MaxParlaySize:   getEnvInt(EnvBetMaxParlaySize, DefaultBetMaxParlaySize),
			MaxOdds:         getEnvFloat64(EnvBetMaxOdds, DefaultBetMaxOdds),
			MinOdds:         getEnvFloat64(EnvBetMinOdds, DefaultBetMinOdds),
			StakeTimeout:    getEnvDuration(EnvBetStakeTimeout, DefaultBetStakeTimeout),
			SettlementDelay: getEnvDuration(EnvBetSettlementDelay, DefaultBetSettlementDelay),
			CancelDelay:     getEnvDuration(EnvBetCancelDelay, DefaultBetCancelDelay),
		},
		Logging: LoggingConfig{
			Level:      getEnv(EnvLoggingLevel, DefaultLoggingLevel),
			Format:     getEnv(EnvLoggingFormat, DefaultLoggingFormat),
			Output:     getEnv(EnvLoggingOutput, DefaultLoggingOutput),
			File:       getEnv(EnvLoggingFile, DefaultLoggingFile),
			MaxSize:    getEnvInt(EnvLoggingMaxSize, DefaultLoggingMaxSize),
			MaxBackups: getEnvInt(EnvLoggingMaxBackups, DefaultLoggingMaxBackups),
			MaxAge:     getEnvInt(EnvLoggingMaxAge, DefaultLoggingMaxAge),
			Compress:   getEnvBool(EnvLoggingCompress, DefaultLoggingCompress),
		},
		Features: FeatureFlags{
			EnableLiveBetting:       getEnvBool(EnvFeatureEnableLiveBetting, DefaultFeatureEnableLiveBetting),
			EnableVirtualSports:     getEnvBool(EnvFeatureEnableVirtualSports, DefaultFeatureEnableVirtualSports),
			EnableJackpot:           getEnvBool(EnvFeatureEnableJackpot, DefaultFeatureEnableJackpot),
			EnablePromotions:        getEnvBool(EnvFeatureEnablePromotions, DefaultFeatureEnablePromotions),
			EnableNotifications:     getEnvBool(EnvFeatureEnableNotifications, DefaultFeatureEnableNotifications),
			EnableAnalytics:         getEnvBool(EnvFeatureEnableAnalytics, DefaultFeatureEnableAnalytics),
			EnableResponsibleGaming: getEnvBool(EnvFeatureEnableResponsibleGaming, DefaultFeatureEnableResponsibleGaming),
			EnableMultiCurrency:     getEnvBool(EnvFeatureEnableMultiCurrency, DefaultFeatureEnableMultiCurrency),
			EnableMobileApp:         getEnvBool(EnvFeatureEnableMobileApp, DefaultFeatureEnableMobileApp),
			EnableAPIV2:             getEnvBool(EnvFeatureEnableAPIV2, DefaultFeatureEnableAPIV2),
			EnableBetaFeatures:      getEnvBool(EnvFeatureEnableBetaFeatures, DefaultFeatureEnableBetaFeatures),
		},
	}

	return config, nil
}

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) error {
	if config.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}

	if config.Service.Port <= 0 || config.Service.Port > 65535 {
		return fmt.Errorf("invalid service port: %d", config.Service.Port)
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Database.Port <= 0 || config.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", config.Database.Port)
	}

	if config.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if config.JWT.ExpirationTime <= 0 {
		return fmt.Errorf("JWT expiration time must be positive")
	}

	if config.Security.PasswordMinLength < 6 {
		return fmt.Errorf("password minimum length must be at least 6")
	}

	return nil
}

// LoadFromFile loads configuration from a JSON file
func LoadFromFile(filename string) (*Config, error) {
	// This would typically use a JSON library to load from file
	// For now, we'll return an error as this isn't fully implemented
	return nil, fmt.Errorf("loading from file not implemented")
}

// GetEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// GetEnvInt gets environment variable as integer with fallback
func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// GetEnvInt64 gets environment variable as int64 with fallback
func getEnvInt64(key string, fallback int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return fallback
}

// GetEnvFloat64 gets environment variable as float64 with fallback
func getEnvFloat64(key string, fallback float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return fallback
}

// GetEnvBool gets environment variable as boolean with fallback
func getEnvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

// GetEnvDuration gets environment variable as duration with fallback
func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}
