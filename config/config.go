package config

import (
	"log"

	"errors"
	"sync"

	"github.com/spf13/viper"
)

// Config : A struct that holds a bunch of configurations
type Config struct {
	DevServer           string
	DevPort             int
	DevUploadURLPattern string
	ServerFilesLocation string
}

var instance *Config
var once sync.Once

// GetInstance : Using singleton to get the global config instance
func GetInstance() *Config {
	once.Do(func() {
		validConfig, err := loadConfig()
		if err != nil {
			log.Fatal("Config not initialized")
		} else {
			instance = &validConfig
		}
	})
	return instance
}

// loadConfig : Load config from toml file and returns a config struct
func loadConfig() (Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("$GOPATH/src/github.com/tonychol/sink")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Config not found", err.Error())
		return Config{}, errors.New("Config not found")
	} else {
		devServer := viper.GetString("development.ServerAddr")
		devPort := viper.GetInt("development.Port")
		devUploadURLPattern := viper.GetString("development.UploadAddrPattern")
		ServerFilesLocation := viper.GetString("development.ServerFilesLocation")
		return Config{devServer,
			devPort,
			devUploadURLPattern,
			ServerFilesLocation,
		}, nil
	}

}
