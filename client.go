// Copyright (c) 2018, Bruno M V Souza <github@b.bmvs.io>. All rights reserved.
// Use of this source code is governed by a BSD-2-Clause license that can be
// found in the LICENSE file.

// Package ynab implements the client API
package ynab // import "github.com/brunomvsouza/ynab.go"

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/brunomvsouza/ynab.go/api"
	"github.com/brunomvsouza/ynab.go/api/account"
	"github.com/brunomvsouza/ynab.go/api/budget"
	"github.com/brunomvsouza/ynab.go/api/category"
	"github.com/brunomvsouza/ynab.go/api/month"
	"github.com/brunomvsouza/ynab.go/api/payee"
	"github.com/brunomvsouza/ynab.go/api/transaction"
	"github.com/brunomvsouza/ynab.go/api/user"
	"github.com/brunomvsouza/ynab.go/oauth"
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
}

// NewClient facilitates the creation of a new client instance
func NewClient(accessToken string) ClientServicer {
	c := &client{
		accessToken: accessToken,
		httpClient:  api.NewHTTPClient(),
		rateLimiter: api.NewYNABRateLimitTracker(),
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

	accessToken string

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
	err := c.httpClient.DoRequest(context.Background(), method, url, responseModel, requestBody, c.accessToken)
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
	return oauth.NewConfig(clientID, clientSecret, redirectURI)
}

// NewOAuthClient creates a new OAuth-enabled YNAB client
func NewOAuthClient(config *oauth.Config, tokenManager *oauth.TokenManager) *oauth.OAuthClient {
	return oauth.NewOAuthClient(config, tokenManager)
}

// NewOAuthClientFromToken creates a new OAuth client with an existing token
func NewOAuthClientFromToken(config *oauth.Config, token *oauth.Token) (*oauth.OAuthClient, error) {
	return oauth.NewOAuthClientFromToken(config, token)
}

// NewOAuthClientFromStorage creates a new OAuth client with token storage
func NewOAuthClientFromStorage(config *oauth.Config, storage oauth.TokenStorage) (*oauth.OAuthClient, error) {
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
