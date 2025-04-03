package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/chickenfresh/go-rustplus/rustplus"
	"github.com/chickenfresh/go-rustplus/rustplus/proto"
)

func main() {
	if len(os.Args) < 6 {
		fmt.Println("Usage: entity_changed_broadcasts <server_ip> <server_port> <player_id> <player_token> <entity_id>")
		os.Exit(1)
	}

	serverIP := os.Args[1]
	serverPort, _ := strconv.Atoi(os.Args[2])
	playerID, _ := strconv.ParseUint(os.Args[3], 10, 64)
	playerToken, _ := strconv.Atoi(os.Args[4])
	entityID, _ := strconv.ParseUint(os.Args[5], 10, 32)

	// Create a new client
	client := rustplus.NewClient(serverIP, serverPort, playerID, playerToken, false)

	// Set up entity changed handler
	client.AddMessageHandler(func(msg *proto.AppMessage) bool {
		if msg.Broadcast != nil && msg.Broadcast.EntityChanged != nil {
			entityChanged := msg.Broadcast.EntityChanged
			fmt.Printf("Entity %d is now %s\n",
				entityChanged.GetEntityId(),
				map[bool]string{true: "active", false: "inactive"}[entityChanged.GetPayload().GetValue()])
			return true
		}
		return false
	})

	// Connect to the server
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Get entity info to start receiving broadcasts
	fmt.Printf("Getting entity info for %d...\n", entityID)
	info, err := client.GetEntityInfo(uint32(entityID))
	if err != nil {
		log.Fatalf("Failed to get entity info: %v", err)
	}
	fmt.Printf("Entity info: %+v\n", info)

	// Wait for Ctrl+C to exit
	fmt.Println("Listening for entity changes. Press Ctrl+C to exit.")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
