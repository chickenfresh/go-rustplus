package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
)

// fcmSend handles the FCM send command
func fcmSend(configFile string, args []string) error {
	// Parse command-specific flags
	sendCmd := flag.NewFlagSet("fcm-send", flag.ExitOnError)
	serverKey := sendCmd.String("server-key", "", "FCM server key")
	token := sendCmd.String("token", "", "FCM token to send to")
	title := sendCmd.String("title", "Hello world", "Notification title")
	body := sendCmd.String("body", "Test message", "Notification body")

	// Parse the flags
	if err := sendCmd.Parse(args); err != nil {
		return err
	}

	// Check required flags
	if *serverKey == "" {
		return fmt.Errorf("missing required flag: -server-key")
	}

	if *token == "" {
		// Try to get token from config
		config, err := readConfig(configFile)
		if err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}

		if config.FCMCredentials.FCM.Token == "" {
			return fmt.Errorf("missing required flag: -token (and no token found in config)")
		}

		*token = config.FCMCredentials.FCM.Token
	}

	// Create request body
	requestBody, err := json.Marshal(map[string]interface{}{
		"to": *token,
		"notification": map[string]string{
			"title": *title,
			"body":  *body,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create request body: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("key=%s", *serverKey))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Print response
	fmt.Println(string(responseBody))

	return nil
}
