package bestbuy

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// MockClient is a mock implementation of the Best Buy API client for testing
type MockClient struct {
	// Simulate network latency
	latency time.Duration
}

// NewMockClient creates a new mock client
func NewMockClient() *MockClient {
	return &MockClient{
		latency: 100 * time.Millisecond, // Simulate 100ms API latency
	}
}

// mockStores contains realistic mock store data
var mockStores = []Store{
	{
		StoreID:    1118,
		Name:       "Best Buy - San Francisco",
		Address:    "1717 Harrison St",
		City:       "San Francisco",
		State:      "CA",
		PostalCode: "94103",
		Phone:      "(415) 626-9682",
		StoreType:  "Big Box",
		Lat:        37.7699,
		Lng:        -122.4134,
	},
	{
		StoreID:    1009,
		Name:       "Best Buy - Daly City",
		Address:    "133 Serramonte Center",
		City:       "Daly City",
		State:      "CA",
		PostalCode: "94015",
		Phone:      "(650) 991-9289",
		StoreType:  "Big Box",
		Lat:        37.6710,
		Lng:        -122.4687,
	},
	{
		StoreID:    187,
		Name:       "Best Buy - Emeryville",
		Address:    "3700 Mandela Pkwy",
		City:       "Emeryville",
		State:      "CA",
		PostalCode: "94608",
		Phone:      "(510) 596-1531",
		StoreType:  "Big Box",
		Lat:        37.8358,
		Lng:        -122.2914,
	},
	{
		StoreID:    1444,
		Name:       "Best Buy - San Bruno",
		Address:    "899 El Camino Real",
		City:       "San Bruno",
		State:      "CA",
		PostalCode: "94066",
		Phone:      "(650) 873-3688",
		StoreType:  "Big Box",
		Lat:        37.6252,
		Lng:        -122.4117,
	},
	{
		StoreID:    573,
		Name:       "Best Buy - Colma",
		Address:    "4821 Colma Blvd",
		City:       "Colma",
		State:      "CA",
		PostalCode: "94014",
		Phone:      "(650) 757-0381",
		StoreType:  "Big Box",
		Lat:        37.6769,
		Lng:        -122.4583,
	},
	{
		StoreID:    499,
		Name:       "Best Buy - Oakland",
		Address:    "2110 Broadway",
		City:       "Oakland",
		State:      "CA",
		PostalCode: "94612",
		Phone:      "(510) 625-0565",
		StoreType:  "Big Box",
		Lat:        37.8124,
		Lng:        -122.2685,
	},
}

// mockProducts contains realistic Pokemon card products
var mockProducts = []Product{
	{
		SKU:                 6579543,
		Name:                "Pokemon Trading Card Game: Scarlet & Violet Prismatic Evolutions Elite Trainer Box",
		SalePrice:           59.99,
		RegularPrice:        59.99,
		ThumbnailImage:      "https://pisces.bbystatic.com/image2/BestBuy_US/images/products/6579/6579543_sd.jpg",
		URL:                 "https://www.bestbuy.com/site/pokemon-trading-card-game-scarlet-violet-prismatic-evolutions-elite-trainer-box/6579543.p",
		ShortDescription:    "Get ready for battle with the Prismatic Evolutions Elite Trainer Box!",
		Manufacturer:        "Pokemon",
		InStoreAvailability: true,
		OnlineAvailability:  false,
	},
	{
		SKU:                 6579544,
		Name:                "Pokemon Trading Card Game: Scarlet & Violet Prismatic Evolutions Booster Bundle",
		SalePrice:           29.99,
		RegularPrice:        29.99,
		ThumbnailImage:      "https://pisces.bbystatic.com/image2/BestBuy_US/images/products/6579/6579544_sd.jpg",
		URL:                 "https://www.bestbuy.com/site/pokemon-trading-card-game-scarlet-violet-prismatic-evolutions-booster-bundle/6579544.p",
		ShortDescription:    "Collect amazing cards with the Prismatic Evolutions Booster Bundle!",
		Manufacturer:        "Pokemon",
		InStoreAvailability: true,
		OnlineAvailability:  false,
	},
	{
		SKU:                 6579545,
		Name:                "Pokemon Trading Card Game: Scarlet & Violet Prismatic Evolutions Booster Pack",
		SalePrice:           4.99,
		RegularPrice:        4.99,
		ThumbnailImage:      "https://pisces.bbystatic.com/image2/BestBuy_US/images/products/6579/6579545_sd.jpg",
		URL:                 "https://www.bestbuy.com/site/pokemon-trading-card-game-scarlet-violet-prismatic-evolutions-booster-pack/6579545.p",
		ShortDescription:    "Each booster pack contains 10 cards from the Prismatic Evolutions expansion!",
		Manufacturer:        "Pokemon",
		InStoreAvailability: true,
		OnlineAvailability:  true,
	},
	{
		SKU:                 6543210,
		Name:                "Pokemon Trading Card Game: Scarlet & Violet 151 Ultra Premium Collection",
		SalePrice:           139.99,
		RegularPrice:        139.99,
		ThumbnailImage:      "https://pisces.bbystatic.com/image2/BestBuy_US/images/products/6543/6543210_sd.jpg",
		URL:                 "https://www.bestbuy.com/site/pokemon-trading-card-game-scarlet-violet-151-ultra-premium-collection/6543210.p",
		ShortDescription:    "The ultimate Pokemon 151 collection featuring exclusive cards!",
		Manufacturer:        "Pokemon",
		InStoreAvailability: false,
		OnlineAvailability:  false,
	},
	{
		SKU:                 6543211,
		Name:                "Pokemon Trading Card Game: Scarlet & Violet 151 Elite Trainer Box",
		SalePrice:           49.99,
		RegularPrice:        49.99,
		ThumbnailImage:      "https://pisces.bbystatic.com/image2/BestBuy_US/images/products/6543/6543211_sd.jpg",
		URL:                 "https://www.bestbuy.com/site/pokemon-trading-card-game-scarlet-violet-151-elite-trainer-box/6543211.p",
		ShortDescription:    "Collect the original 151 Pokemon with this Elite Trainer Box!",
		Manufacturer:        "Pokemon",
		InStoreAvailability: true,
		OnlineAvailability:  false,
	},
	{
		SKU:                 6578901,
		Name:                "Pokemon Trading Card Game: Surging Sparks Elite Trainer Box",
		SalePrice:           54.99,
		RegularPrice:        54.99,
		ThumbnailImage:      "https://pisces.bbystatic.com/image2/BestBuy_US/images/products/6578/6578901_sd.jpg",
		URL:                 "https://www.bestbuy.com/site/pokemon-trading-card-game-surging-sparks-elite-trainer-box/6578901.p",
		ShortDescription:    "Power up with the Surging Sparks Elite Trainer Box!",
		Manufacturer:        "Pokemon",
		InStoreAvailability: true,
		OnlineAvailability:  true,
	},
	{
		SKU:                 6578902,
		Name:                "Pokemon Trading Card Game: Surging Sparks Booster Bundle",
		SalePrice:           24.99,
		RegularPrice:        24.99,
		ThumbnailImage:      "https://pisces.bbystatic.com/image2/BestBuy_US/images/products/6578/6578902_sd.jpg",
		URL:                 "https://www.bestbuy.com/site/pokemon-trading-card-game-surging-sparks-booster-bundle/6578902.p",
		ShortDescription:    "Get 6 booster packs in this Surging Sparks bundle!",
		Manufacturer:        "Pokemon",
		InStoreAvailability: true,
		OnlineAvailability:  true,
	},
	{
		SKU:                 6512345,
		Name:                "Pokemon Trading Card Game: Paldean Fates Elite Trainer Box",
		SalePrice:           59.99,
		RegularPrice:        59.99,
		ThumbnailImage:      "https://pisces.bbystatic.com/image2/BestBuy_US/images/products/6512/6512345_sd.jpg",
		URL:                 "https://www.bestbuy.com/site/pokemon-trading-card-game-paldean-fates-elite-trainer-box/6512345.p",
		ShortDescription:    "Discover shiny Pokemon with the Paldean Fates Elite Trainer Box!",
		Manufacturer:        "Pokemon",
		InStoreAvailability: true,
		OnlineAvailability:  false,
	},
}

// simulateLatency adds a small delay to simulate network latency
func (c *MockClient) simulateLatency(ctx context.Context) error {
	select {
	case <-time.After(c.latency):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SearchStores returns mock stores based on postal code
func (c *MockClient) SearchStores(ctx context.Context, postalCode string, radiusMiles int) ([]Store, error) {
	if err := c.simulateLatency(ctx); err != nil {
		return nil, err
	}

	// Return stores with calculated mock distances
	stores := make([]Store, len(mockStores))
	for i, store := range mockStores {
		stores[i] = store
		// Generate a random distance between 1 and radiusMiles
		stores[i].Distance = float64(rand.Intn(radiusMiles)) + rand.Float64()
	}

	return stores, nil
}

// SearchProducts searches for products by keyword, optionally filtered by subclass
func (c *MockClient) SearchProducts(ctx context.Context, query string, subclass string) ([]Product, error) {
	if err := c.simulateLatency(ctx); err != nil {
		return nil, err
	}

	// Filter products that match the query (case-insensitive)
	queryLower := strings.ToLower(query)
	var results []Product

	for _, product := range mockProducts {
		if query == "" || strings.Contains(strings.ToLower(product.Name), queryLower) ||
			strings.Contains(fmt.Sprintf("%d", product.SKU), queryLower) ||
			strings.Contains(strings.ToLower(product.ShortDescription), queryLower) {
			results = append(results, product)
		}
	}

	// If no matches found and query looks like it could be Pokemon related, return all
	if len(results) == 0 && (strings.Contains(queryLower, "pokemon") || strings.Contains(queryLower, "card")) {
		return mockProducts, nil
	}

	return results, nil
}

// GetProductBySKU gets a single product by SKU
func (c *MockClient) GetProductBySKU(ctx context.Context, sku string) (*Product, error) {
	if err := c.simulateLatency(ctx); err != nil {
		return nil, err
	}

	for _, product := range mockProducts {
		if fmt.Sprintf("%d", product.SKU) == sku {
			return &product, nil
		}
	}
	return nil, fmt.Errorf("product not found: %s", sku)
}

// CheckAvailability checks product availability using postal code
func (c *MockClient) CheckAvailability(ctx context.Context, sku string, postalCode string) ([]StoreAvailability, error) {
	if err := c.simulateLatency(ctx); err != nil {
		return nil, err
	}

	// Find the product first
	var product *Product
	for _, p := range mockProducts {
		if fmt.Sprintf("%d", p.SKU) == sku {
			product = &p
			break
		}
	}

	if product == nil {
		return nil, fmt.Errorf("product not found: %s", sku)
	}

	// Generate availability for all mock stores (simulating postal code search)
	availability := make([]StoreAvailability, 0)

	for _, store := range mockStores {
		storeID := fmt.Sprintf("%d", store.StoreID)

		// Determine availability based on product and some randomness
		// Use a seeded random based on store+product to get consistent results
		seed := int64(0)
		for _, c := range storeID + sku {
			seed += int64(c)
		}
		r := rand.New(rand.NewSource(seed))
		roll := r.Float64()

		var inStock, lowStock bool

		if !product.InStoreAvailability {
			// Product not typically in stores - very rare to find
			inStock = roll < 0.1
			lowStock = false
		} else {
			// Normal product - 50% in stock, 20% low stock, 30% out of stock
			inStock = roll < 0.7
			lowStock = roll >= 0.5 && roll < 0.7
		}

		// Only add stores that have stock (like the real API)
		if inStock {
			availability = append(availability, StoreAvailability{
				StoreID:        storeID,
				StoreName:      store.Name,
				City:           store.City,
				State:          store.State,
				Distance:       store.Distance,
				InStock:        inStock,
				LowStock:       lowStock,
				PickupEligible: inStock,
			})
		}
	}

	return availability, nil
}

// SearchProductsInCategory searches for products within a specific category
func (c *MockClient) SearchProductsInCategory(ctx context.Context, categoryID string, query string) ([]Product, error) {
	// For mock, just delegate to regular search
	return c.SearchProducts(ctx, query, "")
}

// BrowsePokemonProducts returns Pokemon TCG products
func (c *MockClient) BrowsePokemonProducts(ctx context.Context) ([]Product, error) {
	if err := c.simulateLatency(ctx); err != nil {
		return nil, err
	}
	// Return all mock products (they're all Pokemon)
	return mockProducts, nil
}
