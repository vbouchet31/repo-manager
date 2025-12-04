package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Organization string   `mapstructure:"organization"`
	Prefix       string   `mapstructure:"prefix"`
	Users        []string `mapstructure:"users"`
}

var AppConfig Config

func LoadConfig(cfgFile string) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.meo-repo-manager")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			return fmt.Errorf("config file not found")
		} else {
			// Config file was found but another error was produced
			return fmt.Errorf("fatal error config file: %w", err)
		}
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return fmt.Errorf("unable to decode into struct: %w", err)
	}

	return nil
}
