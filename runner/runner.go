package runner

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/thienhaole92/uframework/container"
	"github.com/thienhaole92/uframework/goredis"
	"github.com/thienhaole92/uframework/httpserver"
	"github.com/thienhaole92/uframework/postgres"
)

type Server interface {
	Run()
	Close()
}

type AppRunner interface {
	Server
	Name() string
}

type Runner struct {
	container *container.Container
	servers   []Server
	runners   []AppRunner
}

type Option func(*Runner)

func New(opts ...Option) *Runner {
	rnn := &Runner{
		container: container.New(),
		servers:   []Server{},
		runners:   []AppRunner{},
	}

	for _, opt := range opts {
		opt(rnn)
	}

	return rnn
}

func (r *Runner) Run() {
	for _, svr := range r.servers {
		go svr.Run()
	}

	for _, rnn := range r.runners {
		go rnn.Run()
	}

	r.handleShutdown()
}

func (r *Runner) handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Info().Msgf("received signal: %s, shutting down gracefully...", sig)

	// Close all servers
	for _, svr := range r.servers {
		svr.Close()
	}

	// Close all runners
	for _, rnn := range r.runners {
		rnn.Close()
	}

	log.Info().Msg("all servers stopped, exiting...")
}

func WithServer(svr Server, name ...string) Option {
	return func(r *Runner) {
		r.servers = append(r.servers, svr)

		if name != nil {
			log.Info().Msgf("%s server registered", name)
		} else {
			log.Info().Msgf("server registered")
		}
	}
}

func WithAppRunner(svr AppRunner, name string) Option {
	return func(r *Runner) {
		r.runners = append(r.runners, svr)

		log.Info().Msgf("%s app runner registered", name)
	}
}

func WithHTTPServer(hook func(*container.Container) *httpserver.Server) Option {
	return func(r *Runner) {
		svr := hook(r.container)
		r.servers = append(r.servers, svr)
		r.container.SetEchoGroup(svr.Root)

		log.Info().Msg("http server registered")
	}
}

func WithRedis(redis *goredis.Redis) Option {
	return func(r *Runner) {
		r.container.SetRedis(redis)

		log.Info().Msg("redis registered")
	}
}

func WithPostgres(p *postgres.Postgres) Option {
	return func(r *Runner) {
		r.container.SetPostgres(p)

		log.Info().Msg("postgres registered")
	}
}

func WithRestAPIService(hook func(*container.Container)) Option {
	return func(r *Runner) {
		hook(r.container)

		log.Info().Msg("rest api service registered")
	}
}

func WithConsumers(hook func(*container.Container)) Option {
	return func(r *Runner) {
		hook(r.container)

		log.Info().Msg("consumers registered")
	}
}
