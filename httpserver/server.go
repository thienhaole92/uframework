package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/thienhaole92/uframework/middleware"
	"github.com/thienhaole92/uframework/validator"
)

type Option struct {
	Host             string
	Port             int
	EnableCors       bool
	BodyLimit        string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	GracePeriod      time.Duration
	Subsystem        string
	RequireRequestID bool
}

type Server struct {
	gracePeriod time.Duration
	Echo        *echo.Echo
	Server      *http.Server
	Root        *echo.Group
}

func New(opts *Option) *Server {
	ech := echo.New()

	ech.HideBanner = true
	ech.Validator = validator.DefaultRestValidator()
	ech.HTTPErrorHandler = middleware.ErrorHandler(ech.DefaultHTTPErrorHandler)

	ech.Pre(middleware.RequestID(requestIDSkipper(opts.RequireRequestID)))
	ech.Pre(echoprometheus.NewMiddleware(opts.Subsystem))
	ech.Pre(middleware.RequestLogger(log.Logger, RestLogFieldsExtractor))
	ech.Pre(echomiddleware.BodyLimit(opts.BodyLimit))

	if opts.EnableCors {
		ech.Use(echomiddleware.CORS())
	}

	root := ech.Group("")

	server := &http.Server{
		Handler:                      ech,
		ReadTimeout:                  opts.ReadTimeout,
		WriteTimeout:                 opts.WriteTimeout,
		Addr:                         net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port)),
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadHeaderTimeout:            0,
		IdleTimeout:                  0,
		MaxHeaderBytes:               0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext:                  nil,
	}

	return &Server{
		gracePeriod: opts.GracePeriod,
		Echo:        ech,
		Server:      server,
		Root:        root,
	}
}

func (s *Server) Run() {
	go func() {
		log.Info().Msg("start server")

		if err := s.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Panic().Err(err).Msg("failed to start server")
		}
	}()
}

func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.TODO(), s.gracePeriod)
	defer cancel()

	if err := s.Server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("could not gracefully shut down web server")
	}
}

func RestLogFieldsExtractor(ectx echo.Context) map[string]any {
	if req := ectx.Get(RequestObjectKey); req != nil {
		var reqObject string

		if b, err := json.Marshal(req); err != nil {
			reqObject = fmt.Sprintf("failed to parse request object as string: %+v", err)
		} else {
			reqObject = string(b)
		}

		return map[string]any{"request_object": reqObject}
	}

	return nil
}

func requestIDSkipper(skip bool) echomiddleware.Skipper {
	return func(_ echo.Context) bool {
		return skip
	}
}
