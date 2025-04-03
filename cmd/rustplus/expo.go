package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

var (
	expoAPIEndpoint   = "https://exp.host/--/api/v2/push/getExpoPushToken"
	rustPlusAppID     = "com.facepunch.rust.companion"
	rustPlusProjectID = "49451aca-a822-41e6-ad59-955718d0ff9c"
)

// getExpoPushToken gets an Expo Push Token for the FCM token
func getExpoPushToken(fcmToken string) (string, error) {
	// Create request body
	requestBody, err := json.Marshal(map[string]interface{}{
		"type":        "fcm",
		"deviceId":    uuid.New().String(),
		"development": false,
		"appId":       rustPlusAppID,
		"deviceToken": fcmToken,
		"projectId":   rustPlusProjectID,
	})
	if err != nil {
		return "", err
	}

	// Create request
	req, err := http.NewRequest("POST", expoAPIEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse response
	var response struct {
		Data struct {
			ExpoPushToken string `json:"expoPushToken"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return response.Data.ExpoPushToken, nil
}

// registerWithRustPlus registers with the Rust+ API
func registerWithRustPlus(authToken, expoPushToken string) error {
	rustPlusAPIEndpoint := "https://companion-rust.facepunch.com:443/api/push/register"

	// Create request body
	requestBody, err := json.Marshal(map[string]interface{}{
		"AuthToken": authToken,
		"DeviceId":  "rustplus.js",
		"PushKind":  3,
		"PushToken": expoPushToken,
	})
	if err != nil {
		return err
	}

	// Create request
	req, err := http.NewRequest("POST", rustPlusAPIEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
