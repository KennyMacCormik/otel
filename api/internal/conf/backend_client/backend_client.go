package backend_client

import (
	"time"

	"github.com/KennyMacCormik/common/log"
	"github.com/KennyMacCormik/common/val"
	"github.com/spf13/viper"

	"github.com/KennyMacCormik/otel/api/internal/conf"
)

type backendClientConf struct {
	ClientEndpoint       string        `mapstructure:"backend_client_endpoint" validate:"url,required"`
	ClientRequestTimeout time.Duration `mapstructure:"backend_client_request_timeout" validate:"min=100ms,max=1s"`
}

func NewBackendClientConf() conf.BackendClientConf {
	c := &backendClientConf{}

	err := viper.BindEnv("backend_client_endpoint")
	if err != nil {
		log.Error("Failed to bind backend_client_endpoint")
	}

	viper.SetDefault("backend_client_request_timeout", "200ms")
	err = viper.BindEnv("backend_client_request_timeout")
	if err != nil {
		log.Error("Failed to bind backend_client_request_timeout")
	}

	err = viper.Unmarshal(c)
	if err != nil {
		log.Error("Failed to unmarshal backendClientConf")
	}

	err = val.ValidateStruct(c)
	if err != nil {
		log.Error("Failed to validate backendClientConf", "err", err)
	}

	if err != nil {
		return nil
	}

	return c
}

func (l *backendClientConf) Endpoint() string {
	return l.ClientEndpoint
}

func (l *backendClientConf) RequestTimeout() time.Duration {
	return l.ClientRequestTimeout
}
