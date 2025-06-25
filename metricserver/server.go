package metricserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type Option struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	GracePeriod  time.Duration
	MetricPath   string
	StatusPath   string
}

type Server struct {
	gracePeriod time.Duration
	Echo        *echo.Echo
	Server      *http.Server
}

func New(opts *Option) *Server {
	ech := echo.New()

	ech.HideBanner = true
	ech.GET(opts.StatusPath, func(ctx echo.Context) error {
		return ctx.JSON(http.StatusOK, map[string]any{"status": "ok"})
	})
	ech.GET(opts.MetricPath, echoprometheus.NewHandler())

	server := &http.Server{
		Addr:                         net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port)),
		Handler:                      ech,
		ReadTimeout:                  opts.ReadTimeout,
		WriteTimeout:                 opts.WriteTimeout,
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

	log.Info().Msg("shutdown server")
}
