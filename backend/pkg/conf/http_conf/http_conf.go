package http_conf

import (
	"strconv"
	"strings"
	"time"

	"github.com/KennyMacCormik/common/log"
	"github.com/KennyMacCormik/common/val"
	"github.com/spf13/viper"

	"github.com/KennyMacCormik/otel/backend/pkg/conf"
)

type httpConf struct {
	HttpHost            string        `mapstructure:"http_host" validate:"ip4_addr|fqdn,required"`
	HttpPort            int           `mapstructure:"http_port" validate:"numeric,gt=1024,lt=65536,required"`
	HttpReadTimeout     time.Duration `mapstructure:"http_read_timeout" validate:"min=100ms,max=1s"`
	HttpWriteTimeout    time.Duration `mapstructure:"http_write_timeout" validate:"min=100ms,max=1s"`
	HttpIdleTimeout     time.Duration `mapstructure:"http_idle_timeout" validate:"min=100ms,max=1s"`
	HttpShutdownTimeout time.Duration `mapstructure:"http_shutdown_timeout" validate:"min=100ms,max=30s"`
}

func NewHTTPConf() conf.HttpConf {
	c := &httpConf{}

	err := viper.BindEnv("http_host")
	if err != nil {
		log.Error("Failed to bind http_host")
	}

	err = viper.BindEnv("http_port")
	if err != nil {
		log.Error("Failed to bind http_port")
	}

	viper.SetDefault("http_read_timeout", "100ms")
	err = viper.BindEnv("http_read_timeout")
	if err != nil {
		log.Error("Failed to bind http_read_timeout")
	}

	viper.SetDefault("http_write_timeout", "100ms")
	err = viper.BindEnv("http_write_timeout")
	if err != nil {
		log.Error("Failed to bind http_write_timeout")
	}

	viper.SetDefault("http_idle_timeout", "100ms")
	err = viper.BindEnv("http_idle_timeout")
	if err != nil {
		log.Error("Failed to bind http_idle_timeout")
	}

	viper.SetDefault("http_shutdown_timeout", "10s")
	err = viper.BindEnv("http_shutdown_timeout")
	if err != nil {
		log.Error("Failed to bind http_shutdown_timeout")
	}

	err = viper.Unmarshal(c)
	if err != nil {
		log.Error("Failed to unmarshal httpConf")
	}

	err = val.ValidateStruct(c)
	if err != nil {
		log.Error("Failed to validate httpConf", "err", err)
	}

	if err != nil {
		return nil
	}

	return c
}

func (h httpConf) Endpoint() string {
	return strings.Join([]string{h.HttpHost, strconv.Itoa(h.HttpPort)}, ":")
}

func (h httpConf) ReadTimeout() time.Duration {
	return h.HttpReadTimeout
}

func (h httpConf) WriteTimeout() time.Duration {
	return h.HttpWriteTimeout
}

func (h httpConf) IdleTimeout() time.Duration {
	return h.HttpIdleTimeout
}

func (h httpConf) ShutdownTimeout() time.Duration {
	return h.HttpShutdownTimeout
}
