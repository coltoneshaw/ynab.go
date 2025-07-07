package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const APIEndpoint = "https://api.youneedabudget.com/v1"

// HTTPClient represents a configurable HTTP client
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTP client with default configuration
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: http.DefaultClient,
	}
}

// NewHTTPClientWithClient creates a new HTTP client with custom http.Client
func NewHTTPClientWithClient(client *http.Client) *HTTPClient {
	return &HTTPClient{
		client: client,
	}
}

// WithHTTPClient sets a custom HTTP client
func (h *HTTPClient) WithHTTPClient(client *http.Client) *HTTPClient {
	h.client = client
	return h
}

// PrepareRequest prepares an HTTP request with common headers
func (h *HTTPClient) PrepareRequest(ctx context.Context, method, url string, requestBody []byte) (*http.Request, error) {
	fullURL := fmt.Sprintf("%s%s", APIEndpoint, url)

	var bodyReader io.Reader
	if requestBody != nil {
		bodyReader = bytes.NewBuffer(requestBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set common headers
	req.Header.Set("Accept", "application/json")
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// SetAuthorizationHeader sets the Authorization header with Bearer token
func (h *HTTPClient) SetAuthorizationHeader(req *http.Request, accessToken string) {
	req.Header.Set("Authorization", "Bearer "+accessToken)
}

// ExecuteRequest sends the HTTP request and returns the response
func (h *HTTPClient) ExecuteRequest(req *http.Request) (*http.Response, error) {
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

// HandleResponse processes the HTTP response and handles errors
func (h *HTTPClient) HandleResponse(resp *http.Response, responseModel any) error {
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		response := struct {
			Error *Error `json:"error"`
		}{}

		if err := json.Unmarshal(body, &response); err != nil {
			// Return a forged *Error for ease of use
			apiError := &Error{
				ID:     strconv.Itoa(resp.StatusCode),
				Name:   "unknown_api_error",
				Detail: "Unknown API error",
			}
			return apiError
		}

		return response.Error
	}

	// Parse successful response
	if responseModel != nil {
		if err := json.Unmarshal(body, responseModel); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// DoRequest performs a complete HTTP request with error handling
func (h *HTTPClient) DoRequest(ctx context.Context, method, url string, responseModel any, requestBody []byte, accessToken string) error {
	req, err := h.PrepareRequest(ctx, method, url, requestBody)
	if err != nil {
		return err
	}

	h.SetAuthorizationHeader(req, accessToken)

	resp, err := h.ExecuteRequest(req)
	if err != nil {
		return err
	}

	return h.HandleResponse(resp, responseModel)
}

// DoRequestWithContext performs a complete HTTP request with context
func (h *HTTPClient) DoRequestWithContext(ctx context.Context, method, url string, responseModel any, requestBody []byte, accessToken string) error {
	return h.DoRequest(ctx, method, url, responseModel, requestBody, accessToken)
}
