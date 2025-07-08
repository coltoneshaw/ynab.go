package api

import (
	"net/http"
	"time"
)

// ClientReader defines the interface for read-only HTTP operations
type ClientReader interface {
	GET(url string, responseModel any) error
}

// ClientWriter defines the interface for write HTTP operations
type ClientWriter interface {
	POST(url string, responseModel any, requestBody []byte) error
	PUT(url string, responseModel any, requestBody []byte) error
	PATCH(url string, responseModel any, requestBody []byte) error
	DELETE(url string, responseModel any) error
}

// ClientReaderWriter combines read and write operations
type ClientReaderWriter interface {
	ClientReader
	ClientWriter
}

// RateLimiter defines the interface for rate limiting functionality
type RateLimiter interface {
	RequestsRemaining() int
	RequestsInWindow() int
	TimeUntilReset() time.Duration
	IsAtLimit() bool
}

// HTTPClientConfigurer defines the interface for HTTP client configuration
type HTTPClientConfigurer interface {
	WithHTTPClient(client *http.Client) HTTPClientConfigurer
}
