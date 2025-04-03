package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chickenfresh/go-rustplus/fcm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fcm_register_listen <sender_id>")
		os.Exit(1)
	}

	senderId := os.Args[1]
	credentialsFile := "credentials.json"
	persistentIdsFile := "persistent_ids.json"

	// Check if credentials file exists
	credentials, err := loadCredentials(credentialsFile)
	if err != nil || credentials.FCM.Token == "" {
		fmt.Println("No valid credentials found, registering with FCM...")

		// Register with FCM
		credentials, err = fcm.Register(senderId)
		if err != nil {
			log.Fatalf("Failed to register with FCM: %v", err)
		}

		// Save credentials
		if err := saveCredentials(credentialsFile, credentials); err != nil {
			log.Fatalf("Failed to save credentials: %v", err)
		}

		fmt.Println("Successfully registered with FCM")
		fmt.Printf("FCM Token: %s\n", credentials.FCM.Token)
	} else {
		fmt.Println("Using existing credentials")
		fmt.Printf("FCM Token: %s\n", credentials.FCM.Token)
	}

	// Load persistent IDs
	var persistentIds []string
	if _, err := os.Stat(persistentIdsFile); err == nil {
		persistentIds = loadPersistentIds(persistentIdsFile)
		fmt.Printf("Loaded %d persistent IDs\n", len(persistentIds))
	}

	// Listen for notifications
	fmt.Println("Listening for notifications...")
	notifications, err := fcm.Listen(fcm.ListenConfig{
		Credentials:   credentials,
		PersistentIds: persistentIds,
	})

	if err != nil {
		log.Fatalf("Failed to start listening: %v", err)
	}

	// Set up channel for handling Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Track persistent IDs
	currentPersistentIds := make([]string, 0)
	if len(persistentIds) > 0 {
		currentPersistentIds = append(currentPersistentIds, persistentIds...)
	}

	// Process notifications until interrupted
	fmt.Println("Press Ctrl+C to exit")
	for {
		select {
		case notification := <-notifications:
			fmt.Printf("\nReceived notification: %+v\n", notification)

			// Add persistent ID to list
			if notification.PersistentId != "" {
				currentPersistentIds = append(currentPersistentIds, notification.PersistentId)
			}

		case <-sigChan:
			fmt.Println("\nShutting down...")

			// Save persistent IDs
			if err := savePersistentIds(persistentIdsFile, currentPersistentIds); err != nil {
				log.Fatalf("Failed to save persistent IDs: %v", err)
			}

			fmt.Printf("Saved %d persistent IDs\n", len(currentPersistentIds))
			return
		}
	}
}

// loadCredentials loads FCM credentials from a file
func loadCredentials(filename string) (fcm.Credentials, error) {
	var credentials fcm.Credentials

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return credentials, err
	}

	err = json.Unmarshal(data, &credentials)
	return credentials, err
}

// saveCredentials saves FCM credentials to a file
func saveCredentials(filename string, credentials fcm.Credentials) error {
	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0600)
}

// savePersistentIds saves persistent IDs to a file
func savePersistentIds(filename string, persistentIds []string) error {
	data, err := json.MarshalIndent(persistentIds, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0600)
}

// loadPersistentIds loads persistent IDs from a file
func loadPersistentIds(filename string) []string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read persistent IDs file: %v", err)
	}

	var persistentIds []string
	err = json.Unmarshal(data, &persistentIds)
	if err != nil {
		log.Fatalf("Failed to unmarshal persistent IDs: %v", err)
	}

	return persistentIds
}
