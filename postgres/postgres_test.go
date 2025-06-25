package postgres_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/tracelog"
	"github.com/stretchr/testify/require"
	"github.com/thienhaole92/uframework/postgres"
	"github.com/thienhaole92/uframework/testutil"
)

func TestPostgresConnection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Start the test container
	container := testutil.SetupPostgresContainer(ctx, t)

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		container.User,
		container.Password,
		net.JoinHostPort(container.Host, container.Port.Port()),
		container.Database,
	)

	// Define options
	opts := &postgres.Option{
		URL:                   dbURL,
		MaxConnection:         5,
		MinConnection:         1,
		MaxConnectionIdleTime: 60 * time.Second,
		PingTimeout:           10 * time.Second,
		LogLevel:              tracelog.LogLevelTrace,
	}

	// Create Postgres instance
	postgresDB := postgres.New(ctx, opts)
	require.NotNil(t, postgresDB)

	// Check if connection is alive
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := postgresDB.Ping(ctx)
	require.NoError(t, err)

	// Close the DB pool
	postgresDB.Close()
}
