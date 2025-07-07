package api

import (
	"fmt"
	"strings"
)

// YNAB API Error Constants
// These constants represent the documented error IDs from the YNAB API
const (
	// 400 Bad Request
	ErrorBadRequest = "400"

	// 401 Unauthorized
	ErrorUnauthorized = "401"

	// 403 Forbidden errors
	ErrorSubscriptionLapsed = "403.1" // Subscription for account has lapsed
	ErrorTrialExpired       = "403.2" // Trial for account has expired
	ErrorUnauthorizedScope  = "403.3" // Access token scope does not allow access
	ErrorDataLimitReached   = "403.4" // Request will exceed data limits

	// 404 Not Found errors
	ErrorNotFound         = "404.1" // Specified URI does not exist
	ErrorResourceNotFound = "404.2" // Requested resource does not exist

	// 409 Conflict
	ErrorConflict = "409" // Resource cannot be saved due to conflict with existing resource

	// 429 Too Many Requests
	ErrorRateLimit = "429" // Too many API requests in a short time period

	// 500 Internal Server Error
	ErrorInternalServer = "500" // Unexpected API error occurred

	// 503 Service Unavailable
	ErrorServiceUnavailable = "503" // API temporarily disabled or request timeout
)

// Error represents an API Error
type Error struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Detail string `json:"detail"`
}

// Error returns the string version of the error
func (e Error) Error() string {
	return fmt.Sprintf("api: error id=%s name=%s detail=%s",
		e.ID, e.Name, e.Detail)
}

// Account/Subscription related error checks

// IsSubscriptionLapsed returns true if the error indicates a lapsed subscription
func (e *Error) IsSubscriptionLapsed() bool {
	return e.ID == ErrorSubscriptionLapsed
}

// IsTrialExpired returns true if the error indicates an expired trial
func (e *Error) IsTrialExpired() bool {
	return e.ID == ErrorTrialExpired
}

// IsAccountError returns true if the error is related to account/subscription issues
func (e *Error) IsAccountError() bool {
	return e.IsSubscriptionLapsed() || e.IsTrialExpired()
}

// Authentication/Authorization related error checks

// IsUnauthorized returns true if the error indicates authentication failure
func (e *Error) IsUnauthorized() bool {
	return e.ID == ErrorUnauthorized
}

// IsUnauthorizedScope returns true if the error indicates insufficient permissions
func (e *Error) IsUnauthorizedScope() bool {
	return e.ID == ErrorUnauthorizedScope
}

// IsAuthenticationError returns true if the error is related to authentication or authorization
func (e *Error) IsAuthenticationError() bool {
	return e.IsUnauthorized() || e.IsUnauthorizedScope()
}

// Resource related error checks

// IsNotFound returns true if the error indicates a resource was not found
func (e *Error) IsNotFound() bool {
	return e.ID == ErrorNotFound || e.ID == ErrorResourceNotFound
}

// IsConflict returns true if the error indicates a resource conflict
func (e *Error) IsConflict() bool {
	return e.ID == ErrorConflict
}

// IsDataLimitReached returns true if the error indicates data limits were exceeded
func (e *Error) IsDataLimitReached() bool {
	return e.ID == ErrorDataLimitReached
}

// Rate limiting error checks

// IsRateLimit returns true if the error indicates rate limiting
func (e *Error) IsRateLimit() bool {
	return e.ID == ErrorRateLimit
}

// Server error checks

// IsInternalServerError returns true if the error indicates a server error
func (e *Error) IsInternalServerError() bool {
	return e.ID == ErrorInternalServer
}

// IsServiceUnavailable returns true if the error indicates service unavailability
func (e *Error) IsServiceUnavailable() bool {
	return e.ID == ErrorServiceUnavailable
}

// General error categorization

// IsClientError returns true if the error is a client error (4xx)
func (e *Error) IsClientError() bool {
	return strings.HasPrefix(e.ID, "4")
}

// IsServerError returns true if the error is a server error (5xx)
func (e *Error) IsServerError() bool {
	return strings.HasPrefix(e.ID, "5")
}

// IsRetryable returns true if the error might be resolved by retrying the request
func (e *Error) IsRetryable() bool {
	return e.IsRateLimit() || e.IsInternalServerError() || e.IsServiceUnavailable()
}

// IsValidationError returns true if the error is related to input validation
func (e *Error) IsValidationError() bool {
	return e.ID == ErrorBadRequest
}

// RequiresUserAction returns true if the error requires user intervention
func (e *Error) RequiresUserAction() bool {
	return e.IsAccountError() || e.IsAuthenticationError() || e.IsDataLimitReached()
}
