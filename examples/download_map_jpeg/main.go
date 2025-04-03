package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/chickenfresh/go-rustplus/rustplus"
)

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: download_map_jpeg <server_ip> <server_port> <player_id> <player_token> [output_file]")
		os.Exit(1)
	}

	serverIP := os.Args[1]
	serverPort, _ := strconv.Atoi(os.Args[2])
	playerID, _ := strconv.ParseUint(os.Args[3], 10, 64)
	playerToken, _ := strconv.Atoi(os.Args[4])

	outputFile := "map.jpg"
	if len(os.Args) > 5 {
		outputFile = os.Args[5]
	}

	// Create a new client
	client := rustplus.NewClient(serverIP, serverPort, playerID, playerToken, false)

	// Connect to the server
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Get the map
	fmt.Println("Getting map...")
	mapData, err := client.GetMap()
	if err != nil {
		log.Fatalf("Failed to get map: %v", err)
	}

	// Save the map image
	if err := os.WriteFile(outputFile, mapData.GetJpgImage(), 0644); err != nil {
		log.Fatalf("Failed to save map image: %v", err)
	}

	fmt.Printf("Map saved to %s\n", outputFile)
}
