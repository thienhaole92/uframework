package container

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/thienhaole92/uframework/goredis"
	"github.com/thienhaole92/uframework/postgres"
)

var (
	ErrRedisNotSet      = errors.New("redis is not initialized, please set it up before use")
	ErrPostgresNotSet   = errors.New("postgres is not initialized, please set it up before use")
	ErrEchoServerNotSet = errors.New("echo server is not initialized, please set it up before use")
	ErrEchoGroupNotSet  = errors.New("echo group is not initialized, please set it up before use")
)

type Container struct {
	postgres   *postgres.Postgres
	redis      *goredis.Redis
	echoGroup  *echo.Group
	echoServer *echo.Echo
}

func New() *Container {
	return &Container{
		postgres:   nil,
		redis:      nil,
		echoGroup:  nil,
		echoServer: nil,
	}
}

func (c *Container) SetPostgres(p *postgres.Postgres) {
	c.postgres = p
}

func (c *Container) Postgres() (*postgres.Postgres, error) {
	if c.postgres == nil {
		return nil, ErrPostgresNotSet
	}

	return c.postgres, nil
}

func (c *Container) MustPostgres() *postgres.Postgres {
	if c.postgres == nil {
		panic(ErrPostgresNotSet)
	}

	return c.postgres
}

func (c *Container) SetRedis(r *goredis.Redis) {
	c.redis = r
}

func (c *Container) Redis() (*goredis.Redis, error) {
	if c.redis == nil {
		return nil, ErrRedisNotSet
	}

	return c.redis, nil
}

func (c *Container) MustRedis() *goredis.Redis {
	if c.redis == nil {
		panic(ErrRedisNotSet)
	}

	return c.redis
}

func (c *Container) SetEchoGroup(g *echo.Group) {
	c.echoGroup = g
}

func (c *Container) SetEchoServer(e *echo.Echo) {
	c.echoServer = e
}

func (c *Container) EchoGroup() (*echo.Group, error) {
	if c.echoGroup == nil {
		return nil, ErrEchoGroupNotSet
	}

	return c.echoGroup, nil
}

func (c *Container) MustEchoGroup() *echo.Group {
	if c.echoGroup == nil {
		panic(ErrEchoGroupNotSet)
	}

	return c.echoGroup
}

func (c *Container) EchoServer() (*echo.Echo, error) {
	if c.echoServer == nil {
		return nil, ErrEchoServerNotSet
	}

	return c.echoServer, nil
}

func (c *Container) MustEchoServer() *echo.Echo {
	if c.echoServer == nil {
		panic(ErrEchoServerNotSet)
	}

	return c.echoServer
}
