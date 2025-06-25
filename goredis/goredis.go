package goredis

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Option struct {
	Host         string
	Port         int
	Password     string
	DB           int
	TTL          time.Duration
	DialTimeout  time.Duration
	UseTLS       bool
	MaxIdleConns int
	MinIdleConns int
	PingTimeout  time.Duration
}

type Redis struct {
	*redis.Client
}

func New(opts *Option) *Redis {
	opt := buildRedisOptions(opts)

	client := redis.NewClient(&opt)

	ctx, cancel := context.WithTimeout(context.Background(), opts.PingTimeout)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		panic(err)
	}

	return &Redis{Client: client}
}

func buildRedisOptions(opts *Option) redis.Options {
	opt := redis.Options{
		Addr:                       net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port)),
		Password:                   opts.Password,
		DB:                         opts.DB,
		DialTimeout:                opts.DialTimeout,
		MaxIdleConns:               opts.MaxIdleConns,
		MinIdleConns:               opts.MinIdleConns,
		Network:                    "",
		ClientName:                 "",
		Dialer:                     nil,
		OnConnect:                  nil,
		Protocol:                   0,
		Username:                   "",
		CredentialsProvider:        nil,
		CredentialsProviderContext: nil,
		MaxRetries:                 0,
		MinRetryBackoff:            0,
		MaxRetryBackoff:            0,
		ReadTimeout:                0,
		WriteTimeout:               0,
		ContextTimeoutEnabled:      false,
		PoolFIFO:                   false,
		PoolSize:                   0,
		PoolTimeout:                0,
		MaxActiveConns:             0,
		ConnMaxIdleTime:            0,
		ConnMaxLifetime:            0,
		TLSConfig:                  nil,
		Limiter:                    nil,
		DisableIndentity:           false,
		IdentitySuffix:             "",
		UnstableResp3:              false,
	}

	if opts.UseTLS {
		opt.TLSConfig = createTLSConfig()
	}

	return opt
}

func createTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion:                          tls.VersionTLS12,
		Rand:                                nil,
		Time:                                nil,
		Certificates:                        nil,
		NameToCertificate:                   nil,
		GetCertificate:                      nil,
		GetClientCertificate:                nil,
		GetConfigForClient:                  nil,
		VerifyPeerCertificate:               nil,
		VerifyConnection:                    nil,
		RootCAs:                             nil,
		NextProtos:                          nil,
		ServerName:                          "",
		ClientAuth:                          0,
		ClientCAs:                           nil,
		InsecureSkipVerify:                  false,
		CipherSuites:                        nil,
		PreferServerCipherSuites:            true,
		SessionTicketsDisabled:              false,
		SessionTicketKey:                    [32]byte{},
		ClientSessionCache:                  nil,
		UnwrapSession:                       nil,
		WrapSession:                         nil,
		MaxVersion:                          0,
		CurvePreferences:                    nil,
		DynamicRecordSizingDisabled:         false,
		Renegotiation:                       0,
		KeyLogWriter:                        nil,
		EncryptedClientHelloConfigList:      nil,
		EncryptedClientHelloRejectionVerify: nil,
	}
}
