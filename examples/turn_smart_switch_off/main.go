package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/chickenfresh/go-rustplus/rustplus"
)

func main() {
	if len(os.Args) < 6 {
		fmt.Println("Usage: turn_smart_switch_off <server_ip> <server_port> <player_id> <player_token> <entity_id>")
		os.Exit(1)
	}

	serverIP := os.Args[1]
	serverPort, _ := strconv.Atoi(os.Args[2])
	playerID, _ := strconv.ParseUint(os.Args[3], 10, 64)
	playerToken, _ := strconv.Atoi(os.Args[4])
	entityID, _ := strconv.ParseUint(os.Args[5], 10, 32)

	// Create a new client
	client := rustplus.NewClient(serverIP, serverPort, playerID, playerToken, false)

	// Connect to the server
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Turn smart switch off
	fmt.Printf("Turning smart switch %d off...\n", entityID)
	if err := client.SetEntityValue(uint32(entityID), false); err != nil {
		log.Fatalf("Failed to turn off smart switch: %v", err)
	}

	fmt.Println("Smart switch turned off successfully!")
}
