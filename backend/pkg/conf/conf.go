package conf

import "time"

type LoggerConf interface {
	Format() string
	Level() string
}

type HttpConf interface {
	Endpoint() string
	ReadTimeout() time.Duration
	WriteTimeout() time.Duration
	IdleTimeout() time.Duration
	ShutdownTimeout() time.Duration
}

type RateLimiterConf interface {
	MaxRunning() int64
	MaxWaiting() int64
	RetryAfter() int64
}

type OTelConfig interface {
	Endpoint() string
	ShutdownTimeout() time.Duration
}

type GinConfig interface {
	Mode() string
}
