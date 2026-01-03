package config

import (
	"os"
	"strings"
)

// Config holds the application configuration
type Config struct {
	// Server
	Port        string
	FrontendURL string

	// Best Buy API
	BestBuyAPIKey string
	UseMockData   bool

	// Database
	DatabaseURL string

	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// Security
	SecureCookies bool

	// Initial allowed emails (comma-separated)
	InitialAllowedEmails []string
}

// Load loads the configuration from environment variables
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	apiKey := os.Getenv("BESTBUY_API_KEY")
	useMock := apiKey == ""

	databaseURL := os.Getenv("DATABASE_URL")

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if googleRedirectURL == "" {
		googleRedirectURL = "http://localhost:" + port + "/auth/callback"
	}

	secureCookies := os.Getenv("SECURE_COOKIES") == "true"

	var allowedEmails []string
	if emails := os.Getenv("ALLOWED_EMAILS"); emails != "" {
		for _, email := range strings.Split(emails, ",") {
			email = strings.TrimSpace(email)
			if email != "" {
				allowedEmails = append(allowedEmails, email)
			}
		}
	}

	return &Config{
		Port:                 port,
		FrontendURL:          frontendURL,
		BestBuyAPIKey:        apiKey,
		UseMockData:          useMock,
		DatabaseURL:          databaseURL,
		GoogleClientID:       googleClientID,
		GoogleClientSecret:   googleClientSecret,
		GoogleRedirectURL:    googleRedirectURL,
		SecureCookies:        secureCookies,
		InitialAllowedEmails: allowedEmails,
	}
}

// HasAuth returns true if OAuth is configured
func (c *Config) HasAuth() bool {
	return c.GoogleClientID != "" && c.GoogleClientSecret != ""
}

// HasDatabase returns true if database is configured
func (c *Config) HasDatabase() bool {
	return c.DatabaseURL != ""
}
