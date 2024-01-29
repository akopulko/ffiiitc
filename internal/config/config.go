package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-pkgz/lgr"
)

const (
	FireflyAppTimeout = 10               // 10 sec for fftc to app service timeout
	ModelFile         = "data/model.gob" //file name to store model
	apiKeyEnvVar      = "FF_API_KEY"
	appUrlEnvVar      = "FF_APP_URL"
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

func EnvVarIsSet(varName string) bool {
	return os.Getenv(varName) != ""
}

func LookupEnvVarValueFromFile(path string, logger *lgr.Logger) (string, bool) {
	if path == "" {
		logger.Logf("WARN file path is empty!")
		return "", false
	}

	logger.Logf("DEBUG reading file...")
	valueBytes, e := os.ReadFile(path)

	if e != nil {
		panic(e)
	}
	value := string(valueBytes)
	value = strings.TrimSuffix(value, "\n")

	return string(value), true
}

func LookupEnvVar(variableName string, logger *lgr.Logger) (string, bool) {
	logger.Logf("DEBUG looking for env var %s", variableName)

	// Try get value from a file, like Docker secrets.
	if EnvVarIsSet(variableName + "_FILE") {
		logger.Logf("DEBUG var %s is set as file, getting file path...", variableName)
		path := os.Getenv(variableName + "_FILE")
		logger.Logf("DEBUG file path is '%s'", path)
		return LookupEnvVarValueFromFile(path, logger)
	}

	// Try get value ist stored in variable directly.
	logger.Logf("DEBUG extracting value directly from env var")
	return os.LookupEnv(variableName)
}

func FormatEnvNotSetErrorMessage(variableName string) string {
	return fmt.Sprintf("Environment vars '%s' or '%s' not set!", variableName, variableName+"_FILE")
}

func NewConfig(logger *lgr.Logger) (*Config, error) {
	apiKey, apiKeyExists := LookupEnvVar(apiKeyEnvVar, logger)

	if !apiKeyExists {
		return nil, errors.New(FormatEnvNotSetErrorMessage(apiKeyEnvVar))
	}

	appUrl, appUrlExists := LookupEnvVar(appUrlEnvVar, logger)
	if !appUrlExists {
		return nil, errors.New(FormatEnvNotSetErrorMessage(appUrlEnvVar))
	}

	cfg := Config{
		APIKey: apiKey,
		FFApp:  appUrl,
	}

	return &cfg, nil
}
