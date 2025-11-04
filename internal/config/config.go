package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress     string
	DatabaseURI    string
	AccrualAddress string
	JWTSecret      string
}

func Load(args []string) *Config {
	cfg := &Config{}

	flagSet := flag.NewFlagSet("gophermart", flag.ContinueOnError)
	flagSet.StringVar(&cfg.RunAddress, "a", ":8080", "Server address")
	flagSet.StringVar(&cfg.DatabaseURI, "d", "", "Database URI")
	flagSet.StringVar(&cfg.AccrualAddress, "r", "", "Accrual system address")

	// Parse command line arguments
	flagSet.Parse(args)

	// Override with environment variables
	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		cfg.RunAddress = envRunAddr
	}
	if envDBURI := os.Getenv("DATABASE_URI"); envDBURI != "" {
		cfg.DatabaseURI = envDBURI
	}
	if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
		cfg.AccrualAddress = envAccrualAddr
	}

	// JWT secret from environment with fallback
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWTSecret = secret
	} else {
		// in case of empty secret it must throw a panic
		panic("JWT secret is required")
	}

	return cfg
}
