package bestbuy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Known category IDs for Best Buy
const (
	CategoryTradingCards = "pcmcat1604992984556" // Trading Cards category
)

// Client is the interface for Best Buy API operations
type Client interface {
	// SearchStores searches for stores near a postal code within a radius
	SearchStores(ctx context.Context, postalCode string, radiusMiles int) ([]Store, error)

	// SearchProducts searches for products by keyword, optionally filtered by subclass
	SearchProducts(ctx context.Context, query string, subclass string) ([]Product, error)

	// SearchProductsInCategory searches for products within a category
	SearchProductsInCategory(ctx context.Context, categoryID string, query string) ([]Product, error)

	// GetProductBySKU gets a single product by its SKU
	GetProductBySKU(ctx context.Context, sku string) (*Product, error)

	// CheckAvailability checks product availability at specific stores
	CheckAvailability(ctx context.Context, sku string, storeIDs []string) ([]StoreAvailability, error)

	// BrowsePokemonProducts returns Pokemon TCG products from the trading cards category
	BrowsePokemonProducts(ctx context.Context) ([]Product, error)
}

// Store represents a Best Buy store from the API
type Store struct {
	StoreID    int     `json:"storeId"`
	Name       string  `json:"name"`
	Address    string  `json:"address"`
	Address2   string  `json:"address2"`
	City       string  `json:"city"`
	State      string  `json:"region"` // Best Buy uses "region" for state
	PostalCode string  `json:"postalCode"`
	Phone      string  `json:"phone"`
	Distance   float64 `json:"distance"`
	StoreType  string  `json:"storeType"`
	Hours      string  `json:"hours"`
	HoursAmPm  string  `json:"hoursAmPm"`
	GMTOffset  int     `json:"gmtOffset"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
}

// StoreIDString returns the store ID as a string
func (s Store) StoreIDString() string {
	return fmt.Sprintf("%d", s.StoreID)
}

// Product represents a Best Buy product from the API
type Product struct {
	SKU                 int     `json:"sku"`
	Name                string  `json:"name"`
	SalePrice           float64 `json:"salePrice"`
	RegularPrice        float64 `json:"regularPrice"`
	ThumbnailImage      string  `json:"thumbnailImage"`
	Image               string  `json:"image"`
	URL                 string  `json:"url"`
	ShortDescription    string  `json:"shortDescription"`
	LongDescription     string  `json:"longDescription"`
	Manufacturer        string  `json:"manufacturer"`
	ModelNumber         string  `json:"modelNumber"`
	UPC                 string  `json:"upc"`
	InStoreAvailability bool    `json:"inStoreAvailability"`
	OnlineAvailability  bool    `json:"onlineAvailability"`
}

// SKUString returns the SKU as a string
func (p Product) SKUString() string {
	return fmt.Sprintf("%d", p.SKU)
}

// StoreAvailability represents product availability at a store
type StoreAvailability struct {
	StoreID        string  `json:"storeId"`
	StoreName      string  `json:"storeName"`
	City           string  `json:"city"`
	State          string  `json:"state"`
	Distance       float64 `json:"distance"`
	InStock        bool    `json:"inStock"`
	LowStock       bool    `json:"lowStock"`
	PickupEligible bool    `json:"pickupEligible"`
}

// RateLimitError is returned when the API rate limit is exceeded
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %v", e.RetryAfter)
}

// APIClient is the real Best Buy API client implementation
type APIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client

	// Rate limiting
	mu            sync.Mutex
	lastRequest   time.Time
	minInterval   time.Duration // Minimum time between requests
	maxRetries    int
	retryBaseWait time.Duration
}

// NewAPIClient creates a new Best Buy API client
func NewAPIClient(apiKey string) *APIClient {
	return &APIClient{
		apiKey:  apiKey,
		baseURL: "https://api.bestbuy.com/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		minInterval:   350 * time.Millisecond, // ~3 requests per second (safer for Best Buy's rate limits)
		maxRetries:    5,
		retryBaseWait: 1 * time.Second,
	}
}

// doRequest performs an HTTP request with rate limiting and retry logic
func (c *APIClient) doRequest(ctx context.Context, endpoint string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Rate limiting - ensure minimum interval between requests
		c.mu.Lock()
		elapsed := time.Since(c.lastRequest)
		if elapsed < c.minInterval {
			sleepTime := c.minInterval - elapsed
			c.mu.Unlock()
			select {
			case <-time.After(sleepTime):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			c.mu.Lock()
		}
		c.lastRequest = time.Now()
		c.mu.Unlock()

		// Create and execute request
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to execute request: %w", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		// Handle rate limiting (429 Too Many Requests or 403 with rate limit message)
		isRateLimited := resp.StatusCode == http.StatusTooManyRequests ||
			(resp.StatusCode == http.StatusForbidden && strings.Contains(string(body), "per second limit"))

		if isRateLimited {
			retryAfter := c.retryBaseWait * time.Duration(1<<attempt) // Exponential backoff

			// Check for Retry-After header
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if seconds, err := time.ParseDuration(ra + "s"); err == nil {
					retryAfter = seconds
				}
			}

			log.Printf("Rate limited, waiting %v before retry (attempt %d/%d)", retryAfter, attempt+1, c.maxRetries)
			lastErr = &RateLimitError{RetryAfter: retryAfter}

			select {
			case <-time.After(retryAfter):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// Handle other errors
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))

			// Don't retry on client errors (except rate limiting handled above)
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return nil, lastErr
			}

			// Retry on server errors with backoff
			select {
			case <-time.After(c.retryBaseWait * time.Duration(1<<attempt)):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		return body, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// storesResponse is the API response for store searches
type storesResponse struct {
	Stores []Store `json:"stores"`
	Total  int     `json:"total"`
}

// productsResponse is the API response for product searches
type productsResponse struct {
	Products []Product `json:"products"`
	Total    int       `json:"total"`
}

// availabilityResponse is the API response for availability checks
type availabilityResponse struct {
	Stores []struct {
		StoreID   string  `json:"storeId"`
		StoreName string  `json:"name"`
		City      string  `json:"city"`
		State     string  `json:"state"`
		Distance  float64 `json:"distance"`
		LowStock  bool    `json:"lowStock"`
	} `json:"stores"`
}

// SearchStores searches for stores near a postal code
func (c *APIClient) SearchStores(ctx context.Context, postalCode string, radiusMiles int) ([]Store, error) {
	log.Printf("SearchStores called with postalCode: %s, radiusMiles: %d", postalCode, radiusMiles)

	if radiusMiles <= 0 {
		radiusMiles = 25
	}

	endpoint := fmt.Sprintf("%s/stores(area(%s,%d))?format=json&show=storeId,name,address,address2,city,region,postalCode,phone,distance,storeType,hours,hoursAmPm,gmtOffset,lat,lng&pageSize=50&apiKey=%s",
		c.baseURL, url.QueryEscape(postalCode), radiusMiles, c.apiKey)

	log.Printf("Searching stores with endpoint: %s", endpoint)

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		log.Printf("Store search error: %v", err)
		return nil, err
	}

	var result storesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Failed to decode store search response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("Store search returned %d results", len(result.Stores))
	return result.Stores, nil
}

// skuPattern matches strings that look like SKUs (6-8 digits)
var skuPattern = regexp.MustCompile(`^\d{6,8}$`)

// SearchProducts searches for products by keyword or SKU, optionally filtered by subclass
func (c *APIClient) SearchProducts(ctx context.Context, query string, subclass string) ([]Product, error) {
	log.Printf("SearchProducts called with query: %s, subclass: %s", query, subclass)

	// Check if the query looks like a SKU (6-8 digit number)
	if skuPattern.MatchString(query) {
		log.Printf("Query looks like a SKU, trying direct lookup first")
		product, err := c.GetProductBySKU(ctx, query)
		if err == nil && product != nil && product.SKU != 0 {
			log.Printf("Found product by SKU: %s - %s", query, product.Name)
			return []Product{*product}, nil
		}
		log.Printf("SKU lookup failed or returned empty, falling back to search: %v", err)
	}

	// Build the filter query
	var filterParts []string
	if query != "" {
		filterParts = append(filterParts, fmt.Sprintf("search=%s", url.PathEscape(query)))
	}
	if subclass != "" {
		filterParts = append(filterParts, fmt.Sprintf("subclass=%s", url.PathEscape(subclass)))
	}
	filterParts = append(filterParts, "active=*") // Include inactive products

	filter := ""
	for i, part := range filterParts {
		if i > 0 {
			filter += "&"
		}
		filter += part
	}

	endpoint := fmt.Sprintf("%s/products(%s)?format=json&show=sku,name,salePrice,regularPrice,thumbnailImage,image,url,shortDescription,manufacturer,modelNumber,upc,inStoreAvailability,onlineAvailability&pageSize=50&apiKey=%s",
		c.baseURL, filter, c.apiKey)

	log.Printf("Searching products with endpoint: %s", endpoint)

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		log.Printf("Product search error: %v", err)
		return nil, err
	}

	var result productsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Failed to decode product search response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("Product search returned %d results", len(result.Products))
	return result.Products, nil
}

// GetProductBySKU gets a single product by SKU
func (c *APIClient) GetProductBySKU(ctx context.Context, sku string) (*Product, error) {
	endpoint := fmt.Sprintf("%s/products/%s.json?apiKey=%s",
		c.baseURL, url.PathEscape(sku), c.apiKey)

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var product Product
	if err := json.Unmarshal(body, &product); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &product, nil
}

// SearchProductsInCategory searches for products within a specific category
func (c *APIClient) SearchProductsInCategory(ctx context.Context, categoryID string, query string) ([]Product, error) {
	log.Printf("SearchProductsInCategory called with categoryID: %s, query: %s", categoryID, query)

	var endpoint string
	if query != "" {
		endpoint = fmt.Sprintf("%s/products(categoryPath.id=%s&search=%s)?format=json&show=sku,name,salePrice,regularPrice,thumbnailImage,image,url,shortDescription,manufacturer,modelNumber,upc,inStoreAvailability,onlineAvailability&pageSize=100&apiKey=%s",
			c.baseURL, categoryID, url.PathEscape(query), c.apiKey)
	} else {
		endpoint = fmt.Sprintf("%s/products(categoryPath.id=%s)?format=json&show=sku,name,salePrice,regularPrice,thumbnailImage,image,url,shortDescription,manufacturer,modelNumber,upc,inStoreAvailability,onlineAvailability&pageSize=100&apiKey=%s",
			c.baseURL, categoryID, c.apiKey)
	}

	log.Printf("Category search endpoint: %s", endpoint)

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		log.Printf("Category search error: %v", err)
		return nil, err
	}

	var result productsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Failed to decode category search response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("Category search returned %d results", len(result.Products))
	return result.Products, nil
}

// BrowsePokemonProducts returns Pokemon TCG products (including inactive ones)
func (c *APIClient) BrowsePokemonProducts(ctx context.Context) ([]Product, error) {
	log.Printf("BrowsePokemonProducts called")

	// Search for Pokemon TCG cards by subclass, including inactive products
	// Best Buy marks most Pokemon TCG as "inactive" due to invitation system
	endpoint := fmt.Sprintf("%s/products(subclass=POKEMON%%20CARDS&active=*)?format=json&show=sku,name,salePrice,regularPrice,thumbnailImage,image,url,shortDescription,manufacturer,modelNumber,upc,inStoreAvailability,onlineAvailability&pageSize=100&apiKey=%s",
		c.baseURL, c.apiKey)

	log.Printf("Browse Pokemon endpoint: %s", endpoint)

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		log.Printf("Browse Pokemon error: %v", err)
		return nil, err
	}

	var result productsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Failed to decode browse Pokemon response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("Browse Pokemon returned %d results", len(result.Products))
	return result.Products, nil
}

// storesProductsResponse is the API response for combined stores+products query
type storesProductsResponse struct {
	Stores []struct {
		StoreID  int     `json:"storeId"`
		Name     string  `json:"name"`
		City     string  `json:"city"`
		State    string  `json:"region"`
		Distance float64 `json:"distance"`
		Products []struct {
			SKU              int    `json:"sku"`
			Name             string `json:"name"`
			InStorePickup    bool   `json:"inStorePickup"`
			FriendsFamilyPickup bool `json:"friendsAndFamilyPickup"`
		} `json:"products"`
	} `json:"stores"`
	Total int `json:"total"`
}

// CheckAvailability checks product availability at specific stores
func (c *APIClient) CheckAvailability(ctx context.Context, sku string, storeIDs []string) ([]StoreAvailability, error) {
	log.Printf("CheckAvailability called with sku: %s, storeIDs: %v", sku, storeIDs)

	// Build storeId list for the in() operator
	storeIDsParam := ""
	for i, id := range storeIDs {
		if i > 0 {
			storeIDsParam += ","
		}
		storeIDsParam += id
	}

	// Use the stores+products combined query format
	// This returns stores that have the product in stock
	endpoint := fmt.Sprintf("%s/stores(storeId%%20in(%s))+products(sku=%s)?format=json&show=storeId,name,city,region,distance,products.sku,products.name,products.inStorePickup,products.friendsAndFamilyPickup&apiKey=%s",
		c.baseURL, storeIDsParam, url.PathEscape(sku), c.apiKey)

	log.Printf("CheckAvailability endpoint: %s", endpoint)

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		// Check if it's a 403 Forbidden error (Best Buy blocks availability for some products)
		if strings.Contains(err.Error(), "403") {
			log.Printf("CheckAvailability: Access forbidden for SKU %s (likely restricted product)", sku)
			// Return empty availability with a note that it's restricted
			return []StoreAvailability{}, nil
		}
		log.Printf("CheckAvailability error: %v", err)
		return nil, err
	}

	var result storesProductsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Failed to decode availability response: %v, body: %s", err, string(body))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("CheckAvailability returned %d stores with product in stock", len(result.Stores))

	// Convert to StoreAvailability
	// Only stores that have the product show up in the response
	availability := make([]StoreAvailability, 0, len(result.Stores))
	for _, store := range result.Stores {
		availability = append(availability, StoreAvailability{
			StoreID:        fmt.Sprintf("%d", store.StoreID),
			StoreName:      store.Name,
			City:           store.City,
			State:          store.State,
			Distance:       store.Distance,
			InStock:        true, // If store is in response, product is available
			LowStock:       false,
			PickupEligible: len(store.Products) > 0 && store.Products[0].InStorePickup,
		})
	}

	return availability, nil
}
