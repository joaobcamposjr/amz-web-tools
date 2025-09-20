package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Oracle Database
	OracleHost     string
	OraclePort     string
	OracleUser     string
	OraclePassword string
	OracleService  string
	OracleLibDir   string

	// PostgreSQL Database (for integrator)
	PGHost     string
	PGPort     string
	PGUser     string
	PGPassword string
	PGDatabase string
	PGSSLMode  string

	// Server
	ServerPort     string
	JWTSecret      string
	JWTExpireHours int

	// API
	PlateAPIURL string
	PlateAPIKey string

	// Environment
	Environment string

	// CORS
	CORSAllowedOrigins []string

	// Logging
	LogLevel string
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "1433"),
		DBUser:     getEnv("DB_USER", "sa"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "portal"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

		OracleHost:     getEnv("ORACLE_HOST", ""),
		OraclePort:     getEnv("ORACLE_PORT", "1521"),
		OracleUser:     getEnv("ORACLE_USER", ""),
		OraclePassword: getEnv("ORACLE_PASSWORD", ""),
		OracleService:  getEnv("ORACLE_SERVICE", ""),
		OracleLibDir:   getEnv("ORACLE_LIB_DIR", getDefaultOracleLibDir()),

		PGHost:     getEnv("PG_HOST", ""),
		PGPort:     getEnv("PG_PORT", "5433"),
		PGUser:     getEnv("PG_USER", ""),
		PGPassword: getEnv("PG_PASSWORD", ""),
		PGDatabase: getEnv("PG_DATABASE", ""),
		PGSSLMode:  getEnv("PG_SSL_MODE", "disable"),

		ServerPort:     getEnv("SERVER_PORT", "8080"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpireHours: getEnvAsInt("JWT_EXPIRE_HOURS", 168), // 7 dias

		PlateAPIURL: getEnv("PLATE_API_URL", ""),
		PlateAPIKey: getEnv("PLATE_API_KEY", ""),

		Environment: getEnv("ENVIRONMENT", "development"),

		CORSAllowedOrigins: strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ","),

		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDefaultOracleLibDir() string {
	// Check if running on macOS (development)
	if _, err := os.Stat("/Applications/oracle/client/instantclient_23"); err == nil {
		return "/Applications/oracle/client/instantclient_23"
	}
	// Default to Linux path (production)
	return "/opt/oracle/instantclient_21_13"
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
