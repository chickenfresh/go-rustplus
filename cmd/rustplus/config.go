package main

import (
	"encoding/json"
	"os"

	"github.com/chickenfresh/go-rustplus/fcm"
)

// Config represents the configuration for the Rust+ CLI
type Config struct {
	FCMCredentials    fcm.Credentials `json:"fcm_credentials"`
	ExpoPushToken     string          `json:"expo_push_token"`
	RustPlusAuthToken string          `json:"rustplus_auth_token"`
}

// readConfig reads the configuration from the specified file
func readConfig(configFile string) (Config, error) {
	var config Config

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil // Return empty config if file doesn't exist
		}
		return config, err
	}

	err = json.Unmarshal(data, &config)
	return config, err
}

// saveConfig saves the configuration to the specified file
func saveConfig(configFile string, config Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0600)
}
