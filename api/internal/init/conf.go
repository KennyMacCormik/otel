package init

import (
	"time"

	"github.com/KennyMacCormik/otel/backend/pkg/conf/gin_conf"
	"github.com/KennyMacCormik/otel/backend/pkg/conf/http_conf"
	"github.com/KennyMacCormik/otel/backend/pkg/conf/logger_conf"
	"github.com/KennyMacCormik/otel/backend/pkg/conf/otel_config"
	"github.com/KennyMacCormik/otel/backend/pkg/conf/rate_limiter_conf"

	"github.com/KennyMacCormik/otel/api/internal/conf/backend_client"
)

type Config struct {
	Log         Log
	OTel        OTel
	RateLimiter RateLimiter
	Http        Http
	Gin         Gin
	Client      Client
}
type Client struct {
	Endpoint       string
	RequestTimeout time.Duration
}
type Gin struct {
	Mode string
}
type Log struct {
	Format string
	Level  string
}
type OTel struct {
	Endpoint        string
	ShutdownTimeout time.Duration
}
type RateLimiter struct {
	MaxRunning int64
	MaxWait    int64
	RetryAfter int64
}
type Http struct {
	Endpoint        string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func GetConfig() *Config {
	cfg := &Config{}

	fns := []func() bool{
		cfg.getLoggingConfig,
		cfg.getHttpConfig,
		cfg.getOTelConfig,
		cfg.getRateLimiterConfig,
		cfg.getGinConfig,
		cfg.getBackendClientConfig,
	}

	for _, fn := range fns {
		if !fn() {
			return nil
		}
	}

	return cfg
}

func (c *Config) getGinConfig() bool {
	i := gin_conf.NewGinConf()
	if i == nil {
		return false
	}

	c.Gin.Mode = i.Mode()

	return true
}

func (c *Config) getLoggingConfig() bool {
	i := logger_conf.NewLoggerConf()
	if i == nil {
		return false
	}

	c.Log.Format = i.Format()
	c.Log.Level = i.Level()

	return true
}

func (c *Config) getHttpConfig() bool {
	i := http_conf.NewHTTPConf()
	if i == nil {
		return false
	}

	c.Http.Endpoint = i.Endpoint()
	c.Http.IdleTimeout = i.IdleTimeout()
	c.Http.ReadTimeout = i.ReadTimeout()
	c.Http.WriteTimeout = i.WriteTimeout()
	c.Http.ShutdownTimeout = i.ShutdownTimeout()

	return true
}

func (c *Config) getOTelConfig() bool {
	i := otel_config.NewOTelConfig()
	if i == nil {
		return false
	}

	c.OTel.Endpoint = i.Endpoint()
	c.OTel.ShutdownTimeout = i.ShutdownTimeout()

	return true
}

func (c *Config) getRateLimiterConfig() bool {
	i := rate_limiter_conf.NewRateLimiterConfig()
	if i == nil {
		return false
	}

	c.RateLimiter.RetryAfter = i.RetryAfter()
	c.RateLimiter.MaxRunning = i.MaxRunning()
	c.RateLimiter.MaxWait = i.MaxWaiting()

	return true
}

func (c *Config) getBackendClientConfig() bool {
	i := backend_client.NewBackendClientConf()
	if i == nil {
		return false
	}

	c.Client.Endpoint = i.Endpoint()
	c.Client.RequestTimeout = i.RequestTimeout()

	return true
}
