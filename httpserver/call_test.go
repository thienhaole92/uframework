//nolint:forcetypeassert,lll
package httpserver_test

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/thienhaole92/uframework/httpserver"
	"github.com/thienhaole92/uframework/notifylog"
)

type MockDelegate[REQ any] struct {
	mock.Mock
}

func (m *MockDelegate[REQ]) Invoke(log notifylog.NotifyLog, e echo.Context, req *REQ) (*httpserver.Response, *echo.HTTPError) {
	args := m.Called(log, e, req)
	if args.Get(1) != nil {
		return args.Get(0).(*httpserver.Response), args.Get(1).(*echo.HTTPError)
	}

	return args.Get(0).(*httpserver.Response), nil
}

func TestCall_Success(t *testing.T) {
	t.Parallel()

	e := echo.New()
	ctx := e.NewContext(nil, nil)
	ctx.Set(httpserver.RequestIDContextKey, "12345")

	var req struct{}

	res := &httpserver.Response{
		RequestID: "12345",
		Data:      nil,
		Pagination: &httpserver.Pagination{
			Limit:     0,
			Total:     0,
			TotalPage: 0,
		},
	}

	mockDelegate := new(MockDelegate[struct{}])
	mockDelegate.On("Invoke", notifylog.New("test_log", notifylog.JSON), ctx, &req).Return(res, nil)

	res, err := httpserver.Call(ctx, &req, "test_log", mockDelegate.Invoke)

	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, "12345", res.RequestID)
	mockDelegate.AssertExpectations(t)
}

func TestCall_DelegateReturnsError(t *testing.T) {
	t.Parallel()

	e := echo.New()
	ctx := e.NewContext(nil, nil)

	var req struct{}

	httpErr := &echo.HTTPError{Code: 500, Message: "Internal Server Error", Internal: nil}
	mockDelegate := new(MockDelegate[struct{}])
	mockDelegate.On("Invoke", notifylog.New("test_log", notifylog.JSON), ctx, &req).Return((*httpserver.Response)(nil), httpErr)

	res, err := httpserver.Call(ctx, &req, "test_log", mockDelegate.Invoke)

	require.Nil(t, res)
	require.Equal(t, httpErr, err)
	mockDelegate.AssertExpectations(t)
}

func TestCall_NoRequestIDInContext(t *testing.T) {
	t.Parallel()

	e := echo.New()
	ctx := e.NewContext(nil, nil)

	var req struct{}

	res := &httpserver.Response{
		RequestID: "",
		Data:      nil,
		Pagination: &httpserver.Pagination{
			Limit:     0,
			Total:     0,
			TotalPage: 0,
		},
	}

	mockDelegate := new(MockDelegate[struct{}])
	mockDelegate.On("Invoke", notifylog.New("test_log", notifylog.JSON), ctx, &req).Return(res, nil)

	res, err := httpserver.Call(ctx, &req, "test_log", mockDelegate.Invoke)

	require.Nil(t, err)
	require.NotNil(t, res)
	require.Empty(t, res.RequestID)
	mockDelegate.AssertExpectations(t)
}
