package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
)

// Note: Migrations are read from the migrations directory at runtime

// DB wraps the database connection
type DB struct {
	*sql.DB
}

// New creates a new database connection
func New(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// RunMigrations runs all SQL migrations
func (db *DB) RunMigrations(migrationsDir string) error {
	// Find migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to find migrations: %w", err)
	}

	for _, file := range files {
		migration, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file, err)
		}

		// Execute migration
		_, err = db.Exec(string(migration))
		if err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file, err)
		}

		log.Printf("Applied migration: %s", filepath.Base(file))
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// User represents a user in the database
type User struct {
	ID         int
	GoogleID   string
	Email      string
	Name       string
	PictureURL string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Store represents a saved store
type Store struct {
	ID         int
	UserID     int
	StoreID    string
	Name       string
	Address    string
	City       string
	State      string
	PostalCode string
	Phone      string
	CreatedAt  time.Time
}

// Product represents a saved product
type Product struct {
	ID           int
	UserID       int
	SKU          string
	Name         string
	SalePrice    float64
	ThumbnailURL string
	ProductURL   string
	CreatedAt    time.Time
}

// Session represents an auth session
type Session struct {
	ID        int
	Token     string
	UserID    int
	ExpiresAt time.Time
	CreatedAt time.Time
}

// IsEmailAllowed checks if an email is in the whitelist
func (db *DB) IsEmailAllowed(ctx context.Context, email string) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM allowed_emails WHERE LOWER(email) = LOWER($1)",
		email,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// AddAllowedEmail adds an email to the whitelist
func (db *DB) AddAllowedEmail(ctx context.Context, email string, addedBy *int) error {
	_, err := db.ExecContext(ctx,
		"INSERT INTO allowed_emails (email, added_by) VALUES (LOWER($1), $2) ON CONFLICT (email) DO NOTHING",
		email, addedBy,
	)
	return err
}

// GetOrCreateUser gets or creates a user by Google ID
func (db *DB) GetOrCreateUser(ctx context.Context, googleID, email, name, pictureURL string) (*User, error) {
	var user User
	err := db.QueryRowContext(ctx,
		`INSERT INTO users (google_id, email, name, picture_url)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (google_id) DO UPDATE SET
		   email = EXCLUDED.email,
		   name = EXCLUDED.name,
		   picture_url = EXCLUDED.picture_url,
		   updated_at = CURRENT_TIMESTAMP
		 RETURNING id, google_id, email, name, picture_url, created_at, updated_at`,
		googleID, email, name, pictureURL,
	).Scan(&user.ID, &user.GoogleID, &user.Email, &user.Name, &user.PictureURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID gets a user by ID
func (db *DB) GetUserByID(ctx context.Context, id int) (*User, error) {
	var user User
	err := db.QueryRowContext(ctx,
		"SELECT id, google_id, email, name, picture_url, created_at, updated_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.GoogleID, &user.Email, &user.Name, &user.PictureURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateSession creates a new session for a user
func (db *DB) CreateSession(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	_, err := db.ExecContext(ctx,
		"INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID, token, expiresAt,
	)
	return err
}

// GetSession gets a valid session by token
func (db *DB) GetSession(ctx context.Context, token string) (*Session, error) {
	var session Session
	err := db.QueryRowContext(ctx,
		"SELECT id, token, user_id, expires_at, created_at FROM sessions WHERE token = $1 AND expires_at > NOW()",
		token,
	).Scan(&session.ID, &session.Token, &session.UserID, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteSession deletes a session by token
func (db *DB) DeleteSession(ctx context.Context, token string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM sessions WHERE token = $1", token)
	return err
}

// CleanExpiredSessions removes expired sessions
func (db *DB) CleanExpiredSessions(ctx context.Context) error {
	_, err := db.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at < NOW()")
	return err
}

// GetUserStores gets all stores for a user
func (db *DB) GetUserStores(ctx context.Context, userID int) ([]Store, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT id, user_id, store_id, name, address, city, state, postal_code, phone, created_at FROM user_stores WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []Store
	for rows.Next() {
		var s Store
		if err := rows.Scan(&s.ID, &s.UserID, &s.StoreID, &s.Name, &s.Address, &s.City, &s.State, &s.PostalCode, &s.Phone, &s.CreatedAt); err != nil {
			return nil, err
		}
		stores = append(stores, s)
	}
	return stores, rows.Err()
}

// AddUserStore adds a store to user's list
func (db *DB) AddUserStore(ctx context.Context, userID int, store Store) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO user_stores (user_id, store_id, name, address, city, state, postal_code, phone)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (user_id, store_id) DO NOTHING`,
		userID, store.StoreID, store.Name, store.Address, store.City, store.State, store.PostalCode, store.Phone,
	)
	return err
}

// RemoveUserStore removes a store from user's list
func (db *DB) RemoveUserStore(ctx context.Context, userID int, storeID string) error {
	_, err := db.ExecContext(ctx,
		"DELETE FROM user_stores WHERE user_id = $1 AND store_id = $2",
		userID, storeID,
	)
	return err
}

// GetUserProducts gets all products for a user
func (db *DB) GetUserProducts(ctx context.Context, userID int) ([]Product, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT id, user_id, sku, name, sale_price, thumbnail_url, product_url, created_at FROM user_products WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.UserID, &p.SKU, &p.Name, &p.SalePrice, &p.ThumbnailURL, &p.ProductURL, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

// AddUserProduct adds a product to user's list
func (db *DB) AddUserProduct(ctx context.Context, userID int, product Product) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO user_products (user_id, sku, name, sale_price, thumbnail_url, product_url)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (user_id, sku) DO NOTHING`,
		userID, product.SKU, product.Name, product.SalePrice, product.ThumbnailURL, product.ProductURL,
	)
	return err
}

// RemoveUserProduct removes a product from user's list
func (db *DB) RemoveUserProduct(ctx context.Context, userID int, sku string) error {
	_, err := db.ExecContext(ctx,
		"DELETE FROM user_products WHERE user_id = $1 AND sku = $2",
		userID, sku,
	)
	return err
}
