package config

import (
	"github.com/spf13/viper"
)

var (
	// Config is the global configuration object
	Config *viper.Viper
)


// Init initializes the configuration
func InitConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	Config = viper.GetViper()
	return nil
}


