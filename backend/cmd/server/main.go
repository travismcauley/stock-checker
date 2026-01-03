package main

import (
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/tmcauley/stock-checker/backend/gen/stockchecker/v1/stockcheckerv1connect"
	"github.com/tmcauley/stock-checker/backend/internal/bestbuy"
	"github.com/tmcauley/stock-checker/backend/internal/config"
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

	// Create the handler
	stockCheckerHandler := handler.NewStockCheckerHandler(bbClient)

	// Create the Connect service path and handler
	path, connectHandler := stockcheckerv1connect.NewStockCheckerServiceHandler(
		stockCheckerHandler,
		connect.WithInterceptors(),
	)

	// Create a new mux and register the handler
	mux := http.NewServeMux()
	mux.Handle(path, connectHandler)

	// Add CORS middleware for local development
	corsHandler := corsMiddleware(mux)

	log.Printf("Starting server on :%s", cfg.Port)
	log.Printf("StockCheckerService available at http://localhost:%s%s", cfg.Port, path)

	// Use h2c for HTTP/2 without TLS (needed for Connect)
	err := http.ListenAndServe(
		":"+cfg.Port,
		h2c.NewHandler(corsHandler, &http2.Server{}),
	)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// corsMiddleware adds CORS headers for local development
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from the Vite dev server
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:5173"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version")
		w.Header().Set("Access-Control-Expose-Headers", "Connect-Protocol-Version")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
