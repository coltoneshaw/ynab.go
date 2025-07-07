package ynab

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/coltoneshaw/ynab.go/api"
	"github.com/stretchr/testify/assert"
	"gopkg.in/jarcoal/httpmock.v1"
)

func TestClient_GET(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "application/json", req.Header.Get("Accept"))
				assert.Equal(t, "Bearer 6zL9vh8]B9H3BEecwL%Vzh^VwKR3C2CNZ3Bv%=fFxm$z)duY[U+2=3CydZrkQFnA", req.Header.Get("Authorization"))

				res := httpmock.NewStringResponse(http.StatusOK, `{"foo":"bar"}`)
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("6zL9vh8]B9H3BEecwL%Vzh^VwKR3C2CNZ3Bv%=fFxm$z)duY[U+2=3CydZrkQFnA")
		err := c.(*client).GET("/foo", &response)
		assert.NoError(t, err)
		assert.Equal(t, "bar", response.Foo)
	})

	t.Run("failure with with expected API error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				res := httpmock.NewStringResponse(http.StatusBadRequest, `{
  "error": {
    "id": "400",
    "name": "error_name",
    "detail": "Error detail"
  }
}`)
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).GET("/foo", &response)
		expectedErrStr := "api: error id=400 name=error_name detail=Error detail"
		assert.EqualError(t, err, expectedErrStr)
	})

	t.Run("failure with rate limit error (429)", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				res := httpmock.NewStringResponse(http.StatusTooManyRequests, `{
  "error": {
    "id": "429",
    "name": "too_many_requests",
    "detail": "Too many requests"
  }
}`)
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).GET("/foo", &response)
		expectedErrStr := "api: error id=429 name=too_many_requests detail=Too many requests"
		assert.EqualError(t, err, expectedErrStr)

		// Test that we can detect rate limiting errors
		if apiErr, ok := err.(*api.Error); ok {
			assert.Equal(t, "429", apiErr.ID)
			assert.Equal(t, "too_many_requests", apiErr.Name)
			assert.Equal(t, "Too many requests", apiErr.Detail)
		} else {
			t.Fatal("Expected api.Error type")
		}
	})

	t.Run("failure with with unexpected API error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(http.StatusInternalServerError, "Internal Server Error"), nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).GET("/foo", &response)
		expectedErrStr := "api: error id=500 name=unknown_api_error detail=Unknown API error"
		assert.EqualError(t, err, expectedErrStr)
	})

	t.Run("silent failure due to invalid response model", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodGet, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				res := httpmock.NewStringResponse(http.StatusOK, `{"bar":"foo"}`)
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).GET("/foo", &response)
		assert.NoError(t, err)
		assert.Equal(t, struct {
			Foo string `json:"foo"`
		}{}, response)
	})
}

func TestClient_POST(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPost, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				buf, err := io.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, `{"bar":"foo"}`, string(buf))
				assert.Equal(t, "application/json", req.Header.Get("Accept"))
				assert.Equal(t, "Bearer 6zL9vh8]B9H3BEecwL%Vzh^VwKR3C2CNZ3Bv%=fFxm$z)duY[U+2=3CydZrkQFnA", req.Header.Get("Authorization"))

				res := httpmock.NewStringResponse(http.StatusOK, string(buf))
				return res, nil
			},
		)

		response := struct {
			Bar string `json:"bar"`
		}{}

		c := NewClient("6zL9vh8]B9H3BEecwL%Vzh^VwKR3C2CNZ3Bv%=fFxm$z)duY[U+2=3CydZrkQFnA")
		err := c.(*client).POST("/foo", &response, []byte(`{"bar":"foo"}`))
		assert.NoError(t, err)
		assert.Equal(t, "foo", response.Bar)
	})

	t.Run("failure with with expected API error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPost, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				res := httpmock.NewStringResponse(http.StatusBadRequest, `{
  "error": {
    "id": "400",
    "name": "error_name",
    "detail": "Error detail"
  }
}`)
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).POST("/foo", &response, []byte(`{"bar":"foo"}`))
		expectedErrStr := "api: error id=400 name=error_name detail=Error detail"
		assert.EqualError(t, err, expectedErrStr)
	})

	t.Run("failure with with unexpected API error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPost, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(http.StatusInternalServerError, "Internal Server Error"), nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).POST("/foo", &response, []byte(`{"bar":"foo"}`))
		expectedErrStr := "api: error id=500 name=unknown_api_error detail=Unknown API error"
		assert.EqualError(t, err, expectedErrStr)
	})

	t.Run("silent failure due to invalid response model", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPost, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				res := httpmock.NewStringResponse(http.StatusOK, `{"bar":"foo"}`)
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).POST("/foo", &response, []byte(`{"bar":"foo"}`))
		assert.NoError(t, err)
		assert.Equal(t, struct {
			Foo string `json:"foo"`
		}{}, response)
	})

	t.Run("regression test existence of request header content-type = application/json", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPost, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, req.Header.Get("Content-Type"), "application/json")
				res := httpmock.NewStringResponse(http.StatusOK, `{"bar":"foo"}`)
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).POST("/foo", &response, []byte(`{"bar":"foo"}`))
		assert.NoError(t, err)
		assert.Equal(t, struct {
			Foo string `json:"foo"`
		}{}, response)
	})
}

func TestClient_PUT(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPut, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				buf, err := io.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, `{"bar":"foo"}`, string(buf))
				assert.Equal(t, "application/json", req.Header.Get("Accept"))
				assert.Equal(t, "Bearer 6zL9vh8]B9H3BEecwL%Vzh^VwKR3C2CNZ3Bv%=fFxm$z)duY[U+2=3CydZrkQFnA", req.Header.Get("Authorization"))

				res := httpmock.NewStringResponse(http.StatusOK, string(buf))
				return res, nil
			},
		)

		response := struct {
			Bar string `json:"bar"`
		}{}

		c := NewClient("6zL9vh8]B9H3BEecwL%Vzh^VwKR3C2CNZ3Bv%=fFxm$z)duY[U+2=3CydZrkQFnA")
		err := c.(*client).PUT("/foo", &response, []byte(`{"bar":"foo"}`))
		assert.NoError(t, err)
		assert.Equal(t, "foo", response.Bar)
	})

	t.Run("failure with with expected API error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPut, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				res := httpmock.NewStringResponse(http.StatusBadRequest, `{
  "error": {
    "id": "400",
    "name": "error_name",
    "detail": "Error detail"
  }
}`)
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).PUT("/foo", &response, []byte(`{"bar":"foo"}`))
		expectedErrStr := "api: error id=400 name=error_name detail=Error detail"
		assert.EqualError(t, err, expectedErrStr)
	})

	t.Run("failure with with unexpected API error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPut, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(http.StatusInternalServerError, "Internal Server Error"), nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).PUT("/foo", &response, []byte(`{"bar":"foo"}`))
		expectedErrStr := "api: error id=500 name=unknown_api_error detail=Unknown API error"
		assert.EqualError(t, err, expectedErrStr)
	})

	t.Run("silent failure due to invalid response model", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPut, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				res := httpmock.NewStringResponse(http.StatusOK, `{"bar":"foo"}`)
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).PUT("/foo", &response, []byte(`{"bar":"foo"}`))
		assert.NoError(t, err)
		assert.Equal(t, struct {
			Foo string `json:"foo"`
		}{}, response)
	})

	t.Run("regression test existence of request header content-type = application/json", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPut, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, req.Header.Get("Content-Type"), "application/json")
				res := httpmock.NewStringResponse(http.StatusOK, `{"bar":"foo"}`)
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).PUT("/foo", &response, []byte(`{"bar":"foo"}`))
		assert.NoError(t, err)
		assert.Equal(t, struct {
			Foo string `json:"foo"`
		}{}, response)
	})
}

func TestClient_PATCH(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPatch, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				buf, err := io.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.Equal(t, `{"bar":"foo"}`, string(buf))
				assert.Equal(t, "application/json", req.Header.Get("Accept"))
				assert.Equal(t, "Bearer 6zL9vh8]B9H3BEecwL%Vzh^VwKR3C2CNZ3Bv%=fFxm$z)duY[U+2=3CydZrkQFnA", req.Header.Get("Authorization"))

				res := httpmock.NewStringResponse(http.StatusOK, string(buf))
				return res, nil
			},
		)

		response := struct {
			Bar string `json:"bar"`
		}{}

		c := NewClient("6zL9vh8]B9H3BEecwL%Vzh^VwKR3C2CNZ3Bv%=fFxm$z)duY[U+2=3CydZrkQFnA")
		err := c.(*client).PATCH("/foo", &response, []byte(`{"bar":"foo"}`))
		assert.NoError(t, err)
		assert.Equal(t, "foo", response.Bar)
	})

	t.Run("failure with with expected API error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPatch, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				res := httpmock.NewStringResponse(http.StatusBadRequest, `{
  "error": {
    "id": "400",
    "name": "error_name",
    "detail": "Error detail"
  }
}`)
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).PATCH("/foo", &response, []byte(`{"bar":"foo"}`))
		expectedErrStr := "api: error id=400 name=error_name detail=Error detail"
		assert.EqualError(t, err, expectedErrStr)
	})

	t.Run("failure with with unexpected API error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPatch, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(http.StatusInternalServerError, "Internal Server Error"), nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).PATCH("/foo", &response, []byte(`{"bar":"foo"}`))
		expectedErrStr := "api: error id=500 name=unknown_api_error detail=Unknown API error"
		assert.EqualError(t, err, expectedErrStr)
	})

	t.Run("silent failure due to invalid response model", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPatch, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				res := httpmock.NewStringResponse(http.StatusOK, `{"bar":"foo"}`)
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).PATCH("/foo", &response, []byte(`{"bar":"foo"}`))
		assert.NoError(t, err)
		assert.Equal(t, struct {
			Foo string `json:"foo"`
		}{}, response)
	})

	t.Run("regression test existence of request header content-type = application/json", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodPatch, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				res := httpmock.NewStringResponse(http.StatusOK, `{"bar":"foo"}`)
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).PATCH("/foo", &response, []byte(`{"bar":"foo"}`))
		assert.NoError(t, err)
		assert.Equal(t, struct {
			Foo string `json:"foo"`
		}{}, response)
	})
}

func TestClient_DELETE(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodDelete, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "application/json", req.Header.Get("Accept"))
				assert.Equal(t, "Bearer 6zL9vh8]B9H3BEecwL%Vzh^VwKR3C2CNZ3Bv%=fFxm$z)duY[U+2=3CydZrkQFnA", req.Header.Get("Authorization"))

				res := httpmock.NewStringResponse(http.StatusOK, `{
  "data": {
    "transaction": {
      "id": "some_id"
	}
  }
}`)
				res.Header.Add("X-Rate-Limit", "36/200")
				return res, nil
			},
		)

		response := struct {
			Data struct {
				Transaction struct {
					ID string `json:"id"`
				} `json:"transaction"`
			} `json:"data"`
		}{}

		c := NewClient("6zL9vh8]B9H3BEecwL%Vzh^VwKR3C2CNZ3Bv%=fFxm$z)duY[U+2=3CydZrkQFnA")
		err := c.(*client).DELETE("/foo", &response)
		assert.NoError(t, err)
		assert.Equal(t, "some_id", response.Data.Transaction.ID)
	})

	t.Run("failure with with expected API error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodDelete, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				res := httpmock.NewStringResponse(http.StatusBadRequest, `{
	  "error": {
		"id": "400",
		"name": "error_name",
		"detail": "Error detail"
	  }
	}`)
				res.Header.Add("X-Rate-Limit", "36/200")
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).DELETE("/foo", &response)
		expectedErrStr := "api: error id=400 name=error_name detail=Error detail"
		assert.EqualError(t, err, expectedErrStr)
	})

	t.Run("failure with with unexpected API error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodDelete, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				return httpmock.NewStringResponse(http.StatusInternalServerError, "Internal Server Error"), nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).DELETE("/foo", &response)
		expectedErrStr := "api: error id=500 name=unknown_api_error detail=Unknown API error"
		assert.EqualError(t, err, expectedErrStr)
	})

	t.Run("silent failure due to invalid response model", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(http.MethodDelete, fmt.Sprintf("%s%s", apiEndpoint, "/foo"),
			func(req *http.Request) (*http.Response, error) {
				res := httpmock.NewStringResponse(http.StatusOK, `{"bar":"foo"}`)
				res.Header.Add("X-Rate-Limit", "36/200")
				return res, nil
			},
		)

		response := struct {
			Foo string `json:"foo"`
		}{}

		c := NewClient("")
		err := c.(*client).DELETE("/foo", &response)
		assert.NoError(t, err)
		assert.Equal(t, struct {
			Foo string `json:"foo"`
		}{}, response)
	})
}

func TestClient_RateLimitingMethods(t *testing.T) {
	c := NewClient("test-token")

	// Test that rate limiting methods are available
	assert.Equal(t, 200, c.RequestsRemaining())           // Should start with full quota
	assert.Equal(t, 0, c.RequestsInWindow())              // No requests made yet
	assert.False(t, c.IsAtLimit())                        // Not at limit
	assert.Equal(t, time.Duration(0), c.TimeUntilReset()) // No requests to reset
}

func TestClient_AutomaticRateTracking(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodGet, fmt.Sprintf("%s%s", apiEndpoint, "/test"),
		func(req *http.Request) (*http.Response, error) {
			res := httpmock.NewStringResponse(http.StatusOK, `{"success": true}`)
			return res, nil
		},
	)

	c := NewClient("test-token")

	// Check initial state
	assert.Equal(t, 200, c.RequestsRemaining())
	assert.Equal(t, 0, c.RequestsInWindow())

	// Make a request
	response := struct {
		Success bool `json:"success"`
	}{}
	err := c.(*client).GET("/test", &response)
	assert.NoError(t, err)

	// Verify rate limiting was tracked automatically
	assert.Equal(t, 199, c.RequestsRemaining()) // Should decrease
	assert.Equal(t, 1, c.RequestsInWindow())    // Should increase
	assert.False(t, c.IsAtLimit())              // Still not at limit

	// Make another request
	err = c.(*client).GET("/test", &response)
	assert.NoError(t, err)

	// Verify tracking continues
	assert.Equal(t, 198, c.RequestsRemaining())
	assert.Equal(t, 2, c.RequestsInWindow())
}

func TestClient_RateLimitingNotTrackedOnError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodGet, fmt.Sprintf("%s%s", apiEndpoint, "/test"),
		func(req *http.Request) (*http.Response, error) {
			res := httpmock.NewStringResponse(http.StatusBadRequest, `{
				"error": {
					"id": "400",
					"name": "bad_request",
					"detail": "Bad request"
				}
			}`)
			return res, nil
		},
	)

	c := NewClient("test-token")

	// Check initial state
	assert.Equal(t, 200, c.RequestsRemaining())
	assert.Equal(t, 0, c.RequestsInWindow())

	// Make a failing request
	response := struct {
		Success bool `json:"success"`
	}{}
	err := c.(*client).GET("/test", &response)
	assert.Error(t, err)

	// Verify rate limiting was NOT tracked for failed request
	assert.Equal(t, 200, c.RequestsRemaining()) // Should remain unchanged
	assert.Equal(t, 0, c.RequestsInWindow())    // Should remain unchanged
}
