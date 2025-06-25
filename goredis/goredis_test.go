package goredis_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thienhaole92/uframework/goredis"
	"github.com/thienhaole92/uframework/testutil"
)

func TestRedisConnection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Start the test container
	container := testutil.SetupRedisContainer(ctx, t)

	port, err := strconv.Atoi(container.Port.Port())
	require.NoError(t, err)

	opts := &goredis.Option{
		Host:         container.Host,
		Port:         port,
		Password:     "",
		DB:           0,
		DialTimeout:  5 * time.Second,
		UseTLS:       false,
		MaxIdleConns: 5,
		MinIdleConns: 1,
		PingTimeout:  2 * time.Second,
		TTL:          time.Minute,
	}

	redis := goredis.New(opts)

	_, err = redis.Ping(ctx).Result()
	require.NoError(t, err, "failed to ping Redis")
}
