package httpserver_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thienhaole92/uframework/httpserver"
)

//nolint:ireturn
func SetupTestServer(
	method string,
	path string,
	headers map[string]string,
	body []byte,
	opts httpserver.Option,
) (echo.Context, *http.Request, *httptest.ResponseRecorder, *httpserver.Server) {
	server := httpserver.New(&opts)

	req := httptest.NewRequest(method, path, bytes.NewBuffer(body))

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rec := httptest.NewRecorder()

	ctx := server.Echo.NewContext(req, rec)

	return ctx, req, rec, server
}

func TestNewServer(t *testing.T) {
	t.Parallel()

	// Test options
	opts := httpserver.Option{
		Host:             "0.0.0.0",
		Port:             80808,
		EnableCors:       true,
		BodyLimit:        "1M",
		ReadTimeout:      time.Second * 10,
		WriteTimeout:     time.Second * 10,
		GracePeriod:      time.Second * 10,
		Subsystem:        "echo",
		RequireRequestID: true,
	}

	headers := map[string]string{
		"Origin":                        "http://example.com",
		"Access-Control-Request-Method": "POST",
		echo.HeaderXRequestID:           uuid.NewString(),
	}
	_, req, rec, server := SetupTestServer(http.MethodPost, "/", headers, nil, opts)

	// Create a route that responds to preflight requests
	server.Echo.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")

			return next(c)
		}
	})

	server.Echo.ServeHTTP(rec, req)

	// Assert CORS header was set
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestServerWithMiddleware_Success(t *testing.T) {
	t.Parallel()

	// Create options for the server
	opts := httpserver.Option{
		Host:             "0.0.0.0",
		Port:             80809,
		EnableCors:       false,
		BodyLimit:        "2M", // 2MB body limit
		ReadTimeout:      time.Second * 15,
		WriteTimeout:     time.Second * 15,
		GracePeriod:      time.Second * 10,
		Subsystem:        "success",
		RequireRequestID: true,
	}

	headers := map[string]string{
		"Content-Type":        "application/json",
		echo.HeaderXRequestID: uuid.NewString(),
	}

	// Setup the server and test context
	_, req, rec, server := SetupTestServer(http.MethodPost, "/test", headers, []byte("test body"), opts)

	// Define a simple handler to test body limit middleware
	server.Echo.POST("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Call the handler with a valid body
	server.Echo.ServeHTTP(rec, req)

	// Test if the request passes through with the defined body limit middleware
	assert.Equal(t, http.StatusOK, rec.Code, "Expected status 200 OK")
}

func TestServerWithMiddleware_Failure(t *testing.T) {
	t.Parallel()

	// Create options for the server
	opts := httpserver.Option{
		Host:             "0.0.0.0",
		Port:             80807,
		EnableCors:       false,
		BodyLimit:        "2M", // 2MB body limit
		ReadTimeout:      time.Second * 15,
		WriteTimeout:     time.Second * 15,
		GracePeriod:      time.Second * 10,
		Subsystem:        "failure",
		RequireRequestID: true,
	}

	headers := map[string]string{
		"Content-Type":        "application/json",
		echo.HeaderXRequestID: uuid.NewString(),
	}

	// Setup the server and test context
	_, req, rec, server := SetupTestServer(http.MethodPost, "/test", headers, make([]byte, 3*1024*1024), opts)

	// Define a simple handler to test body limit middleware
	server.Echo.POST("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Call the handler and check for failure
	server.Echo.ServeHTTP(rec, req)

	// Assert 413 Payload Too Large when the body size exceeds the limit
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code, "Expected status 413 Payload Too Large")
}

func TestRestLogFieldsExtractor(t *testing.T) {
	t.Parallel()

	// Create a mock request
	payload := map[string]interface{}{
		"username": "testuser",
		"password": "password123",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	// Create an echo context
	e := echo.New()
	ctx := e.NewContext(req, rec)

	// Set the request object in context
	ctx.Set(httpserver.RequestObjectKey, payload)

	// Extract log fields using RestLogFieldsExtractor
	logFields := httpserver.RestLogFieldsExtractor(ctx)

	// Verify the extracted log fields
	assert.NotNil(t, logFields)
	assert.Contains(t, logFields, "request_object")

	requestObject, ok := logFields["request_object"].(string)
	if !ok {
		t.Fatal("Expected 'request_object' to be of type string")
	}

	assert.JSONEq(t, `{"username":"testuser","password":"password123"}`, requestObject)
}
