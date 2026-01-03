package bestbuy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client is the interface for Best Buy API operations
type Client interface {
	// SearchStores searches for stores near a postal code within a radius
	SearchStores(ctx context.Context, postalCode string, radiusMiles int) ([]Store, error)

	// SearchProducts searches for products by keyword
	SearchProducts(ctx context.Context, query string) ([]Product, error)

	// GetProductBySKU gets a single product by its SKU
	GetProductBySKU(ctx context.Context, sku string) (*Product, error)

	// CheckAvailability checks product availability at specific stores
	CheckAvailability(ctx context.Context, sku string, storeIDs []string) ([]StoreAvailability, error)
}

// Store represents a Best Buy store from the API
type Store struct {
	StoreID       string  `json:"storeId"`
	Name          string  `json:"name"`
	Address       string  `json:"address"`
	Address2      string  `json:"address2"`
	City          string  `json:"city"`
	State         string  `json:"state"`
	PostalCode    string  `json:"postalCode"`
	Phone         string  `json:"phone"`
	Distance      float64 `json:"distance"`
	StoreType     string  `json:"storeType"`
	Hours         string  `json:"hours"`
	HoursAmPm     string  `json:"hoursAmPm"`
	GMTOffset     int     `json:"gmtOffset"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
}

// Product represents a Best Buy product from the API
type Product struct {
	SKU                 string  `json:"sku"`
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

// APIClient is the real Best Buy API client implementation
type APIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient creates a new Best Buy API client
func NewAPIClient(apiKey string) *APIClient {
	return &APIClient{
		apiKey:  apiKey,
		baseURL: "https://api.bestbuy.com/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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
		StoreID   string `json:"storeId"`
		StoreName string `json:"name"`
		City      string `json:"city"`
		State     string `json:"state"`
		Distance  float64 `json:"distance"`
		LowStock  bool   `json:"lowStock"`
	} `json:"stores"`
}

// SearchStores searches for stores near a postal code
func (c *APIClient) SearchStores(ctx context.Context, postalCode string, radiusMiles int) ([]Store, error) {
	if radiusMiles <= 0 {
		radiusMiles = 25
	}

	// Build the URL: /stores(area(postalCode,radius))?format=json&apiKey=XXX
	endpoint := fmt.Sprintf("%s/stores(area(%s,%d))?format=json&show=storeId,name,address,address2,city,state,postalCode,phone,distance,storeType,hours,hoursAmPm,gmtOffset,lat,lng&pageSize=50&apiKey=%s",
		c.baseURL, url.QueryEscape(postalCode), radiusMiles, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result storesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Stores, nil
}

// SearchProducts searches for products by keyword
func (c *APIClient) SearchProducts(ctx context.Context, query string) ([]Product, error) {
	// Build the URL: /products(search=query)?format=json&apiKey=XXX
	endpoint := fmt.Sprintf("%s/products(search=%s)?format=json&show=sku,name,salePrice,regularPrice,thumbnailImage,image,url,shortDescription,manufacturer,modelNumber,upc,inStoreAvailability,onlineAvailability&pageSize=50&apiKey=%s",
		c.baseURL, url.QueryEscape(query), c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result productsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Products, nil
}

// GetProductBySKU gets a single product by SKU
func (c *APIClient) GetProductBySKU(ctx context.Context, sku string) (*Product, error) {
	// Build the URL: /products/{sku}.json?apiKey=XXX
	endpoint := fmt.Sprintf("%s/products/%s.json?apiKey=%s",
		c.baseURL, url.PathEscape(sku), c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("product not found: %s", sku)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var product Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &product, nil
}

// CheckAvailability checks product availability at specific stores
func (c *APIClient) CheckAvailability(ctx context.Context, sku string, storeIDs []string) ([]StoreAvailability, error) {
	// For each store, we need to check availability
	// The API endpoint is: /products/{sku}/stores.json?storeId=1,2,3&apiKey=XXX

	storeIDsParam := ""
	for i, id := range storeIDs {
		if i > 0 {
			storeIDsParam += ","
		}
		storeIDsParam += id
	}

	endpoint := fmt.Sprintf("%s/products/%s/stores.json?storeId=%s&apiKey=%s",
		c.baseURL, url.PathEscape(sku), url.QueryEscape(storeIDsParam), c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result availabilityResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to StoreAvailability
	availability := make([]StoreAvailability, 0, len(result.Stores))
	for _, store := range result.Stores {
		availability = append(availability, StoreAvailability{
			StoreID:        store.StoreID,
			StoreName:      store.StoreName,
			City:           store.City,
			State:          store.State,
			Distance:       store.Distance,
			InStock:        true, // If store is in response, product is available
			LowStock:       store.LowStock,
			PickupEligible: true, // Assume pickup eligible if in-store available
		})
	}

	return availability, nil
}
