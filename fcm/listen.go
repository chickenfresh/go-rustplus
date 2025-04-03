package fcm

import (
	"fmt"
	"log"
	"sync"

	"github.com/chickenfresh/go-rustplus/fcm/client"
)

// ListenConfig contains configuration for listening to notifications
type ListenConfig struct {
	Credentials   Credentials
	PersistentIds []string
}

// Notification represents a received FCM notification
type Notification struct {
	Data         map[string]interface{}
	From         string
	CollapseKey  string
	PersistentId string
	// Other FCM notification fields
}

// Listen starts listening for FCM notifications
func Listen(config ListenConfig) (<-chan Notification, error) {
	// Create notification channel
	notificationChan := make(chan Notification)

	// Create FCM client
	fcmClient := client.NewClient(
		config.Credentials.GCM.AndroidId,
		config.Credentials.GCM.SecurityToken,
		config.PersistentIds,
	)

	// Set crypto credentials for decryption
	fcmClient.SetCredentials(
		config.Credentials.FCM.Keys.Private,
		config.Credentials.FCM.Keys.Auth,
	)

	// Set up notification handler
	var wg sync.WaitGroup
	wg.Add(1)

	fcmClient.OnConnect(func() {
		log.Println("Connected to FCM")
		wg.Done()
	})

	fcmClient.OnDisconnect(func() {
		log.Println("Disconnected from FCM")
	})

	fcmClient.OnNotification(func(n client.Notification) {
		// Check if we have a valid notification
		if n.Object == nil || n.Object.From == nil || n.Object.Token == nil {
			fmt.Printf("Warning: Received invalid notification: %+v\n", n)
			return
		}

		// Convert client notification to our notification format
		notification := Notification{
			Data:         n.Message,
			From:         *n.Object.From,
			CollapseKey:  *n.Object.Token,
			PersistentId: n.PersistentID,
		}

		// Send to channel
		notificationChan <- notification
	})

	// Handle connection errors
	fcmClient.OnError(func(err error) {
		fmt.Printf("FCM client error: %v\n", err)
	})

	// Connect to FCM
	if err := fcmClient.Connect(); err != nil {
		close(notificationChan)
		return nil, fmt.Errorf("failed to connect to FCM: %w", err)
	}

	// Wait for connection to be established
	wg.Wait()

	// Start a goroutine to handle cleanup when the channel is closed
	go func() {
		// When this goroutine exits, stop the client
		defer fcmClient.Stop()

		// Wait for the stop signal
		<-notificationChan
	}()

	return notificationChan, nil
}

// connectToFCM establishes a connection to FCM
func connectToFCM(credentials Credentials) error {
	// TODO: Implement actual FCM connection
	// This is a placeholder implementation

	return nil
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
