# Pokemon Stock Checker - Implementation Plan

## Overview

A local-first application to track Pokemon card inventory at Best Buy stores. Users can configure their preferred stores and products, then check real-time stock availability.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Frontend (React + Vite)                         │
│                                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Store Search│  │ Product     │  │ My Lists    │  │ Stock Check Results │ │
│  │ Page        │  │ Search Page │  │ Page        │  │ Page                │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│                                    │                                         │
│                         localStorage (persistence)                           │
└────────────────────────────────────┬────────────────────────────────────────┘
                                     │ Connect (protobuf)
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Backend (Go + Connect)                          │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                         Connect-Go Service                               ││
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────────────┐ ││
│  │  │ SearchStores│ │SearchProducts│ │CheckStock  │  │ Best Buy API Client│ ││
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────────────┘ ││
│  └─────────────────────────────────────────────────────────────────────────┘│
└────────────────────────────────────┬────────────────────────────────────────┘
                                     │ JSON/REST
                                     ▼
                          ┌─────────────────────┐
                          │   Best Buy API      │
                          │   - Stores API      │
                          │   - Products API    │
                          │   - Availability API│
                          └─────────────────────┘
```

## Tech Stack

### Frontend
- **Framework:** React 18 + TypeScript
- **Build Tool:** Vite
- **Styling:** Tailwind CSS
- **State Management:** React Context + localStorage
- **API Client:** Connect-Web (generated from protobufs)
- **Routing:** React Router v6

### Backend
- **Language:** Go 1.21+
- **API Framework:** Connect-Go (buf.build/connectrpc)
- **Protobuf Tooling:** Buf CLI
- **HTTP Server:** net/http (stdlib)

### Best Buy API Endpoints
- **Stores:** `GET /v1/stores` - Search by postal code, area, or ID
- **Products:** `GET /v1/products` - Search by keyword or SKU
- **Availability:** `GET /v1/products/{sku}/stores.json` - In-store stock status

---

## Project Structure

```
stock-checker/
├── frontend/
│   ├── src/
│   │   ├── components/          # Reusable UI components
│   │   ├── pages/               # Page components
│   │   ├── hooks/               # Custom React hooks
│   │   ├── context/             # React context providers
│   │   ├── gen/                 # Generated protobuf code
│   │   ├── lib/                 # Utilities and helpers
│   │   └── App.tsx
│   ├── index.html
│   ├── tailwind.config.js
│   ├── vite.config.ts
│   └── package.json
│
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go          # Entry point
│   ├── internal/
│   │   ├── handler/             # Connect service handlers
│   │   ├── bestbuy/             # Best Buy API client
│   │   └── config/              # Configuration
│   ├── gen/                     # Generated protobuf code
│   └── go.mod
│
├── proto/
│   ├── buf.yaml                 # Buf configuration
│   ├── buf.gen.yaml             # Code generation config
│   └── stockchecker/
│       └── v1/
│           ├── store.proto      # Store-related messages
│           ├── product.proto    # Product-related messages
│           └── service.proto    # RPC service definitions
│
├── PLAN.md
└── README.md
```

---

## Implementation Phases

### Phase 1: Project Setup (COMPLETE)
- [x] Create PLAN.md (this file)
- [x] Initialize frontend with Vite + React + TypeScript
- [x] Configure Tailwind CSS
- [x] Initialize Go backend module
- [x] Set up Buf for protobuf management
- [x] Create initial proto definitions
- [x] Generate Go and TypeScript code from protos
- [x] Set up Connect-Go server skeleton
- [x] Set up Connect-Web client in frontend
- [x] Verify end-to-end connectivity with a health check endpoint

### Phase 2: Best Buy API Integration (Backend) (COMPLETE)
- [x] Create Best Buy API client in Go
- [x] Implement store search by postal code/area
- [x] Implement product search by keyword
- [x] Implement product search by SKU
- [x] Implement inventory/availability check
- [x] Add proper error handling and timeouts
- [ ] Test API integration with real Best Buy API key (waiting for API key)

### Phase 3: Store Features
- [ ] Create SearchStores RPC endpoint
- [ ] Build store search UI component
- [ ] Implement store search results display
- [ ] Create "My Stores" context with localStorage persistence
- [ ] Build "Add to My Stores" functionality
- [ ] Build "My Stores List" view with remove capability

### Phase 4: Product Features
- [ ] Create SearchProducts RPC endpoint
- [ ] Build product search UI component
- [ ] Implement product search results display
- [ ] Create "My Products" context with localStorage persistence
- [ ] Build "Add to My Products" functionality
- [ ] Build "My Products List" view with remove capability

### Phase 5: Stock Check Feature
- [ ] Create CheckStock RPC endpoint (bulk check)
- [ ] Build "Check Stock" button UI
- [ ] Implement stock check loading state
- [ ] Build results table component
- [ ] Implement sorting (in-stock items first)
- [ ] Add visual indicators for stock status (in stock, low stock, out of stock)

### Phase 6: Polish & UX
- [ ] Add loading spinners/skeletons
- [ ] Implement error handling and user feedback
- [ ] Add empty state messages
- [ ] Responsive design for mobile
- [ ] Add refresh/re-check capability

---

## Protobuf Service Definition (Draft)

```protobuf
service StockCheckerService {
  // Store operations
  rpc SearchStores(SearchStoresRequest) returns (SearchStoresResponse);

  // Product operations
  rpc SearchProducts(SearchProductsRequest) returns (SearchProductsResponse);

  // Stock check
  rpc CheckStock(CheckStockRequest) returns (CheckStockResponse);
}

message Store {
  string store_id = 1;
  string name = 2;
  string address = 3;
  string city = 4;
  string state = 5;
  string postal_code = 6;
  string phone = 7;
  double distance_miles = 8;
}

message Product {
  string sku = 1;
  string name = 2;
  double sale_price = 3;
  string image_url = 4;
  string product_url = 5;
}

message StockStatus {
  Store store = 1;
  Product product = 2;
  bool in_stock = 3;
  bool low_stock = 4;
  bool pickup_eligible = 5;
}
```

---

## Environment Variables

```bash
# Backend
BESTBUY_API_KEY=your_api_key_here
PORT=8080  # optional, defaults to 8080

# Frontend
VITE_API_URL=http://localhost:8080  # Backend URL
```

---

## Future Features (Backlog)

These are features we may want to add after the initial implementation:

- [ ] **Notifications:** Email/SMS alerts when products come in stock
- [ ] **Auto-refresh:** Periodic background stock checks
- [ ] **Multiple retailers:** Add Target, Walmart, Costco, Sam's Club
- [ ] **Persistent storage:** Database for user preferences (PostgreSQL)
- [ ] **User accounts:** Multi-user support with authentication
- [ ] **Docker deployment:** Containerized setup for easy deployment
- [ ] **Cloud hosting:** Deploy to cloud provider (Railway, Fly.io, etc.)
- [ ] **Price tracking:** Track price history and alert on drops
- [ ] **Product categories:** Filter by Pokemon set/series
- [ ] **Store hours:** Display store hours in results
- [ ] **Map view:** Visual map of stores with stock
- [ ] **History:** Track past stock check results

---

## API Reference

### Best Buy API Endpoints Used

| Feature | Endpoint | Example |
|---------|----------|---------|
| Search stores | `GET /v1/stores(area({zip},{miles}))` | `/v1/stores(area(55423,25))?format=json&apiKey=XXX` |
| Search products | `GET /v1/products(search={term})` | `/v1/products(search=pokemon+cards)?format=json&apiKey=XXX` |
| Get product by SKU | `GET /v1/products/{sku}.json` | `/v1/products/6543219.json?apiKey=XXX` |
| Check availability | `GET /v1/products/{sku}/stores.json` | `/v1/products/6543219/stores.json?postalCode=55423&apiKey=XXX` |

### Key Response Fields

**Store:**
- `storeId`, `name`, `address`, `city`, `state`, `postalCode`, `phone`, `distance`

**Product:**
- `sku`, `name`, `salePrice`, `thumbnailImage`, `url`, `inStoreAvailability`

**Availability:**
- `stores[].storeId`, `stores[].lowStock`, `stores[].distance`

---

## Development Commands

```bash
# Frontend
cd frontend
npm install
npm run dev          # Start dev server (hot reload)
npm run build        # Production build

# Backend
cd backend
go mod tidy
go run cmd/server/main.go   # Start server

# Protobuf generation
cd proto
buf generate         # Generate Go + TypeScript code
```

---

## Resources

- [Best Buy Developer Portal](https://developer.bestbuy.com/)
- [Best Buy API Documentation](https://bestbuyapis.github.io/api-documentation/)
- [Connect-Go Documentation](https://connectrpc.com/docs/go/getting-started)
- [Connect-Web Documentation](https://connectrpc.com/docs/web/getting-started)
- [Buf CLI](https://buf.build/docs/installation)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [Vite](https://vitejs.dev/guide/)
