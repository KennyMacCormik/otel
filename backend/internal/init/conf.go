package init

import (
	"fmt"
	"github.com/KennyMacCormik/HerdMaster/pkg/cfg"
	"github.com/KennyMacCormik/HerdMaster/pkg/cfg/genCfg"
	"github.com/KennyMacCormik/HerdMaster/pkg/val"
	"reflect"
)

const (
	LoggingConfigName     = "log"
	HttpConfigName        = "http"
	RateLimiterConfigName = "rateLimiter"
	OtelConfigName        = "otel"
)

type Config struct {
	Log         genCfg.LoggingConfig
	Http        genCfg.HttpConfig
	RateLimiter genCfg.RateLimiterConfig
	Otel        genCfg.OtelConfig
}

func InitConfig(validator val.Validator) (*Config, error) {
	var conf Config
	if err := registerCfg(registerLogCfg, registerHttpCfg, registerRateLimiterCfg, registerOtelCfg); err != nil {
		return nil, err
	}
	if err := cfg.NewConfig(); err != nil {
		return nil, err
	}
	if err := bindCfgToConfig(&conf, bindLogCfgToConfig, bindHttpCfgToConfig, bindRateLimiterToConfig, bindOtelToConfig); err != nil {
		return nil, err
	}
	if err := validator.ValidateStruct(&conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func bindCfgToConfig(conf *Config, list ...func(conf *Config) error) error {
	for _, fn := range list {
		if err := fn(conf); err != nil {
			return err
		}
	}
	return nil
}

func registerCfg(list ...func() error) error {
	for _, fn := range list {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

func bindLogCfgToConfig(conf *Config) error {
	anyVal, ok := cfg.GetConfig(LoggingConfigName)
	if !ok {
		return fmt.Errorf("no configuration found for %s", LoggingConfigName)
	}
	logVal, ok := anyVal.(*genCfg.LoggingConfig)
	if !ok {
		return fmt.Errorf("logging config of unexpected type %s: expected %s",
			reflect.TypeOf(anyVal), "*genCfg.LoggingConfig",
		)
	}
	conf.Log = *logVal
	return nil
}

func bindHttpCfgToConfig(conf *Config) error {
	anyVal, ok := cfg.GetConfig(HttpConfigName)
	if !ok {
		return fmt.Errorf("no configuration found for %s", HttpConfigName)
	}
	httpVal, ok := anyVal.(*genCfg.HttpConfig)
	if !ok {
		return fmt.Errorf("logging config of unexpected type %s: expected %s",
			reflect.TypeOf(anyVal), "*genCfg.HttpConfig",
		)
	}
	conf.Http = *httpVal
	return nil
}

func bindRateLimiterToConfig(conf *Config) error {
	anyVal, ok := cfg.GetConfig(RateLimiterConfigName)
	if !ok {
		return fmt.Errorf("no configuration found for %s", RateLimiterConfigName)
	}
	RateLimiter, ok := anyVal.(*genCfg.RateLimiterConfig)
	if !ok {
		return fmt.Errorf("logging config of unexpected type %s: expected %s",
			reflect.TypeOf(anyVal), "*middleware.RateLimiterConfig",
		)
	}
	conf.RateLimiter = *RateLimiter
	return nil
}

func bindOtelToConfig(conf *Config) error {
	anyVal, ok := cfg.GetConfig(OtelConfigName)
	if !ok {
		return fmt.Errorf("no configuration found for %s", OtelConfigName)
	}
	otel, ok := anyVal.(*genCfg.OtelConfig)
	if !ok {
		return fmt.Errorf("logging config of unexpected type %s: expected %s",
			reflect.TypeOf(anyVal), "*OtelConfig",
		)
	}
	conf.Otel = *otel
	return nil
}

func registerLogCfg() error {
	return cfg.RegisterConfig(LoggingConfigName, cfg.ConfigEntry{
		Config: &genCfg.LoggingConfig{},
		BindArray: []cfg.BindValue{
			{
				ValName:    "log_format",
				DefaultVal: "text",
			},
			{
				ValName:    "log_level",
				DefaultVal: "debug",
			},
		},
	})
}

func registerRateLimiterCfg() error {
	return cfg.RegisterConfig(RateLimiterConfigName, cfg.ConfigEntry{
		Config: &genCfg.RateLimiterConfig{},
		BindArray: []cfg.BindValue{
			{
				ValName:    "rate_limiter_max_conn",
				DefaultVal: "100",
			},
			{
				ValName:    "rate_limiter_max_wait",
				DefaultVal: "100",
			},
			{
				ValName:    "rate_limiter_retry_after",
				DefaultVal: "1",
			},
		},
	})
}

func registerOtelCfg() error {
	return cfg.RegisterConfig(OtelConfigName, cfg.ConfigEntry{
		Config: &genCfg.OtelConfig{},
		BindArray: []cfg.BindValue{
			{
				ValName:    "otel_endpoint",
				DefaultVal: "",
			},
			{
				ValName:    "otel_shutdown_timeout",
				DefaultVal: "500ms",
			},
		},
	})
}

func registerHttpCfg() error {
	return cfg.RegisterConfig(HttpConfigName, cfg.ConfigEntry{
		Config: &genCfg.HttpConfig{},
		BindArray: []cfg.BindValue{
			{
				ValName:    "http_host",
				DefaultVal: "",
			},
			{
				ValName:    "http_port",
				DefaultVal: "",
			},
			{
				ValName:    "http_read_timeout",
				DefaultVal: "100ms",
			},
			{
				ValName:    "http_write_timeout",
				DefaultVal: "100ms",
			},
			{
				ValName:    "http_idle_timeout",
				DefaultVal: "100ms",
			},
			{
				ValName:    "http_shutdown_timeout",
				DefaultVal: "10s",
			},
		},
	})
}
