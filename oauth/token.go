package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
)

// TokenManager handles token refresh and management
type TokenManager struct {
	config  *Config
	client  *http.Client
	storage TokenStorage

	// Token access mutex
	mu    sync.RWMutex
	token *Token

	// Callback for token refresh events
	onTokenRefresh func(*Token)
}

// NewTokenManager creates a new token manager
func NewTokenManager(config *Config, storage TokenStorage) *TokenManager {
	return &TokenManager{
		config:  config,
		client:  http.DefaultClient,
		storage: storage,
	}
}

// WithHTTPClient sets a custom HTTP client
func (tm *TokenManager) WithHTTPClient(client *http.Client) *TokenManager {
	tm.client = client
	return tm
}

// WithTokenRefreshCallback sets a callback for token refresh events
func (tm *TokenManager) WithTokenRefreshCallback(callback func(*Token)) *TokenManager {
	tm.onTokenRefresh = callback
	return tm
}

// SetToken sets the current token
func (tm *TokenManager) SetToken(token *Token) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.token = token

	if tm.storage != nil {
		return tm.storage.SaveToken(token)
	}

	return nil
}

// GetToken returns the current token, refreshing if necessary
func (tm *TokenManager) GetToken(ctx context.Context) (*Token, error) {
	tm.mu.RLock()
	currentToken := tm.token
	tm.mu.RUnlock()

	// If no token is loaded, try to load from storage
	if currentToken == nil && tm.storage != nil {
		loadedToken, err := tm.storage.LoadToken()
		if err == nil && loadedToken != nil {
			tm.mu.Lock()
			tm.token = loadedToken
			currentToken = loadedToken
			tm.mu.Unlock()
		}
	}

	// If still no token, return error
	if currentToken == nil {
		return nil, fmt.Errorf("no token available")
	}

	// If token is valid, return it
	if currentToken.IsValid() {
		return currentToken, nil
	}

	// If token is expired but can't be refreshed, return error
	if !currentToken.CanRefresh() {
		return nil, ErrTokenExpired
	}

	// Refresh the token
	refreshedToken, err := tm.refreshToken(ctx, currentToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update the stored token
	if err := tm.SetToken(refreshedToken); err != nil {
		return nil, fmt.Errorf("failed to save refreshed token: %w", err)
	}

	// Call refresh callback if set
	if tm.onTokenRefresh != nil {
		tm.onTokenRefresh(refreshedToken)
	}

	return refreshedToken, nil
}

// ExchangeCode exchanges an authorization code for an access token
func (tm *TokenManager) ExchangeCode(ctx context.Context, code string) (*Token, error) {
	if err := tm.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	tokenRequest := &TokenRequest{
		GrantType:    GrantTypeAuthorizationCode,
		ClientID:     tm.config.ClientID,
		ClientSecret: tm.config.ClientSecret,
		Code:         code,
		RedirectURI:  tm.config.RedirectURI,
	}

	return tm.exchangeToken(ctx, tokenRequest)
}

// RefreshToken refreshes the current token
func (tm *TokenManager) RefreshToken(ctx context.Context) (*Token, error) {
	tm.mu.RLock()
	currentToken := tm.token
	tm.mu.RUnlock()

	if currentToken == nil {
		return nil, fmt.Errorf("no token to refresh")
	}

	if !currentToken.CanRefresh() {
		return nil, fmt.Errorf("token cannot be refreshed")
	}

	refreshedToken, err := tm.refreshToken(ctx, currentToken)
	if err != nil {
		return nil, err
	}

	if err := tm.SetToken(refreshedToken); err != nil {
		return nil, fmt.Errorf("failed to save refreshed token: %w", err)
	}

	return refreshedToken, nil
}

// refreshToken performs the actual token refresh
func (tm *TokenManager) refreshToken(ctx context.Context, token *Token) (*Token, error) {
	tokenRequest := &TokenRequest{
		GrantType:    GrantTypeRefreshToken,
		ClientID:     tm.config.ClientID,
		ClientSecret: tm.config.ClientSecret,
		RefreshToken: token.RefreshToken,
	}

	return tm.exchangeToken(ctx, tokenRequest)
}

// exchangeToken performs the token exchange with YNAB
func (tm *TokenManager) exchangeToken(ctx context.Context, tokenRequest *TokenRequest) (*Token, error) {
	// Prepare form data
	data := url.Values{}
	data.Set("grant_type", string(tokenRequest.GrantType))
	data.Set("client_id", tokenRequest.ClientID)
	data.Set("client_secret", tokenRequest.ClientSecret)

	if tokenRequest.Code != "" {
		data.Set("code", tokenRequest.Code)
		data.Set("redirect_uri", tokenRequest.RedirectURI)
	}

	if tokenRequest.RefreshToken != "" {
		data.Set("refresh_token", tokenRequest.RefreshToken)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tm.config.tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := tm.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Check for error in response
	if tokenResponse.Error != "" {
		return nil, &ErrorResponse{
			ErrorCode:        tokenResponse.Error,
			ErrorDescription: tokenResponse.ErrorDescription,
		}
	}

	// Validate response
	if tokenResponse.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response")
	}

	// Convert to Token
	token := tokenResponse.ToToken()

	// Set default expiration if not provided (YNAB tokens typically last 2 hours)
	if token.ExpiresIn == 0 {
		token.SetExpiration(7200) // 2 hours
	}

	return token, nil
}

// ClearToken removes the current token
func (tm *TokenManager) ClearToken() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.token = nil

	if tm.storage != nil {
		return tm.storage.ClearToken()
	}

	return nil
}

// IsAuthenticated checks if there's a valid token available
func (tm *TokenManager) IsAuthenticated() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.token != nil && tm.token.IsValid()
}

// GetAccessToken returns just the access token string if available
func (tm *TokenManager) GetAccessToken(ctx context.Context) (string, error) {
	token, err := tm.GetToken(ctx)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

// TokenSource creates a token source for use with oauth2 compatible libraries
type TokenSource struct {
	manager *TokenManager
	ctx     context.Context
}

// NewTokenSource creates a new token source
func NewTokenSource(ctx context.Context, manager *TokenManager) *TokenSource {
	return &TokenSource{
		manager: manager,
		ctx:     ctx,
	}
}

// Token returns the current token, implementing oauth2.TokenSource interface
func (ts *TokenSource) Token() (*Token, error) {
	return ts.manager.GetToken(ts.ctx)
}

// AuthenticatedTransport creates an HTTP transport that automatically adds Bearer tokens
type AuthenticatedTransport struct {
	Base    http.RoundTripper
	manager *TokenManager
}

// NewAuthenticatedTransport creates a new authenticated transport
func NewAuthenticatedTransport(manager *TokenManager) *AuthenticatedTransport {
	return &AuthenticatedTransport{
		Base:    http.DefaultTransport,
		manager: manager,
	}
}

// RoundTrip implements http.RoundTripper
func (t *AuthenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqCopy := req.Clone(req.Context())

	// Get access token
	accessToken, err := t.manager.GetAccessToken(req.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Add Authorization header
	reqCopy.Header.Set("Authorization", "Bearer "+accessToken)

	// Execute request
	resp, err := t.Base.RoundTrip(reqCopy)

	// If we get a 401, try refreshing the token once
	if err == nil && resp.StatusCode == http.StatusUnauthorized {
		// Try to refresh token
		if _, refreshErr := t.manager.RefreshToken(req.Context()); refreshErr == nil {
			// Get new access token
			if newAccessToken, tokenErr := t.manager.GetAccessToken(req.Context()); tokenErr == nil {
				// Retry the request with new token
				_ = resp.Body.Close() // Close the original response

				reqRetry := req.Clone(req.Context())
				reqRetry.Header.Set("Authorization", "Bearer "+newAccessToken)
				return t.Base.RoundTrip(reqRetry)
			}
		}
	}

	return resp, err
}
