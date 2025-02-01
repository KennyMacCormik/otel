package gin_conf

import (
	"github.com/KennyMacCormik/common/log"
	"github.com/KennyMacCormik/common/val"
	"github.com/KennyMacCormik/otel/otel-common/conf"
	"github.com/spf13/viper"
)

type ginConf struct {
	GinMode string `mapstructure:"gin_mode" validate:"oneof=debug,release,test"`
}

func NewGinConf() conf.GinConfig {
	c := &ginConf{}

	viper.SetDefault("gin_mode", "release")
	err := viper.BindEnv("gin_mode")
	if err != nil {
		log.Error("Failed to bind gin_mode")
	}

	err = val.ValidateStruct(c)
	if err != nil {
		log.Error("Failed to validate loggerConf", "err", err)
	}

	if err != nil {
		return nil
	}

	return c
}

func (l *ginConf) Mode() string {
	return l.GinMode
}
