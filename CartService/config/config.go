package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Port             string `mapstructure:"PORT"`
	ConnectingString string `mapstructure:"CONNECTING_STRING"`
	ProductURL       string `mapstructure:"PRODUCT_URL"`
	ProductToken     string `mapstructure:"PRODUCT_TOKEN"`
	NetworkType      string `mapstructure:"NETWORK_TYPE"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("conf")
	viper.SetConfigType("env")

	err = viper.ReadInConfig()
	if err != nil {
		_ = fmt.Errorf("do not parse config file:%v", err)
	}
	fmt.Println(err)

	err = viper.Unmarshal(&config)

	return
}
