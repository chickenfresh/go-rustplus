package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chickenfresh/go-rustplus/fcm"
)

// TestConfigReadWrite tests the config read/write functionality
func TestConfigReadWrite(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "rustplus-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file path
	configFile := filepath.Join(tempDir, "test-config.json")

	// Create a test config
	testConfig := Config{
		FCMCredentials: fcm.Credentials{
			FCM: fcm.FCMCredentials{
				Token: "test-fcm-token",
			},
			GCM: fcm.GCMCredentials{
				AndroidId:     "test-android-id",
				SecurityToken: "test-security-token",
			},
		},
		ExpoPushToken:     "test-expo-token",
		RustPlusAuthToken: "test-auth-token",
	}

	// Save the config
	err = saveConfig(configFile, testConfig)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Read the config
	readConfig, err := readConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Verify the config was read correctly
	if readConfig.FCMCredentials.FCM.Token != testConfig.FCMCredentials.FCM.Token {
		t.Errorf("Expected FCM token %s, got %s", testConfig.FCMCredentials.FCM.Token, readConfig.FCMCredentials.FCM.Token)
	}
	if readConfig.FCMCredentials.GCM.AndroidId != testConfig.FCMCredentials.GCM.AndroidId {
		t.Errorf("Expected Android ID %s, got %s", testConfig.FCMCredentials.GCM.AndroidId, readConfig.FCMCredentials.GCM.AndroidId)
	}
	if readConfig.FCMCredentials.GCM.SecurityToken != testConfig.FCMCredentials.GCM.SecurityToken {
		t.Errorf("Expected security token %s, got %s", testConfig.FCMCredentials.GCM.SecurityToken, readConfig.FCMCredentials.GCM.SecurityToken)
	}
	if readConfig.ExpoPushToken != testConfig.ExpoPushToken {
		t.Errorf("Expected Expo token %s, got %s", testConfig.ExpoPushToken, readConfig.ExpoPushToken)
	}
	if readConfig.RustPlusAuthToken != testConfig.RustPlusAuthToken {
		t.Errorf("Expected auth token %s, got %s", testConfig.RustPlusAuthToken, readConfig.RustPlusAuthToken)
	}
}

// TestGetExpoPushToken tests the getExpoPushToken function
func TestGetExpoPushToken(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		// Parse request body
		var request map[string]interface{}
		err = json.Unmarshal(body, &request)
		if err != nil {
			t.Fatalf("Failed to parse request body: %v", err)
		}

		// Verify request fields
		if request["type"] != "fcm" {
			t.Errorf("Expected type fcm, got %v", request["type"])
		}
		if request["appId"] != rustPlusAppID {
			t.Errorf("Expected appId %s, got %v", rustPlusAppID, request["appId"])
		}
		if request["projectId"] != rustPlusProjectID {
			t.Errorf("Expected projectId %s, got %v", rustPlusProjectID, request["projectId"])
		}
		if request["deviceToken"] != "test-fcm-token" {
			t.Errorf("Expected deviceToken test-fcm-token, got %v", request["deviceToken"])
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{"data":{"expoPushToken":"ExponentPushToken[test-expo-token]"}}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Save the original endpoint and restore it after the test
	originalEndpoint := expoAPIEndpoint
	expoAPIEndpoint = server.URL
	defer func() { expoAPIEndpoint = originalEndpoint }()

	// Call the function
	token, err := getExpoPushToken("test-fcm-token")
	if err != nil {
		t.Fatalf("Failed to get Expo push token: %v", err)
	}

	// Verify the token
	expectedToken := "ExponentPushToken[test-expo-token]"
	if token != expectedToken {
		t.Errorf("Expected token %s, got %s", expectedToken, token)
	}
}

// TestRegisterWithRustPlus tests the registerWithRustPlus function
func TestRegisterWithRustPlus(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		// Parse request body
		var request map[string]interface{}
		err = json.Unmarshal(body, &request)
		if err != nil {
			t.Fatalf("Failed to parse request body: %v", err)
		}

		// Verify request fields
		if request["AuthToken"] != "test-auth-token" {
			t.Errorf("Expected AuthToken test-auth-token, got %v", request["AuthToken"])
		}
		if request["DeviceId"] != "rustplus.js" {
			t.Errorf("Expected DeviceId rustplus.js, got %v", request["DeviceId"])
		}
		if request["PushKind"] != float64(3) {
			t.Errorf("Expected PushKind 3, got %v", request["PushKind"])
		}
		if request["PushToken"] != "test-expo-token" {
			t.Errorf("Expected PushToken test-expo-token, got %v", request["PushToken"])
		}

		// Send response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	// Save the original endpoint and restore it after the test
	originalEndpoint := rustPlusAPIEndpoint
	rustPlusAPIEndpoint = server.URL
	defer func() { rustPlusAPIEndpoint = originalEndpoint }()

	// Call the function
	err := registerWithRustPlus("test-auth-token", "test-expo-token")
	if err != nil {
		t.Fatalf("Failed to register with Rust+: %v", err)
	}
}

// TestCallbackHandler tests the callback handler
func TestCallbackHandler(t *testing.T) {
	// Create a token channel
	tokenChan := make(chan string, 1)

	// Create a shutdown channel
	shutdownChan := make(chan struct{})

	// Create a safe close function
	safeCloseChannel := func() {
		close(shutdownChan)
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle the callback
		var token string

		// Check if it's a GET or POST request
		if r.Method == "POST" {
			// For POST requests (from axios)
			var requestData struct {
				Token string `json:"token"`
			}

			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&requestData); err == nil && requestData.Token != "" {
				token = requestData.Token
			} else {
				// If JSON parsing fails, try form data
				if err := r.ParseForm(); err != nil {
					http.Error(w, "Failed to parse form data", http.StatusBadRequest)
					return
				}
				token = r.FormValue("token")
			}

			// Set CORS headers for POST requests
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		} else if r.Method == "OPTIONS" {
			// Handle preflight requests
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			return
		} else {
			// For GET requests (from redirects)
			token = r.URL.Query().Get("token")
		}

		if token == "" {
			http.Error(w, "Token missing from request!", http.StatusBadRequest)
			return
		}

		// For POST requests, just send a simple success response
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success": true}`))
		} else {
			// For GET requests, send a simple success message
			w.Write([]byte("Success"))
		}

		// Send the token to the channel
		select {
		case tokenChan <- token:
			// Signal server to shutdown
			safeCloseChannel()
		default:
			t.Error("Token channel is full, this is unexpected")
		}
	}))
	defer server.Close()

	// Test POST request
	postReq, err := http.NewRequest("POST", server.URL, bytes.NewBufferString(`{"token":"test-token-post"}`))
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	postReq.Header.Set("Content-Type", "application/json")

	// Send POST request
	client := &http.Client{}
	_, err = client.Do(postReq)
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}

	// Wait for token or timeout
	select {
	case token := <-tokenChan:
		if token != "test-token-post" {
			t.Errorf("Expected token test-token-post, got %s", token)
		}
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for token")
	}

	// Reset channels
	tokenChan = make(chan string, 1)
	shutdownChan = make(chan struct{})

	// Test GET request
	_, err = http.Get(server.URL + "?token=test-token-get")
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}

	// Wait for token or timeout
	select {
	case token := <-tokenChan:
		if token != "test-token-get" {
			t.Errorf("Expected token test-token-get, got %s", token)
		}
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for token")
	}
}
