// Package config loads the guestbook application configuration from
// environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	envAddress           = "GUESTBOOK_ADDRESS"
	envStorageDirectory  = "GUESTBOOK_STORAGE_DIR"
	envSimulatorEnabled  = "GUESTBOOK_SIMULATOR_ENABLED"
	envSimulatorInterval = "GUESTBOOK_SIMULATOR_INTERVAL"
)

const (
	defaultAddress           = ":8080"
	defaultStorageDirectory  = "./data"
	defaultSimulatorEnabled  = true
	defaultSimulatorInterval = 5 * time.Second
)

// Config holds the runtime configuration of the guestbook application.
type Config struct {
	Address           string
	StorageDirectory  string
	SimulatorEnabled  bool
	SimulatorInterval time.Duration
}

// Load reads the configuration from environment variables, falling back to
// defaults for any variable that is not set. It returns an error if a variable
// is set but cannot be parsed.
func Load() (config Config, err error) {
	simulatorEnabled, err := boolOrDefault(envSimulatorEnabled, defaultSimulatorEnabled)
	if err != nil {
		return Config{}, err
	}

	simulatorInterval, err := durationOrDefault(envSimulatorInterval, defaultSimulatorInterval)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Address:           stringOrDefault(envAddress, defaultAddress),
		StorageDirectory:  stringOrDefault(envStorageDirectory, defaultStorageDirectory),
		SimulatorEnabled:  simulatorEnabled,
		SimulatorInterval: simulatorInterval,
	}, nil
}

func stringOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func boolOrDefault(key string, fallback bool) (value bool, err error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}

	value, err = strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("parse boolean environment variable %s: %w", key, err)
	}

	return value, nil
}

func durationOrDefault(key string, fallback time.Duration) (value time.Duration, err error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}

	value, err = time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse duration environment variable %s: %w", key, err)
	}

	return value, nil
}
