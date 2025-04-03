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
		fmt.Println("Usage: zoom_ptz_camera <server_ip> <server_port> <player_id> <player_token> <camera_id>")
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

	// Zoom camera 8 times
	fmt.Println("Zooming camera...")
	for i := 0; i < 8; i++ {
		if err := camera.Zoom(); err != nil {
			log.Fatalf("Failed to zoom camera: %v", err)
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Camera zoom complete")
}
