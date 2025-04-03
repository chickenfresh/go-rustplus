package main

import (
	"fmt"

	"github.com/chickenfresh/go-rustplus/fcm"
)

// fcmRegister handles the FCM registration command
func fcmRegister(configFile string) error {
	fmt.Println("Registering with FCM")

	// Rust+ specific constants
	apiKey := "AIzaSyB5y2y-Tzqb4-I4Qnlsh_9naYv_TD8pCvY"
	projectID := "rust-companion-app"
	gcmSenderID := "976529667804"
	gmsAppID := "1:976529667804:android:d6f1ddeb4403b338fea619"
	androidPackageName := "com.facepunch.rust.companion"
	androidPackageCert := "E28D05345FB78A7A1A63D70F4A302DBF426CA5AD"

	// Register with FCM using Android method
	creds, err := fcm.RegisterAndroid(
		apiKey,
		projectID,
		gcmSenderID,
		gmsAppID,
		androidPackageName,
		androidPackageCert,
	)
	if err != nil {
		return fmt.Errorf("failed to register with FCM: %w", err)
	}

	fmt.Println("Successfully registered with FCM")
	fmt.Println("FCM Token:", creds.FCM.Token)

	// Get Expo Push Token
	fmt.Println("Fetching Expo Push Token")
	expoPushToken, err := getExpoPushToken(creds.FCM.Token)
	if err != nil {
		return fmt.Errorf("failed to get Expo Push Token: %w", err)
	}

	fmt.Println("Successfully fetched Expo Push Token")
	fmt.Println("Expo Push Token:", expoPushToken)

	// Link Steam with Rust+
	fmt.Println("Linking Steam account with Rust+")
	rustPlusAuthToken, err := linkSteamWithRustPlus()
	if err != nil {
		return fmt.Errorf("failed to link Steam account with Rust+: %w", err)
	}

	fmt.Println("Successfully linked Steam account with Rust+")

	// Register with Rust Companion API
	fmt.Println("Registering with Rust Companion API")
	err = registerWithRustPlus(rustPlusAuthToken, expoPushToken)
	if err != nil {
		return fmt.Errorf("failed to register with Rust Companion API: %w", err)
	}

	fmt.Println("Successfully registered with Rust Companion API")

	// Save to config
	config := Config{
		FCMCredentials:    creds,
		ExpoPushToken:     expoPushToken,
		RustPlusAuthToken: rustPlusAuthToken,
	}

	if err := saveConfig(configFile, config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("FCM, Expo and Rust+ auth tokens have been saved to %s\n", configFile)
	return nil
}
