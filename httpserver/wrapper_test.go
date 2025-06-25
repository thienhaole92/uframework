package httpserver_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/thienhaole92/uframework/httpserver"
	"github.com/thienhaole92/uframework/testutil"
)

type MockRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func TestWrapper_Success(t *testing.T) {
	t.Parallel()

	// Prepare valid request payload
	payload := MockRequest{Username: "testuser", Password: "securepassword"}
	jsonPayload, err := json.Marshal(payload)
	require.NoError(t, err, "Failed to marshal JSON payload")

	// Set up test context
	ctx, rec, _ := testutil.SetupEchoContext(t, &testutil.Options{
		Method: http.MethodPost,
		Path:   "/test",
		Body:   jsonPayload,
	})

	// Mock handler function that returns a sample response
	mockHandler := func(_ echo.Context, _ *MockRequest) (any, *echo.HTTPError) {
		return &httpserver.Response{
			RequestID:  "123",
			Data:       "Mock Response Data",
			Pagination: nil,
		}, nil
	}

	// Wrap the mock handler with the Wrapper function
	wrappedHandler := httpserver.Wrapper(mockHandler)

	// Execute the wrapped handler
	err = wrappedHandler(ctx)
	require.NoError(t, err, "Handler execution failed")

	// Verify response status
	require.Equal(t, http.StatusOK, rec.Code, "Unexpected response status")
}

func TestWrapper_Failure_Binding(t *testing.T) {
	t.Parallel()

	// Set up test context with invalid JSON
	ctx, _, _ := testutil.SetupEchoContext(t, &testutil.Options{
		Method: http.MethodPost,
		Path:   "/test",
		Body:   []byte("invalid json format"),
	})

	// Mock handler that should not be executed due to binding failure
	mockHandler := func(_ echo.Context, _ *MockRequest) (any, *echo.HTTPError) {
		t.Fatalf("Handler should not be called on binding failure")

		return nil, nil
	}

	// Wrap the mock handler with the Wrapper function
	wrappedHandler := httpserver.Wrapper(mockHandler)

	// Execute the wrapped handler
	err := wrappedHandler(ctx)

	// Assertions
	require.Error(t, err, "Expected error due to binding failure")

	var httpErr *echo.HTTPError

	require.ErrorAs(t, err, &httpErr, "Expected error of type *echo.HTTPError")
	require.Equal(t, http.StatusBadRequest, httpErr.Code, "Expected HTTP 400 Bad Request")
}

func TestWrapper_Failure_Validation(t *testing.T) {
	t.Parallel()

	// Prepare invalid request payload (missing required fields)
	invalidPayload := MockRequest{Username: "", Password: "Password"}
	jsonPayload, err := json.Marshal(invalidPayload)
	require.NoError(t, err, "Failed to marshal JSON payload")

	// Set up test context
	ctx, _, _ := testutil.SetupEchoContext(t, &testutil.Options{
		Method: http.MethodPost,
		Path:   "/test",
		Body:   jsonPayload,
	})

	// Mock handler that should not be executed due to validation failure
	mockHandler := func(_ echo.Context, _ *MockRequest) (any, *echo.HTTPError) {
		t.Fatalf("Handler should not be called on validation failure")

		return nil, nil
	}

	// Wrap the mock handler with the Wrapper function
	wrappedHandler := httpserver.Wrapper(mockHandler)

	// Execute the wrapped handler
	err = wrappedHandler(ctx)

	// Assertions
	require.Error(t, err, "Expected error due to validation failure")

	var httpErr *echo.HTTPError

	require.ErrorAs(t, err, &httpErr, "Expected error of type *echo.HTTPError")
	require.Equal(t, http.StatusBadRequest, httpErr.Code, "Expected HTTP 400 Bad Request")
}
