package config

import (
	"os"
	"testing"

	"github.com/go-pkgz/lgr"
)

func TestEnvVarExist(t *testing.T) {
	t.Run("ExistingEnvVar", func(t *testing.T) {
		// Set up a temporary environment variable for testing
		os.Setenv("TEST_ENV_VAR", "value")

		// Check if the environment variable exists
		if !EnvVarExist("TEST_ENV_VAR") {
			t.Error("Expected environment variable to exist, but it does not")
		}

		// Clean up the environment variable after the test
		os.Unsetenv("TEST_ENV_VAR")
	})

	t.Run("NonExistingEnvVar", func(t *testing.T) {
		// Check if the environment variable exists
		if EnvVarExist("NON_EXISTING_ENV_VAR") {
			t.Error("Expected environment variable to not exist, but it does")
		}
	})
}

func TestNewConfig(t *testing.T) {
	logger := lgr.New(lgr.Debug, lgr.CallerFunc)

	t.Run("AllEnvVarsExist", func(t *testing.T) {

		// Set up temporary environment variables for testing
		os.Setenv("FF_API_KEY", "test_api_key")
		os.Setenv("FF_APP_URL", "test_app_url")

		// Create a new config
		cfg, err := NewConfig(logger)

		// Check if there is no error
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		// Check if the APIKey and FFApp fields are correct
		if cfg.APIKey != "test_api_key" {
			t.Errorf("Expected APIKey to be 'test_api_key', but got: %s", cfg.APIKey)
		}
		if cfg.FFApp != "test_app_url" {
			t.Errorf("Expected FFApp to be 'test_app_url', but got: %s", cfg.FFApp)
		}

		// Clean up the environment variables after the test
		os.Unsetenv("FF_API_KEY")
		os.Unsetenv("FF_APP_URL")
	})

	t.Run("MissingEnvVars", func(t *testing.T) {
		// Create a new config
		cfg, err := NewConfig(logger)

		// Check if there is an error
		if err == nil {
			t.Error("Expected error due to missing environment variables, but got no error")
		}

		// Check if the config is nil
		if cfg != nil {
			t.Error("Expected config to be nil, but it is not")
		}
	})
}
