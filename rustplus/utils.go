package rustplus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ParsePairingNotification parses a pairing notification from FCM
func ParsePairingNotification(notification map[string]interface{}) (string, int, uint64, int, error) {
	// Extract the data from the notification
	data, ok := notification["data"].(map[string]interface{})
	if !ok {
		return "", 0, 0, 0, fmt.Errorf("invalid notification format: missing data")
	}

	// Extract the body from the data
	bodyStr, ok := data["body"].(string)
	if !ok {
		return "", 0, 0, 0, fmt.Errorf("invalid notification format: missing body")
	}

	// Parse the body as JSON
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(bodyStr), &body); err != nil {
		return "", 0, 0, 0, fmt.Errorf("invalid notification format: body is not valid JSON: %w", err)
	}

	// Extract the server information
	serverStr, ok := body["server"].(string)
	if !ok {
		return "", 0, 0, 0, fmt.Errorf("invalid notification format: missing server")
	}

	// Extract the player token
	var playerToken int
	playerTokenVal, ok := body["playerToken"]
	if ok {
		switch v := playerTokenVal.(type) {
		case float64:
			playerToken = int(v)
		case string:
			if _, err := fmt.Sscanf(v, "%d", &playerToken); err != nil {
				return "", 0, 0, 0, fmt.Errorf("invalid notification format: playerToken is not a valid number: %w", err)
			}
		default:
			return "", 0, 0, 0, fmt.Errorf("invalid notification format: playerToken has unexpected type")
		}
	} else {
		return "", 0, 0, 0, fmt.Errorf("invalid notification format: missing playerToken")
	}

	// Extract the player ID
	playerIDStr, ok := body["playerId"].(string)
	if !ok {
		return "", 0, 0, 0, fmt.Errorf("invalid notification format: missing playerId")
	}
	var playerID uint64
	if _, err := fmt.Sscanf(playerIDStr, "%d", &playerID); err != nil {
		return "", 0, 0, 0, fmt.Errorf("invalid notification format: playerId is not a valid number: %w", err)
	}

	// Parse the server string to extract the IP and port
	parts := strings.Split(serverStr, ":")
	if len(parts) != 2 {
		return "", 0, 0, 0, fmt.Errorf("invalid server format: %s", serverStr)
	}

	server := parts[0]
	var port int
	if _, err := fmt.Sscanf(parts[1], "%d", &port); err != nil {
		return "", 0, 0, 0, fmt.Errorf("invalid server format: port is not a valid number: %w", err)
	}

	return server, port, playerID, playerToken, nil
}

// GetServerInfo gets information about a Rust server from the Steam API
func GetServerInfo(ip string, port int) (map[string]interface{}, error) {
	// Create the request URL
	url := fmt.Sprintf("https://api.steampowered.com/IGameServersService/GetServerList/v1/?filter=\\addr\\%s:%d&key=YOUR_STEAM_API_KEY", ip, port)

	// Send the request
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract the server information
	responseData, ok := response["response"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: missing response")
	}

	servers, ok := responseData["servers"].([]interface{})
	if !ok || len(servers) == 0 {
		return nil, fmt.Errorf("invalid response format: missing servers")
	}

	server, ok := servers[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: server is not an object")
	}

	return server, nil
}
