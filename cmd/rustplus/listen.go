package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chickenfresh/go-rustplus/fcm"
)

// fcmListen handles the FCM listen command
func fcmListen(configFile string) error {
	// Read config file
	config, err := readConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Make sure FCM credentials are in config
	if config.FCMCredentials.GCM.AndroidId == "" || config.FCMCredentials.GCM.SecurityToken == "" {
		return fmt.Errorf("FCM Credentials missing. Please run `fcm-register` first")
	}

	fmt.Println("Listening for FCM Notifications")

	// Start listening for notifications
	notifChan, err := fcm.Listen(fcm.ListenConfig{
		Credentials:   config.FCMCredentials,
		PersistentIds: []string{},
	})
	if err != nil {
		return fmt.Errorf("failed to start listening: %w", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Process notifications
	for {
		select {
		case notif := <-notifChan:
			// Generate timestamp
			timestamp := time.Now().Format("2006-01-02 15:04:05")

			// Log timestamp the notification was received (in green color)
			fmt.Printf("\033[32m[%s] Notification Received\033[0m\n", timestamp)

			// Log notification body
			data, _ := json.MarshalIndent(notif, "", "  ")
			fmt.Println(string(data))

		case <-sigChan:
			fmt.Println("Shutting down...")
			return nil
		}
	}
}
