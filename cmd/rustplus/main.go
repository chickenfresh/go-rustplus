package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	rustPlusAPIEndpoint = "https://companion-rust.facepunch.com:443/api/push/register"
)

// showUsage displays the usage information
func showUsage() {
	fmt.Println("RustPlus - A command line tool for things related to Rust+")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  rustplus <options> <command>")
	fmt.Println()
	fmt.Println("Command List:")
	fmt.Println("  help           Print this usage guide")
	fmt.Println("  fcm-register   Registers with FCM, Expo and links your Steam account with Rust+")
	fmt.Println("  fcm-listen     Listens to notifications received from FCM, such as Rust+ Pairing Notifications")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --config-file  Path to config file (default: rustplus.config.json)")
}

func main() {
	// Parse command line flags
	configFileFlag := flag.String("config-file", "", "Path to config file")
	flag.Parse()

	// Get the command (first non-flag argument)
	args := flag.Args()
	if len(args) == 0 {
		showUsage()
		os.Exit(1)
	}

	command := args[0]

	// Determine config file path
	configFile := *configFileFlag
	if configFile == "" {
		configFile = filepath.Join(".", "rustplus.config.json")
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Handle shutdown in a goroutine
	go func() {
		<-sigChan
		fmt.Println("Shutting down...")
		os.Exit(0)
	}()

	// Execute the appropriate command
	var err error
	switch command {
	case "fcm-register":
		err = fcmRegister(configFile)
	case "fcm-listen":
		err = fcmListen(configFile)
	case "help":
		showUsage()
	default:
		showUsage()
		os.Exit(1)
	}

	// Handle errors
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
