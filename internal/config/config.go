package config

import (
	"errors"
	"os"
)

const (
	FireflyAppTimeout = 10               // 10 sec for fftc to app service timeout
	ModelFile         = "data/model.gob" //file name to store model
)

type Config struct {
	APIKey string
	FFApp  string
}

var envVars = []string{
	"FF_API_KEY",
	"FF_APP_URL",
}

func EnvVarExist(varName string) bool {
	_, present := os.LookupEnv(varName)
	return present
}

func NewConfig() (*Config, error) {
	for _, val := range envVars {
		exist := EnvVarExist(val)
		if !exist {
			return nil, errors.New("env variable is not set: " + val)
		}
	}

	cfg := Config{
		APIKey: os.Getenv("FF_API_KEY"),
		FFApp:  os.Getenv("FF_APP_URL"),
	}

	return &cfg, nil
}
