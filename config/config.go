package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"

	"github.com/spf13/viper"
)

type Config struct {
	Cleint struct {
		ClientId     string `mapstructure:"client_id"`
		ClientSecret string `mapstructure:"client_secret"`
	} `mapstructure:"client"`
	Proxy struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"proxy"`
	MySQL []*MySqlConfig `mapstructure:"mysql"`
}

type MySqlConfig struct {
	Addr      string `mapstructure:"addr,omitempty" json:"addr,omitempty" `
	Username  string `mapstructure:"username,omitempty" json:"username,omitempty" `
	Password  string `mapstructure:"password,omitempty" json:"password,omitempty" `
	Database  string `mapstructure:"database,omitempty" json:"database,omitempty" `
	ConfigKey string `mapstructure:"config_key,omitempty" json:"config_key,omitempty" `
}

var config = &Config{}
var configName = "config"
var configType = "yaml"

func GetConfig() *Config {
	return config
}

func init() {
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logger := zap.L()
		logger.Info(fmt.Sprintf("Config file changed: %s", e.Name))
		err := viper.Unmarshal(config)
		if err != nil {
			panic(err)
		}
	})
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(config)
	if err != nil {
		panic(err)
	}
}
