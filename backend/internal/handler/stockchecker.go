package handler

import (
	"context"
	"log"

	"connectrpc.com/connect"
	stockcheckerv1 "github.com/tmcauley/stock-checker/backend/gen/stockchecker/v1"
	"github.com/tmcauley/stock-checker/backend/gen/stockchecker/v1/stockcheckerv1connect"
	"github.com/tmcauley/stock-checker/backend/internal/bestbuy"
)

// StockCheckerHandler implements the StockCheckerService
type StockCheckerHandler struct {
	stockcheckerv1connect.UnimplementedStockCheckerServiceHandler
	bbClient bestbuy.Client
}

// NewStockCheckerHandler creates a new StockCheckerHandler
func NewStockCheckerHandler(bbClient bestbuy.Client) *StockCheckerHandler {
	return &StockCheckerHandler{
		bbClient: bbClient,
	}
}

// SearchStores searches for Best Buy stores near a location
func (h *StockCheckerHandler) SearchStores(
	ctx context.Context,
	req *connect.Request[stockcheckerv1.SearchStoresRequest],
) (*connect.Response[stockcheckerv1.SearchStoresResponse], error) {
	radiusMiles := int(req.Msg.RadiusMiles)
	if radiusMiles <= 0 {
		radiusMiles = 25
	}

	stores, err := h.bbClient.SearchStores(ctx, req.Msg.PostalCode, radiusMiles)
	if err != nil {
		log.Printf("Error searching stores: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to protobuf messages
	pbStores := make([]*stockcheckerv1.Store, 0, len(stores))
	for _, store := range stores {
		pbStores = append(pbStores, &stockcheckerv1.Store{
			StoreId:       store.StoreID,
			Name:          store.Name,
			Address:       store.Address,
			City:          store.City,
			State:         store.State,
			PostalCode:    store.PostalCode,
			Phone:         store.Phone,
			DistanceMiles: store.Distance,
		})
	}

	return connect.NewResponse(&stockcheckerv1.SearchStoresResponse{
		Stores: pbStores,
	}), nil
}

// SearchProducts searches for products by keyword or SKU
func (h *StockCheckerHandler) SearchProducts(
	ctx context.Context,
	req *connect.Request[stockcheckerv1.SearchProductsRequest],
) (*connect.Response[stockcheckerv1.SearchProductsResponse], error) {
	products, err := h.bbClient.SearchProducts(ctx, req.Msg.Query)
	if err != nil {
		log.Printf("Error searching products: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to protobuf messages
	pbProducts := make([]*stockcheckerv1.Product, 0, len(products))
	for _, product := range products {
		pbProducts = append(pbProducts, &stockcheckerv1.Product{
			Sku:          product.SKU,
			Name:         product.Name,
			SalePrice:    product.SalePrice,
			ThumbnailUrl: product.ThumbnailImage,
			ProductUrl:   product.URL,
		})
	}

	return connect.NewResponse(&stockcheckerv1.SearchProductsResponse{
		Products: pbProducts,
	}), nil
}

// CheckStock checks inventory for products at specified stores
func (h *StockCheckerHandler) CheckStock(
	ctx context.Context,
	req *connect.Request[stockcheckerv1.CheckStockRequest],
) (*connect.Response[stockcheckerv1.CheckStockResponse], error) {
	storeIDs := req.Msg.StoreIds
	skus := req.Msg.Skus

	if len(storeIDs) == 0 || len(skus) == 0 {
		return connect.NewResponse(&stockcheckerv1.CheckStockResponse{
			Results: []*stockcheckerv1.StockStatus{},
		}), nil
	}

	// Check availability for each SKU
	var results []*stockcheckerv1.StockStatus

	for _, sku := range skus {
		// Get product info
		product, err := h.bbClient.GetProductBySKU(ctx, sku)
		if err != nil {
			log.Printf("Error getting product %s: %v", sku, err)
			continue
		}

		// Check availability at stores
		availability, err := h.bbClient.CheckAvailability(ctx, sku, storeIDs)
		if err != nil {
			log.Printf("Error checking availability for %s: %v", sku, err)
			continue
		}

		// Convert to StockStatus
		for _, avail := range availability {
			results = append(results, &stockcheckerv1.StockStatus{
				Store: &stockcheckerv1.Store{
					StoreId: avail.StoreID,
					Name:    avail.StoreName,
					City:    avail.City,
					State:   avail.State,
				},
				Product: &stockcheckerv1.Product{
					Sku:       product.SKU,
					Name:      product.Name,
					SalePrice: product.SalePrice,
				},
				InStock:        avail.InStock,
				LowStock:       avail.LowStock,
				PickupEligible: avail.PickupEligible,
			})
		}
	}

	return connect.NewResponse(&stockcheckerv1.CheckStockResponse{
		Results: results,
	}), nil
}
