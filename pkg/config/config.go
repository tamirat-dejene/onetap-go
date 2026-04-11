package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	ServerPort              string
	JWTSecret               string
	JWTExpiryHours          int
	AdminUsername           string
	AdminPassword           string
	UploaderUsername        string
	UploaderPassword        string
	BcryptCost              int
	RateLimitRequests       int
	RateLimitWindowSeconds  int
	CustomersFile           string
	TransactionsFile        string
	SampleCustomersFile     string
}

// Load reads the .env file (if present) and populates a Config struct.
// Values already set in the environment take precedence.
func Load(envFile string) (*Config, error) {
	// Load .env file; ignore error if file not found (env vars may be set directly)
	_ = godotenv.Load(envFile)

	cfg := &Config{}

	cfg.ServerPort = getEnv("SERVER_PORT", "8080")
	cfg.JWTSecret = requireEnv("JWT_SECRET")
	cfg.AdminUsername = requireEnv("ADMIN_USERNAME")
	cfg.AdminPassword = requireEnv("ADMIN_PASSWORD")
	cfg.UploaderUsername = requireEnv("UPLOADER_USERNAME")
	cfg.UploaderPassword = requireEnv("UPLOADER_PASSWORD")
	cfg.CustomersFile = getEnv("CUSTOMERS_FILE", "inputs/customers.json")
	cfg.TransactionsFile = getEnv("TRANSACTIONS_FILE", "inputs/transactions.json")
	cfg.SampleCustomersFile = getEnv("SAMPLE_CUSTOMERS_FILE", "inputs/sample_customers.csv")

	var err error
	if cfg.JWTExpiryHours, err = getEnvInt("JWT_EXPIRY_HOURS", 24); err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY_HOURS: %w", err)
	}
	if cfg.BcryptCost, err = getEnvInt("BCRYPT_COST", 12); err != nil {
		return nil, fmt.Errorf("invalid BCRYPT_COST: %w", err)
	}
	if cfg.RateLimitRequests, err = getEnvInt("RATE_LIMIT_REQUESTS", 5); err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_REQUESTS: %w", err)
	}
	if cfg.RateLimitWindowSeconds, err = getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60); err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_WINDOW_SECONDS: %w", err)
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func getEnvInt(key string, defaultVal int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal, nil
	}
	return strconv.Atoi(v)
}
