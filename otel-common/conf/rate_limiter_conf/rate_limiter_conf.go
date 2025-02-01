package rate_limiter_conf

import (
	"github.com/KennyMacCormik/common/log"
	"github.com/KennyMacCormik/common/val"
	"github.com/KennyMacCormik/otel/otel-common/conf"
	"github.com/spf13/viper"
)

type rateLimiterConfig struct {
	MaxRun  int64 `mapstructure:"rate_limiter_max_conn" validate:"min=1,max=100000"`
	MaxWait int64 `mapstructure:"rate_limiter_max_wait" validate:"min=1,max=100000"`
	Retry   int64 `mapstructure:"rate_limiter_retry_after" validate:"min=1,max=60"`
}

func NewRateLimiterConfig() conf.RateLimiterConf {
	c := &rateLimiterConfig{}

	viper.SetDefault("rate_limiter_max_conn", "100")
	err := viper.BindEnv("rate_limiter_max_conn")
	if err != nil {
		log.Error("Failed to bind rate_limiter_max_conn")
	}

	viper.SetDefault("rate_limiter_max_wait", "100")
	err = viper.BindEnv("rate_limiter_max_wait")
	if err != nil {
		log.Error("Failed to bind rate_limiter_max_wait")
	}

	viper.SetDefault("rate_limiter_retry_after", "1")
	err = viper.BindEnv("rate_limiter_retry_after")
	if err != nil {
		log.Error("Failed to bind rate_limiter_retry_after")
	}

	err = viper.Unmarshal(c)
	if err != nil {
		log.Error("Failed to unmarshal rateLimiterConfig")
	}

	err = val.ValidateStruct(c)
	if err != nil {
		log.Error("Failed to validate rateLimiterConfig", "err", err)
	}

	if err != nil {
		return nil
	}

	return c
}

func (r *rateLimiterConfig) MaxRunning() int64 {
	return r.MaxRun
}

func (r *rateLimiterConfig) MaxWaiting() int64 {
	return r.MaxWait
}

func (r *rateLimiterConfig) RetryAfter() int64 {
	return r.Retry
}
