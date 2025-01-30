package logger_conf

import (
	"github.com/KennyMacCormik/common/log"
	"github.com/KennyMacCormik/common/val"
	"github.com/KennyMacCormik/otel/otel-common/conf"
	"github.com/spf13/viper"
)

type loggerConf struct {
	LogFormat string `mapstructure:"log_format" validate:"oneof=text json"`
	LogLevel  string `mapstructure:"log_level" validate:"oneof=debug info warn error"`
}

func NewLoggerConf() conf.LoggerConf {
	c := &loggerConf{}

	viper.SetDefault("log_format", "text")
	err := viper.BindEnv("log_format")
	if err != nil {
		log.Error("Failed to bind log_format")
	}

	viper.SetDefault("log_level", "info")
	err = viper.BindEnv("log_level")
	if err != nil {
		log.Error("Failed to bind log_level")
	}

	err = viper.Unmarshal(c)
	if err != nil {
		log.Error("Failed to unmarshal loggerConf")
	}

	err = val.ValidateStruct(*c)
	if err != nil {
		log.Error("Failed to validate loggerConf", "err", err)
	}

	if err != nil {
		return nil
	}

	return c
}

func (l *loggerConf) Format() string {
	return l.LogFormat
}

func (l *loggerConf) Level() string {
	return l.LogLevel
}
