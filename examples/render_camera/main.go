package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"strconv"

	"github.com/chickenfresh/go-rustplus/rustplus"
)

func main() {
	if len(os.Args) < 6 {
		fmt.Println("Usage: render_camera <server_ip> <server_port> <player_id> <player_token> <camera_id> [output_file]")
		os.Exit(1)
	}

	serverIP := os.Args[1]
	serverPort, _ := strconv.Atoi(os.Args[2])
	playerID, _ := strconv.ParseUint(os.Args[3], 10, 64)
	playerToken, _ := strconv.Atoi(os.Args[4])
	cameraID := os.Args[5]

	outputFile := "camera.png"
	if len(os.Args) > 6 {
		outputFile = os.Args[6]
	}

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

	// Wait for a camera frame
	fmt.Println("Waiting for camera frame...")
	for event := range camera.Events() {
		if event.Type == rustplus.CameraEventRender {
			fmt.Println("Received camera frame")

			// Save the frame
			f, err := os.Create(outputFile)
			if err != nil {
				log.Fatalf("Failed to create output file: %v", err)
			}

			if err := png.Encode(f, event.Data.(image.Image)); err != nil {
				f.Close()
				log.Fatalf("Failed to encode image: %v", err)
			}

			f.Close()
			fmt.Printf("Camera frame saved to %s\n", outputFile)
			break
		} else if event.Type == rustplus.CameraEventError {
			log.Fatalf("Camera error: %v", event.Error)
		}
	}
}
