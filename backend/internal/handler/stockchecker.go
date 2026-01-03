package handler

import (
	"context"
	"fmt"
	"log"

	"connectrpc.com/connect"
	stockcheckerv1 "github.com/tmcauley/stock-checker/backend/gen/stockchecker/v1"
	"github.com/tmcauley/stock-checker/backend/gen/stockchecker/v1/stockcheckerv1connect"
	"github.com/tmcauley/stock-checker/backend/internal/auth"
	"github.com/tmcauley/stock-checker/backend/internal/bestbuy"
	"github.com/tmcauley/stock-checker/backend/internal/database"
)

// StockCheckerHandler implements the StockCheckerService
type StockCheckerHandler struct {
	stockcheckerv1connect.UnimplementedStockCheckerServiceHandler
	bbClient bestbuy.Client
	db       *database.DB
}

// NewStockCheckerHandler creates a new StockCheckerHandler
func NewStockCheckerHandler(bbClient bestbuy.Client, db *database.DB) *StockCheckerHandler {
	return &StockCheckerHandler{
		bbClient: bbClient,
		db:       db,
	}
}

// getUserFromContext gets the authenticated user from context
func getUserFromContext(ctx context.Context) (*database.User, error) {
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not authenticated"))
	}
	return user, nil
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
			StoreId:       fmt.Sprintf("%d", store.StoreID),
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
	products, err := h.bbClient.SearchProducts(ctx, req.Msg.Query, req.Msg.Category)
	if err != nil {
		log.Printf("Error searching products: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to protobuf messages
	pbProducts := make([]*stockcheckerv1.Product, 0, len(products))
	for _, product := range products {
		pbProducts = append(pbProducts, &stockcheckerv1.Product{
			Sku:          fmt.Sprintf("%d", product.SKU),
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
					Sku:       fmt.Sprintf("%d", product.SKU),
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

// GetCurrentUser returns the currently authenticated user
func (h *StockCheckerHandler) GetCurrentUser(
	ctx context.Context,
	req *connect.Request[stockcheckerv1.GetCurrentUserRequest],
) (*connect.Response[stockcheckerv1.GetCurrentUserResponse], error) {
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&stockcheckerv1.GetCurrentUserResponse{
		User: &stockcheckerv1.User{
			Id:         int32(user.ID),
			Email:      user.Email,
			Name:       user.Name,
			PictureUrl: user.PictureURL,
		},
	}), nil
}

// GetMyStores returns the user's saved stores
func (h *StockCheckerHandler) GetMyStores(
	ctx context.Context,
	req *connect.Request[stockcheckerv1.GetMyStoresRequest],
) (*connect.Response[stockcheckerv1.GetMyStoresResponse], error) {
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	stores, err := h.db.GetUserStores(ctx, user.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbStores := make([]*stockcheckerv1.Store, 0, len(stores))
	for _, store := range stores {
		pbStores = append(pbStores, &stockcheckerv1.Store{
			StoreId:    store.StoreID,
			Name:       store.Name,
			Address:    store.Address,
			City:       store.City,
			State:      store.State,
			PostalCode: store.PostalCode,
			Phone:      store.Phone,
		})
	}

	return connect.NewResponse(&stockcheckerv1.GetMyStoresResponse{
		Stores: pbStores,
	}), nil
}

// AddMyStore adds a store to the user's list
func (h *StockCheckerHandler) AddMyStore(
	ctx context.Context,
	req *connect.Request[stockcheckerv1.AddMyStoreRequest],
) (*connect.Response[stockcheckerv1.AddMyStoreResponse], error) {
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	store := req.Msg.Store
	if store == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("store is required"))
	}

	dbStore := database.Store{
		StoreID:    store.StoreId,
		Name:       store.Name,
		Address:    store.Address,
		City:       store.City,
		State:      store.State,
		PostalCode: store.PostalCode,
		Phone:      store.Phone,
	}

	if err := h.db.AddUserStore(ctx, user.ID, dbStore); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&stockcheckerv1.AddMyStoreResponse{}), nil
}

// RemoveMyStore removes a store from the user's list
func (h *StockCheckerHandler) RemoveMyStore(
	ctx context.Context,
	req *connect.Request[stockcheckerv1.RemoveMyStoreRequest],
) (*connect.Response[stockcheckerv1.RemoveMyStoreResponse], error) {
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.db.RemoveUserStore(ctx, user.ID, req.Msg.StoreId); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&stockcheckerv1.RemoveMyStoreResponse{}), nil
}

// GetMyProducts returns the user's saved products
func (h *StockCheckerHandler) GetMyProducts(
	ctx context.Context,
	req *connect.Request[stockcheckerv1.GetMyProductsRequest],
) (*connect.Response[stockcheckerv1.GetMyProductsResponse], error) {
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	products, err := h.db.GetUserProducts(ctx, user.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbProducts := make([]*stockcheckerv1.Product, 0, len(products))
	for _, product := range products {
		pbProducts = append(pbProducts, &stockcheckerv1.Product{
			Sku:          product.SKU,
			Name:         product.Name,
			SalePrice:    product.SalePrice,
			ThumbnailUrl: product.ThumbnailURL,
			ProductUrl:   product.ProductURL,
		})
	}

	return connect.NewResponse(&stockcheckerv1.GetMyProductsResponse{
		Products: pbProducts,
	}), nil
}

// AddMyProduct adds a product to the user's list
func (h *StockCheckerHandler) AddMyProduct(
	ctx context.Context,
	req *connect.Request[stockcheckerv1.AddMyProductRequest],
) (*connect.Response[stockcheckerv1.AddMyProductResponse], error) {
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	product := req.Msg.Product
	if product == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("product is required"))
	}

	dbProduct := database.Product{
		SKU:          product.Sku,
		Name:         product.Name,
		SalePrice:    product.SalePrice,
		ThumbnailURL: product.ThumbnailUrl,
		ProductURL:   product.ProductUrl,
	}

	if err := h.db.AddUserProduct(ctx, user.ID, dbProduct); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&stockcheckerv1.AddMyProductResponse{}), nil
}

// RemoveMyProduct removes a product from the user's list
func (h *StockCheckerHandler) RemoveMyProduct(
	ctx context.Context,
	req *connect.Request[stockcheckerv1.RemoveMyProductRequest],
) (*connect.Response[stockcheckerv1.RemoveMyProductResponse], error) {
	user, err := getUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.db.RemoveUserProduct(ctx, user.ID, req.Msg.Sku); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&stockcheckerv1.RemoveMyProductResponse{}), nil
}

// BrowsePokemonProducts returns Pokemon products from Best Buy's trading cards category
func (h *StockCheckerHandler) BrowsePokemonProducts(
	ctx context.Context,
	req *connect.Request[stockcheckerv1.BrowsePokemonProductsRequest],
) (*connect.Response[stockcheckerv1.BrowsePokemonProductsResponse], error) {
	products, err := h.bbClient.BrowsePokemonProducts(ctx)
	if err != nil {
		log.Printf("Error browsing Pokemon products: %v", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to protobuf messages
	pbProducts := make([]*stockcheckerv1.Product, 0, len(products))
	for _, product := range products {
		pbProducts = append(pbProducts, &stockcheckerv1.Product{
			Sku:          fmt.Sprintf("%d", product.SKU),
			Name:         product.Name,
			SalePrice:    product.SalePrice,
			ThumbnailUrl: product.ThumbnailImage,
			ProductUrl:   product.URL,
		})
	}

	return connect.NewResponse(&stockcheckerv1.BrowsePokemonProductsResponse{
		Products: pbProducts,
	}), nil
}
