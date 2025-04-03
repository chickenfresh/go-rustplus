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
		fmt.Println("Usage: send_team_chat <server_ip> <server_port> <player_id> <player_token>")
		os.Exit(1)
	}

	serverIP := os.Args[1]
	serverPort, _ := strconv.Atoi(os.Args[2])
	playerID, _ := strconv.ParseUint(os.Args[3], 10, 64)
	playerToken, _ := strconv.Atoi(os.Args[4])

	// Create a new client
	client := rustplus.NewClient(serverIP, serverPort, playerID, playerToken, false)

	// Connect to the server
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Send message to team chat
	fmt.Println("Sending message to team chat...")
	if err := client.SendTeamMessage("Hello from go-rustplus!"); err != nil {
		log.Fatalf("Failed to send team message: %v", err)
	}

	fmt.Println("Message sent successfully!")
}
