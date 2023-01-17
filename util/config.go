package util

import (
	"github.com/spf13/viper"
)

// JSON to Config
type Config struct {
	RequestUrl string `mapstructure:"RequestUrl"`
	RequestNum int    `mapstructure:"RequestNum"`
	WorkerNum  int    `mapstructure:"WorkerNum"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
