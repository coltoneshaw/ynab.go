// Package api implements shared structures and behaviours of
// the API services
package api // import "github.com/coltoneshaw/ynab.go/api"

import (
	"context"
	"net/http"
	"time"
)

// ClientReader contract for a read only client
type ClientReader interface {
	GET(url string, responseModel any) error
}

// ClientWriter contract for a write only client
type ClientWriter interface {
	POST(url string, responseModel any, requestBody []byte) error
	PUT(url string, responseModel any, requestBody []byte) error
	PATCH(url string, responseModel any, requestBody []byte) error
	DELETE(url string, responseModel any) error
}

// ClientReaderWriter contract for a read-write client
type ClientReaderWriter interface {
	ClientReader
	ClientWriter
}

// ContextClientReader contract for a context-aware read only client
type ContextClientReader interface {
	GETWithContext(ctx context.Context, url string, responseModel any) error
}

// ContextClientWriter contract for a context-aware write only client
type ContextClientWriter interface {
	POSTWithContext(ctx context.Context, url string, responseModel any, requestBody []byte) error
	PUTWithContext(ctx context.Context, url string, responseModel any, requestBody []byte) error
	PATCHWithContext(ctx context.Context, url string, responseModel any, requestBody []byte) error
	DELETEWithContext(ctx context.Context, url string, responseModel any) error
}

// ContextClientReaderWriter contract for a context-aware read-write client
type ContextClientReaderWriter interface {
	ContextClientReader
	ContextClientWriter
}

// FullClient contract for a client that supports both context-aware and regular methods
type FullClient interface {
	ClientReaderWriter
	ContextClientReaderWriter
}

// RateLimiter contract for rate limiting functionality
type RateLimiter interface {
	RequestsRemaining() int
	TimeUntilReset() time.Duration
	RequestsInWindow() int
	IsAtLimit() bool
}

// HTTPClientConfigurer contract for HTTP client configuration
type HTTPClientConfigurer interface {
	WithHTTPClient(*http.Client) HTTPClientConfigurer
}
