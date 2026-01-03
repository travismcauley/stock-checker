package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	Port          string
	BestBuyAPIKey string
	UseMockData   bool
}

// Load loads the configuration from environment variables
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	apiKey := os.Getenv("BESTBUY_API_KEY")

	// Use mock data if no API key is provided
	useMock := apiKey == ""

	return &Config{
		Port:          port,
		BestBuyAPIKey: apiKey,
		UseMockData:   useMock,
	}
}
