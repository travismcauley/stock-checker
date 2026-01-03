package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tmcauley/stock-checker/backend/internal/database"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	SessionCookieName = "session_token"
	SessionDuration   = 7 * 24 * time.Hour // 7 days
)

// GoogleUserInfo represents the user info from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// Auth handles authentication
type Auth struct {
	db           *database.DB
	oauthConfig  *oauth2.Config
	frontendURL  string
	secureCookie bool
}

// New creates a new Auth handler
func New(db *database.DB, clientID, clientSecret, redirectURL, frontendURL string, secureCookie bool) *Auth {
	return &Auth{
		db: db,
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		frontendURL:  frontendURL,
		secureCookie: secureCookie,
	}
}

// generateToken generates a random token
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// HandleLogin redirects to Google OAuth
func (a *Auth) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate state token to prevent CSRF
	state, err := generateToken()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Store state in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		Secure:   a.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to Google
	url := a.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleCallback handles the OAuth callback from Google
func (a *Auth) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Verify state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Exchange code for token
	code := r.URL.Query().Get("code")
	token, err := a.oauthConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Get user info from Google
	userInfo, err := a.getUserInfo(ctx, token)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Check if email is allowed
	allowed, err := a.db.IsEmailAllowed(ctx, userInfo.Email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !allowed {
		// Redirect to frontend with error
		http.Redirect(w, r, a.frontendURL+"?error=not_allowed", http.StatusTemporaryRedirect)
		return
	}

	// Create or update user
	user, err := a.db.GetOrCreateUser(ctx, userInfo.ID, userInfo.Email, userInfo.Name, userInfo.Picture)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Create session
	sessionToken, err := generateToken()
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(SessionDuration)
	if err := a.db.CreateSession(ctx, user.ID, sessionToken, expiresAt); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   a.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to frontend
	http.Redirect(w, r, a.frontendURL, http.StatusTemporaryRedirect)
}

// HandleLogout logs out the user
func (a *Auth) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get session cookie
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil {
		// Delete session from database
		_ = a.db.DeleteSession(r.Context(), cookie.Value)
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   a.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to frontend
	http.Redirect(w, r, a.frontendURL, http.StatusTemporaryRedirect)
}

// getUserInfo fetches user info from Google
func (a *Auth) getUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := a.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// GetUserFromRequest gets the current user from the request
func (a *Auth) GetUserFromRequest(r *http.Request) (*database.User, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return nil, fmt.Errorf("no session cookie")
	}

	session, err := a.db.GetSession(r.Context(), cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid session")
	}

	user, err := a.db.GetUserByID(r.Context(), session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// Middleware returns an auth middleware that requires authentication
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.GetUserFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Context key for user
type contextKey string

const userContextKey contextKey = "user"

// UserFromContext gets the user from context
func UserFromContext(ctx context.Context) *database.User {
	user, _ := ctx.Value(userContextKey).(*database.User)
	return user
}
