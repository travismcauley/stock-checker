package main

import (
	"context"
	"log"
	"net/http"
	"path/filepath"

	"connectrpc.com/connect"
	"github.com/tmcauley/stock-checker/backend/gen/stockchecker/v1/stockcheckerv1connect"
	"github.com/tmcauley/stock-checker/backend/internal/auth"
	"github.com/tmcauley/stock-checker/backend/internal/bestbuy"
	"github.com/tmcauley/stock-checker/backend/internal/config"
	"github.com/tmcauley/stock-checker/backend/internal/database"
	"github.com/tmcauley/stock-checker/backend/internal/handler"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create Best Buy API client (mock or real based on config)
	var bbClient bestbuy.Client
	if cfg.UseMockData {
		log.Println("Using mock Best Buy API client (no API key provided)")
		bbClient = bestbuy.NewMockClient()
	} else {
		log.Println("Using real Best Buy API client")
		bbClient = bestbuy.NewAPIClient(cfg.BestBuyAPIKey)
	}

	// Database connection (optional for local development)
	var db *database.DB
	var authHandler *auth.Auth

	if cfg.HasDatabase() {
		var err error
		db, err = database.New(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		// Run migrations
		migrationsDir := filepath.Join("migrations")
		if err := db.RunMigrations(migrationsDir); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

		// Seed initial allowed emails
		for _, email := range cfg.InitialAllowedEmails {
			if err := db.AddAllowedEmail(context.Background(), email, nil); err != nil {
				log.Printf("Warning: failed to add allowed email %s: %v", email, err)
			} else {
				log.Printf("Added allowed email: %s", email)
			}
		}

		log.Println("Database connected and migrated")
	} else {
		log.Println("Running without database (localStorage mode)")
	}

	// Auth handler (optional)
	if cfg.HasAuth() && db != nil {
		authHandler = auth.New(
			db,
			cfg.GoogleClientID,
			cfg.GoogleClientSecret,
			cfg.GoogleRedirectURL,
			cfg.FrontendURL,
			cfg.SecureCookies,
		)
		log.Println("Google OAuth enabled")
	} else {
		log.Println("Running without authentication")
	}

	// Create the handler
	stockCheckerHandler := handler.NewStockCheckerHandler(bbClient, db)

	// Create the Connect service path and handler
	path, connectHandler := stockcheckerv1connect.NewStockCheckerServiceHandler(
		stockCheckerHandler,
		connect.WithInterceptors(),
	)

	// Create a new mux and register the handler
	mux := http.NewServeMux()

	// Auth endpoints (if auth is configured)
	if authHandler != nil {
		mux.HandleFunc("/auth/login", authHandler.HandleLogin)
		mux.HandleFunc("/auth/callback", authHandler.HandleCallback)
		mux.HandleFunc("/auth/logout", authHandler.HandleLogout)

		// Wrap Connect handler with auth middleware for protected endpoints
		mux.Handle(path, authHandler.Middleware(connectHandler))
	} else {
		mux.Handle(path, connectHandler)
	}

	// Add CORS middleware
	corsHandler := corsMiddleware(mux, cfg.FrontendURL)

	log.Printf("Starting server on :%s", cfg.Port)
	log.Printf("StockCheckerService available at http://localhost:%s%s", cfg.Port, path)
	if authHandler != nil {
		log.Printf("Auth endpoints: /auth/login, /auth/callback, /auth/logout")
	}

	// Use h2c for HTTP/2 without TLS (needed for Connect)
	err := http.ListenAndServe(
		":"+cfg.Port,
		h2c.NewHandler(corsHandler, &http2.Server{}),
	)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler, frontendURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = frontendURL
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Cookie")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Expose-Headers", "Connect-Protocol-Version")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
