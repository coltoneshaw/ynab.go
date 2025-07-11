// Package ynab implements the client API
package ynab // import "github.com/coltoneshaw/ynab.go"

import (
	"context"
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
	"github.com/coltoneshaw/ynab.go/oauth"
)

const apiEndpoint = "https://api.youneedabudget.com/v1"

// ClientServicer contract for a client service API
type ClientServicer interface {
	User() *user.Service
	Budget() *budget.Service
	Account() *account.Service
	Category() *category.Service
	Payee() *payee.Service
	Month() *month.Service
	Transaction() *transaction.Service

	// Rate limiting interface
	api.RateLimiter

	// HTTP client configuration interface
	api.HTTPClientConfigurer

	// Token management interface
	api.TokenProvider
}

// NewClient facilitates the creation of a new client instance with a static token
func NewClient(accessToken string) ClientServicer {
	tokenProvider := api.NewStaticTokenProvider(accessToken)
	return NewClientWithTokenProvider(tokenProvider)
}

// NewClientWithTokenProvider creates a new client with a custom token provider
func NewClientWithTokenProvider(tokenProvider api.TokenProvider) ClientServicer {
	c := &client{
		tokenProvider: tokenProvider,
		httpClient:    api.NewHTTPClient(),
		rateLimiter:   api.NewYNABRateLimitTracker(),
	}

	c.user = user.NewService(c)
	c.budget = budget.NewService(c)
	c.account = account.NewService(c)
	c.category = category.NewService(c)
	c.payee = payee.NewService(c)
	c.month = month.NewService(c)
	c.transaction = transaction.NewService(c)
	return c
}

// client API
type client struct {
	sync.Mutex

	tokenProvider api.TokenProvider

	httpClient *api.HTTPClient

	rateLimiter *api.RateLimitTracker

	user        *user.Service
	budget      *budget.Service
	account     *account.Service
	category    *category.Service
	payee       *payee.Service
	month       *month.Service
	transaction *transaction.Service
}

// WithHTTPClient sets a custom HTTP client and returns the client for chaining
func (c *client) WithHTTPClient(httpClient *http.Client) api.HTTPClientConfigurer {
	c.httpClient = c.httpClient.WithHTTPClient(httpClient)
	return c
}

// User returns user.Service API instance
func (c *client) User() *user.Service {
	return c.user
}

// Budget returns budget.Service API instance
func (c *client) Budget() *budget.Service {
	return c.budget
}

// Account returns account.Service API instance
func (c *client) Account() *account.Service {
	return c.account
}

// Category returns category.Service API instance
func (c *client) Category() *category.Service {
	return c.category
}

// Payee returns payee.Service API instance
func (c *client) Payee() *payee.Service {
	return c.payee
}

// Month returns month.Service API instance
func (c *client) Month() *month.Service {
	return c.month
}

// Transaction returns transaction.Service API instance
func (c *client) Transaction() *transaction.Service {
	return c.transaction
}

// RequestsRemaining returns how many requests can be made before hitting the rate limit
func (c *client) RequestsRemaining() int {
	return c.rateLimiter.RequestsRemaining()
}

// TimeUntilReset returns the duration until the oldest request falls out of the rolling window.
// In your scenario: if 200 API calls were made over 50 minutes, this returns ~10 minutes
// (when the oldest request will be 1 hour old and fall off the rolling window).
func (c *client) TimeUntilReset() time.Duration {
	return c.rateLimiter.TimeUntilReset()
}

// RequestsInWindow returns the number of requests made in the current rolling window
func (c *client) RequestsInWindow() int {
	return c.rateLimiter.RequestsInWindow()
}

// IsAtLimit returns true if the rate limit has been reached
func (c *client) IsAtLimit() bool {
	return c.rateLimiter.IsAtLimit()
}

// Token management methods

// SetAccessToken updates the access token for hot-swapping at runtime
func (c *client) SetAccessToken(token string) error {
	return c.tokenProvider.SetAccessToken(token)
}

// GetAccessToken returns the current access token
func (c *client) GetAccessToken(ctx context.Context) (string, error) {
	return c.tokenProvider.GetAccessToken(ctx)
}

// GetAccessTokenString returns the current access token without context
func (c *client) GetAccessTokenString() string {
	return c.tokenProvider.GetAccessTokenString()
}

// IsAuthenticated returns true if the client has a valid token
func (c *client) IsAuthenticated() bool {
	return c.tokenProvider.IsAuthenticated()
}

// GET sends a GET request to the YNAB API
func (c *client) GET(url string, responseModel any) error {
	return c.do(http.MethodGet, url, responseModel, nil)
}

// POST sends a POST request to the YNAB API
func (c *client) POST(url string, responseModel any, requestBody []byte) error {
	return c.do(http.MethodPost, url, responseModel, requestBody)
}

// PUT sends a PUT request to the YNAB API
func (c *client) PUT(url string, responseModel any, requestBody []byte) error {
	return c.do(http.MethodPut, url, responseModel, requestBody)
}

// PATCH sends a PATCH request to the YNAB API
func (c *client) PATCH(url string, responseModel any, requestBody []byte) error {
	return c.do(http.MethodPatch, url, responseModel, requestBody)
}

// DELETE sends a DELETE request to the YNAB API
func (c *client) DELETE(url string, responseModel any) error {
	return c.do(http.MethodDelete, url, responseModel, nil)
}

// do sends a request to the YNAB API
func (c *client) do(method, url string, responseModel any, requestBody []byte) error {
	token, err := c.tokenProvider.GetAccessToken(context.Background())
	if err != nil {
		return err
	}

	err = c.httpClient.DoRequest(context.Background(), method, url, responseModel, requestBody, token)
	if err != nil {
		return err
	}

	// Record successful request for rate limiting
	c.rateLimiter.RecordRequest()

	return nil
}

// OAuth convenience functions

// NewOAuthConfig creates a new OAuth configuration
func NewOAuthConfig(clientID, clientSecret, redirectURI string) *oauth.Config {
	return oauth.NewOAuthConfig(oauth.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
	})
}

// NewOAuthClient creates a new OAuth-enabled YNAB client using the unified client
func NewOAuthClient(config *oauth.Config, tokenManager *oauth.TokenManager) ClientServicer {
	tokenProvider := api.NewOAuthTokenProvider(tokenManager)
	return NewClientWithTokenProvider(tokenProvider)
}

// NewLegacyOAuthClient creates the legacy OAuth client (for backward compatibility if needed)
func NewLegacyOAuthClient(config *oauth.Config, tokenManager *oauth.TokenManager) *oauth.OAuthClient {
	return oauth.NewOAuthClient(config, tokenManager)
}

// NewOAuthClientFromToken creates a new OAuth client with an existing token using the unified client
func NewOAuthClientFromToken(config *oauth.Config, token *oauth.Token) (ClientServicer, error) {
	storage := oauth.NewMemoryStorage()
	if err := storage.SaveToken(token); err != nil {
		return nil, err
	}
	tokenManager := oauth.NewTokenManager(config, storage)
	return NewOAuthClient(config, tokenManager), nil
}

// NewOAuthClientFromStorage creates a new OAuth client with token storage using the unified client
func NewOAuthClientFromStorage(config *oauth.Config, storage oauth.TokenStorage) (ClientServicer, error) {
	tokenManager := oauth.NewTokenManager(config, storage)
	return NewOAuthClient(config, tokenManager), nil
}

// Legacy OAuth convenience functions (for backward compatibility)

// NewLegacyOAuthClientFromToken creates a legacy OAuth client with an existing token
func NewLegacyOAuthClientFromToken(config *oauth.Config, token *oauth.Token) (*oauth.OAuthClient, error) {
	return oauth.NewOAuthClientFromToken(config, token)
}

// NewLegacyOAuthClientFromStorage creates a legacy OAuth client with token storage
func NewLegacyOAuthClientFromStorage(config *oauth.Config, storage oauth.TokenStorage) (*oauth.OAuthClient, error) {
	return oauth.NewOAuthClientFromStorage(config, storage)
}

// NewOAuthClientBuilder creates a new OAuth client builder
func NewOAuthClientBuilder(config *oauth.Config) *oauth.ClientBuilder {
	return oauth.NewClientBuilder(config)
}

// NewAuthorizationCodeFlow creates a new authorization code flow
func NewAuthorizationCodeFlow(config *oauth.Config) *oauth.AuthorizationCodeFlow {
	return oauth.NewAuthorizationCodeFlow(config)
}

// NewImplicitGrantFlow creates a new implicit grant flow
func NewImplicitGrantFlow(config *oauth.Config) *oauth.ImplicitGrantFlow {
	return oauth.NewImplicitGrantFlow(config)
}

// NewFlowManager creates a new OAuth flow manager
func NewFlowManager(config *oauth.Config) *oauth.FlowManager {
	return oauth.NewFlowManager(config)
}

// NewTokenManager creates a new token manager
func NewTokenManager(config *oauth.Config, storage oauth.TokenStorage) *oauth.TokenManager {
	return oauth.NewTokenManager(config, storage)
}

// Storage convenience functions

// NewFileStorage creates a new file-based token storage
func NewFileStorage(filePath string) oauth.TokenStorage {
	return oauth.NewFileStorage(filePath)
}

// NewMemoryStorage creates a new in-memory token storage
func NewMemoryStorage() oauth.TokenStorage {
	return oauth.NewMemoryStorage()
}

// DefaultTokenPath returns the default token file path
func DefaultTokenPath() string {
	return oauth.DefaultTokenPath()
}
