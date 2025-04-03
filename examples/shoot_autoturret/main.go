package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/chickenfresh/go-rustplus/rustplus"
)

func main() {
	if len(os.Args) < 6 {
		fmt.Println("Usage: shoot_autoturret <server_ip> <server_port> <player_id> <player_token> <turret_id>")
		os.Exit(1)
	}

	serverIP := os.Args[1]
	serverPort, _ := strconv.Atoi(os.Args[2])
	playerID, _ := strconv.ParseUint(os.Args[3], 10, 64)
	playerToken, _ := strconv.Atoi(os.Args[4])
	turretID := os.Args[5]

	// Create a new client
	client := rustplus.NewClient(serverIP, serverPort, playerID, playerToken, false)

	// Connect to the server
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Get a camera (auto turret)
	turret := client.GetCamera(turretID)

	// Subscribe to the camera
	if err := turret.Subscribe(); err != nil {
		log.Fatalf("Failed to subscribe to turret: %v", err)
	}
	defer turret.Unsubscribe()

	// Wait for subscription to complete
	for event := range turret.Events() {
		if event.Type == rustplus.CameraEventSubscribed {
			fmt.Println("Subscribed to auto turret")
			break
		} else if event.Type == rustplus.CameraEventError {
			log.Fatalf("Turret error: %v", event.Error)
		}
	}

	// Check if camera is an auto turret
	if !turret.IsAutoTurret() {
		log.Fatalf("Camera is not an auto turret!")
	}

	const shootCount = 3
	const shootDelay = 250 * time.Millisecond
	const moveDelay = 500 * time.Millisecond
	const moveAmount = 5

	// Shoot auto turret
	fmt.Println("Shooting...")
	for i := 0; i < shootCount; i++ {
		time.Sleep(shootDelay)
		if err := turret.Shoot(); err != nil {
			log.Fatalf("Failed to shoot: %v", err)
		}
	}

	// Move auto turret left
	fmt.Println("Moving left...")
	time.Sleep(moveDelay)
	if err := turret.Move(rustplus.ButtonNone, -moveAmount, 0); err != nil {
		log.Fatalf("Failed to move: %v", err)
	}

	// Shoot auto turret
	fmt.Println("Shooting...")
	for i := 0; i < shootCount; i++ {
		time.Sleep(shootDelay)
		if err := turret.Shoot(); err != nil {
			log.Fatalf("Failed to shoot: %v", err)
		}
	}

	// Move auto turret up
	fmt.Println("Moving up...")
	time.Sleep(moveDelay)
	if err := turret.Move(rustplus.ButtonNone, 0, moveAmount); err != nil {
		log.Fatalf("Failed to move: %v", err)
	}

	// Shoot auto turret
	fmt.Println("Shooting...")
	for i := 0; i < shootCount; i++ {
		time.Sleep(shootDelay)
		if err := turret.Shoot(); err != nil {
			log.Fatalf("Failed to shoot: %v", err)
		}
	}

	// Move auto turret right
	fmt.Println("Moving right...")
	time.Sleep(moveDelay)
	if err := turret.Move(rustplus.ButtonNone, moveAmount, 0); err != nil {
		log.Fatalf("Failed to move: %v", err)
	}

	// Shoot auto turret
	fmt.Println("Shooting...")
	for i := 0; i < shootCount; i++ {
		time.Sleep(shootDelay)
		if err := turret.Shoot(); err != nil {
			log.Fatalf("Failed to shoot: %v", err)
		}
	}

	// Move auto turret down
	fmt.Println("Moving down...")
	time.Sleep(moveDelay)
	if err := turret.Move(rustplus.ButtonNone, 0, -moveAmount); err != nil {
		log.Fatalf("Failed to move: %v", err)
	}

	// Shoot auto turret
	fmt.Println("Shooting...")
	for i := 0; i < shootCount; i++ {
		time.Sleep(shootDelay)
		if err := turret.Shoot(); err != nil {
			log.Fatalf("Failed to shoot: %v", err)
		}
	}

	fmt.Println("Auto turret sequence complete")
}
