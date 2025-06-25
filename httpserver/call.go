package httpserver

import (
	"github.com/labstack/echo/v4"
	"github.com/thienhaole92/uframework/notifylog"
)

type Delegate[REQ any] func(notifylog.NotifyLog, echo.Context, *REQ) (*Response, *echo.HTTPError)

func Call[REQ any](e echo.Context, request *REQ, name string, delegate Delegate[REQ]) (*Response, *echo.HTTPError) {
	log := notifylog.New(name, notifylog.JSON)

	res, err := delegate(log, e, request)
	if res != nil {
		requestID, ok := e.Get(RequestIDContextKey).(string)
		if ok && len(res.RequestID) == 0 {
			res.RequestID = requestID
		}
	}

	return res, err
}
