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
		fmt.Println("Usage: move_ptz_camera <server_ip> <server_port> <player_id> <player_token> <camera_id>")
		os.Exit(1)
	}

	serverIP := os.Args[1]
	serverPort, _ := strconv.Atoi(os.Args[2])
	playerID, _ := strconv.ParseUint(os.Args[3], 10, 64)
	playerToken, _ := strconv.Atoi(os.Args[4])
	cameraID := os.Args[5]

	// Create a new client
	client := rustplus.NewClient(serverIP, serverPort, playerID, playerToken, false)

	// Connect to the server
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Get a camera
	camera := client.GetCamera(cameraID)

	// Subscribe to the camera
	if err := camera.Subscribe(); err != nil {
		log.Fatalf("Failed to subscribe to camera: %v", err)
	}
	defer camera.Unsubscribe()

	// Wait for subscription to complete
	for event := range camera.Events() {
		if event.Type == rustplus.CameraEventSubscribed {
			fmt.Println("Subscribed to camera")
			break
		} else if event.Type == rustplus.CameraEventError {
			log.Fatalf("Camera error: %v", event.Error)
		}
	}

	// Move camera up 10 times
	fmt.Println("Moving camera up...")
	for i := 0; i < 10; i++ {
		if err := camera.Move(rustplus.ButtonNone, 0, 1); err != nil {
			log.Fatalf("Failed to move camera: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Move camera down 10 times
	fmt.Println("Moving camera down...")
	for i := 0; i < 10; i++ {
		if err := camera.Move(rustplus.ButtonNone, 0, -1); err != nil {
			log.Fatalf("Failed to move camera: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Move camera left 10 times
	fmt.Println("Moving camera left...")
	for i := 0; i < 10; i++ {
		if err := camera.Move(rustplus.ButtonNone, -1, 0); err != nil {
			log.Fatalf("Failed to move camera: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Move camera right 10 times
	fmt.Println("Moving camera right...")
	for i := 0; i < 10; i++ {
		if err := camera.Move(rustplus.ButtonNone, 1, 0); err != nil {
			log.Fatalf("Failed to move camera: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("Camera movement complete")
}
