package oauth

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coltoneshaw/ynab.go/api"
	"github.com/coltoneshaw/ynab.go/api/account"
	"github.com/coltoneshaw/ynab.go/api/budget"
	"github.com/coltoneshaw/ynab.go/api/category"
	"github.com/coltoneshaw/ynab.go/api/month"
	"github.com/coltoneshaw/ynab.go/api/payee"
	"github.com/coltoneshaw/ynab.go/api/transaction"
	"github.com/coltoneshaw/ynab.go/api/user"
)

// OAuthClient is a YNAB client that uses OAuth for authentication
type OAuthClient struct {
	sync.Mutex

	config       *Config
	tokenManager *TokenManager
	httpClient   *api.HTTPClient

	rateLimiter *api.RateLimitTracker

	// Service instances
	user        *user.Service
	budget      *budget.Service
	account     *account.Service
	category    *category.Service
	payee       *payee.Service
	month       *month.Service
	transaction *transaction.Service
}

// NewOAuthClient creates a new OAuth-enabled YNAB client
func NewOAuthClient(config *Config, tokenManager *TokenManager) *OAuthClient {
	client := &OAuthClient{
		config:       config,
		tokenManager: tokenManager,
		httpClient:   api.NewHTTPClient(),
		rateLimiter:  api.NewYNABRateLimitTracker(),
	}

	// Initialize services
	client.user = user.NewService(client)
	client.budget = budget.NewService(client)
	client.account = account.NewService(client)
	client.category = category.NewService(client)
	client.payee = payee.NewService(client)
	client.month = month.NewService(client)
	client.transaction = transaction.NewService(client)

	return client
}

// NewOAuthClientFromToken creates a new OAuth client with an existing token
func NewOAuthClientFromToken(config *Config, token *Token) (*OAuthClient, error) {
	storage := NewMemoryStorage()
	if err := storage.SaveToken(token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	tokenManager := NewTokenManager(config, storage)
	return NewOAuthClient(config, tokenManager), nil
}

// NewOAuthClientFromStorage creates a new OAuth client with token storage
func NewOAuthClientFromStorage(config *Config, storage TokenStorage) (*OAuthClient, error) {
	tokenManager := NewTokenManager(config, storage)
	return NewOAuthClient(config, tokenManager), nil
}

// WithHTTPClient sets a custom HTTP client
func (c *OAuthClient) WithHTTPClient(httpClient *http.Client) api.HTTPClientConfigurer {
	c.httpClient = c.httpClient.WithHTTPClient(httpClient)
	c.tokenManager.WithHTTPClient(httpClient)
	return c
}

// WithTokenRefreshCallback sets a callback for token refresh events
func (c *OAuthClient) WithTokenRefreshCallback(callback func(*Token)) *OAuthClient {
	c.tokenManager.WithTokenRefreshCallback(callback)
	return c
}

// Config returns the OAuth configuration
func (c *OAuthClient) Config() *Config {
	return c.config
}

// TokenManager returns the token manager
func (c *OAuthClient) TokenManager() *TokenManager {
	return c.tokenManager
}

// IsAuthenticated checks if the client has a valid token
func (c *OAuthClient) IsAuthenticated() bool {
	return c.tokenManager.IsAuthenticated()
}

// GetToken returns the current token
func (c *OAuthClient) GetToken(ctx context.Context) (*Token, error) {
	return c.tokenManager.GetToken(ctx)
}

// RefreshToken manually refreshes the token
func (c *OAuthClient) RefreshToken(ctx context.Context) (*Token, error) {
	return c.tokenManager.RefreshToken(ctx)
}

// SetToken sets a new token
func (c *OAuthClient) SetToken(token *Token) error {
	return c.tokenManager.SetToken(token)
}

// ClearToken clears the current token
func (c *OAuthClient) ClearToken() error {
	return c.tokenManager.ClearToken()
}

// Service methods (implementing ClientServicer interface)

// User returns user.Service API instance
func (c *OAuthClient) User() *user.Service {
	return c.user
}

// Budget returns budget.Service API instance
func (c *OAuthClient) Budget() *budget.Service {
	return c.budget
}

// Account returns account.Service API instance
func (c *OAuthClient) Account() *account.Service {
	return c.account
}

// Category returns category.Service API instance
func (c *OAuthClient) Category() *category.Service {
	return c.category
}

// Payee returns payee.Service API instance
func (c *OAuthClient) Payee() *payee.Service {
	return c.payee
}

// Month returns month.Service API instance
func (c *OAuthClient) Month() *month.Service {
	return c.month
}

// Transaction returns transaction.Service API instance
func (c *OAuthClient) Transaction() *transaction.Service {
	return c.transaction
}

// RequestsRemaining returns how many requests can be made before hitting the rate limit
func (c *OAuthClient) RequestsRemaining() int {
	return c.rateLimiter.RequestsRemaining()
}

// TimeUntilReset returns the duration until the oldest request falls out of the rolling window.
// In your scenario: if 200 API calls were made over 50 minutes, this returns ~10 minutes
// (when the oldest request will be 1 hour old and fall off the rolling window).
func (c *OAuthClient) TimeUntilReset() time.Duration {
	return c.rateLimiter.TimeUntilReset()
}

// RequestsInWindow returns the number of requests made in the current rolling window
func (c *OAuthClient) RequestsInWindow() int {
	return c.rateLimiter.RequestsInWindow()
}

// IsAtLimit returns true if the rate limit has been reached
func (c *OAuthClient) IsAtLimit() bool {
	return c.rateLimiter.IsAtLimit()
}

// HTTP methods (implementing api.ClientReaderWriter interface)

// GET sends a GET request to the YNAB API
func (c *OAuthClient) GET(url string, responseModel any) error {
	return c.do(context.Background(), http.MethodGet, url, responseModel, nil)
}

// POST sends a POST request to the YNAB API
func (c *OAuthClient) POST(url string, responseModel any, requestBody []byte) error {
	return c.do(context.Background(), http.MethodPost, url, responseModel, requestBody)
}

// PUT sends a PUT request to the YNAB API
func (c *OAuthClient) PUT(url string, responseModel any, requestBody []byte) error {
	return c.do(context.Background(), http.MethodPut, url, responseModel, requestBody)
}

// PATCH sends a PATCH request to the YNAB API
func (c *OAuthClient) PATCH(url string, responseModel any, requestBody []byte) error {
	return c.do(context.Background(), http.MethodPatch, url, responseModel, requestBody)
}

// DELETE sends a DELETE request to the YNAB API
func (c *OAuthClient) DELETE(url string, responseModel any) error {
	return c.do(context.Background(), http.MethodDelete, url, responseModel, nil)
}

// Context-aware HTTP methods

// GETWithContext sends a GET request with context
func (c *OAuthClient) GETWithContext(ctx context.Context, url string, responseModel any) error {
	return c.do(ctx, http.MethodGet, url, responseModel, nil)
}

// POSTWithContext sends a POST request with context
func (c *OAuthClient) POSTWithContext(ctx context.Context, url string, responseModel any, requestBody []byte) error {
	return c.do(ctx, http.MethodPost, url, responseModel, requestBody)
}

// PUTWithContext sends a PUT request with context
func (c *OAuthClient) PUTWithContext(ctx context.Context, url string, responseModel any, requestBody []byte) error {
	return c.do(ctx, http.MethodPut, url, responseModel, requestBody)
}

// PATCHWithContext sends a PATCH request with context
func (c *OAuthClient) PATCHWithContext(ctx context.Context, url string, responseModel any, requestBody []byte) error {
	return c.do(ctx, http.MethodPatch, url, responseModel, requestBody)
}

// DELETEWithContext sends a DELETE request with context
func (c *OAuthClient) DELETEWithContext(ctx context.Context, url string, responseModel any) error {
	return c.do(ctx, http.MethodDelete, url, responseModel, nil)
}

// do sends a request to the YNAB API with OAuth authentication
func (c *OAuthClient) do(ctx context.Context, method, url string, responseModel any, requestBody []byte) error {
	// Get access token
	accessToken, err := c.tokenManager.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Try the request with current token
	err = c.httpClient.DoRequestWithContext(ctx, method, url, responseModel, requestBody, accessToken)

	// If we get an authentication error, try token refresh once
	if err != nil {
		if apiErr, ok := err.(*api.Error); ok && apiErr.ID == "401" {
			// Try to refresh token
			if _, refreshErr := c.tokenManager.RefreshToken(ctx); refreshErr == nil {
				// Get new access token and retry
				if newAccessToken, tokenErr := c.tokenManager.GetAccessToken(ctx); tokenErr == nil {
					err = c.httpClient.DoRequestWithContext(ctx, method, url, responseModel, requestBody, newAccessToken)
				}
			}
		}
	}

	if err != nil {
		return err
	}

	// Record successful request for rate limiting
	c.rateLimiter.RecordRequest()

	return nil
}

// ClientBuilder helps build OAuth clients with fluent interface
type ClientBuilder struct {
	config               *Config
	storage              TokenStorage
	token                *Token
	httpClient           *http.Client
	tokenRefreshCallback func(*Token)
}

// NewClientBuilder creates a new client builder
func NewClientBuilder(config *Config) *ClientBuilder {
	return &ClientBuilder{
		config: config,
	}
}

// WithStorage sets the token storage
func (b *ClientBuilder) WithStorage(storage TokenStorage) *ClientBuilder {
	b.storage = storage
	return b
}

// WithFileStorage sets file-based token storage
func (b *ClientBuilder) WithFileStorage(filePath string) *ClientBuilder {
	b.storage = NewFileStorage(filePath)
	return b
}

// WithDefaultFileStorage sets default file-based token storage
func (b *ClientBuilder) WithDefaultFileStorage() *ClientBuilder {
	b.storage = NewFileStorage(DefaultTokenPath())
	return b
}

// WithMemoryStorage sets in-memory token storage
func (b *ClientBuilder) WithMemoryStorage() *ClientBuilder {
	b.storage = NewMemoryStorage()
	return b
}

// WithToken sets an initial token
func (b *ClientBuilder) WithToken(token *Token) *ClientBuilder {
	b.token = token
	return b
}

// WithHTTPClient sets a custom HTTP client
func (b *ClientBuilder) WithHTTPClient(httpClient *http.Client) *ClientBuilder {
	b.httpClient = httpClient
	return b
}

// WithTokenRefreshCallback sets a token refresh callback
func (b *ClientBuilder) WithTokenRefreshCallback(callback func(*Token)) *ClientBuilder {
	b.tokenRefreshCallback = callback
	return b
}

// Build creates the OAuth client
func (b *ClientBuilder) Build() (*OAuthClient, error) {
	// Use memory storage if none specified
	if b.storage == nil {
		b.storage = NewMemoryStorage()
	}

	// Create token manager
	tokenManager := NewTokenManager(b.config, b.storage)

	// Set HTTP client if provided
	if b.httpClient != nil {
		tokenManager.WithHTTPClient(b.httpClient)
	}

	// Set token refresh callback if provided
	if b.tokenRefreshCallback != nil {
		tokenManager.WithTokenRefreshCallback(b.tokenRefreshCallback)
	}

	// Create client
	client := NewOAuthClient(b.config, tokenManager)

	// Set HTTP client if provided
	if b.httpClient != nil {
		client.WithHTTPClient(b.httpClient)
	}

	// Set initial token if provided
	if b.token != nil {
		if err := client.SetToken(b.token); err != nil {
			return nil, fmt.Errorf("failed to set initial token: %w", err)
		}
	}

	return client, nil
}
