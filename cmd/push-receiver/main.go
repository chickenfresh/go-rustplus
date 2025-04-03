package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chickenfresh/go-rustplus/fcm"
)

func main() {
	// Define command-line subcommands
	registerCmd := flag.NewFlagSet("register", flag.ExitOnError)
	listenCmd := flag.NewFlagSet("listen", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	// Register command flags
	registerSenderID := registerCmd.String("sender", "", "FCM sender ID")
	registerAndroid := registerCmd.Bool("android", false, "Use Android registration method")
	registerAPIKey := registerCmd.String("api-key", "", "Firebase API key (for Android method)")
	registerProjectID := registerCmd.String("project", "", "Firebase project ID (for Android method)")
	registerGMSAppID := registerCmd.String("app-id", "", "GMS app ID (for Android method)")
	registerPkgName := registerCmd.String("package", "", "Android package name (for Android method)")
	registerPkgCert := registerCmd.String("cert", "", "Android package certificate (for Android method)")
	registerOutput := registerCmd.String("output", "credentials.json", "Output file for credentials")

	// Listen command flags
	listenCredsFile := listenCmd.String("credentials", "credentials.json", "Credentials file path")
	listenPersistentIDs := listenCmd.String("persistent-ids", "", "Comma-separated list of persistent IDs to ignore")

	// Send command flags
	// sendCredsFile := sendCmd.String("credentials", "credentials.json", "Credentials file path")
	// sendTo := sendCmd.String("to", "", "FCM token to send to")
	// sendData := sendCmd.String("data", "{}", "JSON data to send")

	// Check if a subcommand is provided
	if len(os.Args) < 2 {
		fmt.Println("Expected 'register', 'listen', or 'send' subcommand")
		os.Exit(1)
	}

	// Parse the appropriate subcommand
	switch os.Args[1] {
	case "register":
		registerCmd.Parse(os.Args[2:])
	case "listen":
		listenCmd.Parse(os.Args[2:])
	case "send":
		sendCmd.Parse(os.Args[2:])
	default:
		fmt.Printf("Unknown subcommand: %s\n", os.Args[1])
		fmt.Println("Expected 'register', 'listen', or 'send' subcommand")
		os.Exit(1)
	}

	// Handle register command
	if registerCmd.Parsed() {
		var creds fcm.Credentials
		var err error

		if *registerAndroid {
			// Check required parameters
			if *registerAPIKey == "" || *registerProjectID == "" || *registerSenderID == "" ||
				*registerGMSAppID == "" || *registerPkgName == "" || *registerPkgCert == "" {
				log.Fatal("Android registration requires: api-key, project, sender, app-id, package, cert")
			}

			fmt.Println("Registering with FCM using Android method...")
			creds, err = fcm.RegisterAndroid(
				*registerAPIKey,
				*registerProjectID,
				*registerSenderID,
				*registerGMSAppID,
				*registerPkgName,
				*registerPkgCert,
			)
		} else {
			// Check required parameters
			if *registerSenderID == "" {
				log.Fatal("Standard registration requires: sender")
			}

			fmt.Println("Registering with FCM...")
			creds, err = fcm.Register(*registerSenderID)
		}

		if err != nil {
			log.Fatalf("Registration failed: %v", err)
		}

		// Save credentials to file
		data, err := json.MarshalIndent(creds, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal credentials: %v", err)
		}

		if err := os.WriteFile(*registerOutput, data, 0600); err != nil {
			log.Fatalf("Failed to write credentials file: %v", err)
		}

		fmt.Printf("Registration successful. Credentials saved to %s\n", *registerOutput)
		fmt.Printf("FCM Token: %s\n", creds.FCM.Token)
		fmt.Printf("Android ID: %s\n", creds.GCM.AndroidId)
		return
	}

	// Handle listen command
	if listenCmd.Parsed() {
		// Load credentials
		data, err := os.ReadFile(*listenCredsFile)
		if err != nil {
			log.Fatalf("Failed to read credentials file: %v", err)
		}

		var creds fcm.Credentials
		if err := json.Unmarshal(data, &creds); err != nil {
			log.Fatalf("Failed to parse credentials: %v", err)
		}

		// Parse persistent IDs if provided
		var persistentIDs []string
		if *listenPersistentIDs != "" {
			// In a real implementation, you would parse the comma-separated list
			// For simplicity, we're leaving this empty for now
		}

		fmt.Println("Connecting to FCM...")

		// Start listening
		notifChan, err := fcm.Listen(fcm.ListenConfig{
			Credentials:   creds,
			PersistentIds: persistentIDs,
		})

		if err != nil {
			log.Fatalf("Failed to start listening: %v", err)
		}

		fmt.Println("Listening for notifications...")

		// Handle graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Process notifications
		for {
			select {
			case notif := <-notifChan:
				data, _ := json.MarshalIndent(notif, "", "  ")
				fmt.Printf("Notification received at %s:\n%s\n",
					time.Now().Format(time.RFC3339), string(data))
			case <-sigChan:
				fmt.Println("Shutting down...")
				return
			}
		}
	}

	// Handle send command
	if sendCmd.Parsed() {
		// This is a placeholder for the send functionality
		// In a real implementation, you would implement FCM sending
		fmt.Println("Send functionality is not implemented yet")
		return
	}
}
