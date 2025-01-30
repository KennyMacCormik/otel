package otel_config

import (
	"github.com/KennyMacCormik/common/log"
	"github.com/KennyMacCormik/common/val"
	"github.com/KennyMacCormik/otel/otel-common/conf"
	"github.com/spf13/viper"
	"time"
)

type otelConfig struct {
	OTelEndpoint string        `mapstructure:"otel_endpoint" validate:"required,urlprefix,url"`
	StopTimeout  time.Duration `mapstructure:"otel_shutdown_timeout" validate:"min=100ms,max=30s"`
}

func NewOTelConfig() conf.OTelConfig {
	c := &otelConfig{}

	err := viper.BindEnv("otel_endpoint")
	if err != nil {
		log.Error("Failed to bind otel_endpoint")
	}

	viper.SetDefault("otel_shutdown_timeout", "500ms")
	err = viper.BindEnv("otel_shutdown_timeout")
	if err != nil {
		log.Error("Failed to bind otel_shutdown_timeout")
	}

	err = viper.Unmarshal(c)
	if err != nil {
		log.Error("Failed to unmarshal otelConfig")
	}

	err = val.ValidateStruct(c)
	if err != nil {
		log.Error("Failed to validate otelConfig", "err", err)
	}

	if err != nil {
		return nil
	}

	return c
}

func (o *otelConfig) Endpoint() string {
	return o.OTelEndpoint
}

func (o *otelConfig) ShutdownTimeout() time.Duration {
	return o.StopTimeout
}
