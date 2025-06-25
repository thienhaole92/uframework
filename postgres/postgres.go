package postgres

import (
	"context"
	"os"
	"time"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	pgxzerolog "github.com/jackc/pgx-zerolog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Option struct {
	URL                   string
	MaxConnection         int32
	MinConnection         int32
	MaxConnectionIdleTime time.Duration
	PingTimeout           time.Duration
	LogLevel              tracelog.LogLevel
}

type Postgres struct {
	PgxIface
}

func New(ctx context.Context, opts *Option) *Postgres {
	pgConfig, err := pgxpool.ParseConfig(opts.URL)
	if err != nil {
		log.Panic().Err(err).Msg("can not parse config")
	}

	zlog := zerolog.New(os.Stderr).With().Caller().Stack().Str("logger", "postgres").Timestamp().Logger()
	logger := pgxzerolog.NewLogger(zlog, pgxzerolog.WithoutPGXModule())

	tracer := &tracelog.TraceLog{
		Logger:   logger,
		LogLevel: opts.LogLevel,
		Config:   nil,
	}

	pgConfig.MaxConns = opts.MaxConnection
	pgConfig.MinConns = opts.MinConnection
	pgConfig.MaxConnIdleTime = opts.MaxConnectionIdleTime
	pgConfig.ConnConfig.Tracer = tracer
	pgConfig.AfterConnect = func(_ context.Context, conn *pgx.Conn) error {
		pgxdecimal.Register(conn.TypeMap())

		return nil
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, opts.PingTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctxWithTimeout, pgConfig)
	if err != nil {
		log.Panic().Err(err).Msg("can not create pool")
	}

	if err := pool.Ping(ctxWithTimeout); err != nil {
		log.Panic().Err(err).Msg("can not ping pool")
	}

	return &Postgres{PgxIface: pool}
}
