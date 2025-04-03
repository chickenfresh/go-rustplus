package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

//go:embed pair.html
var pairHTML embed.FS

// linkSteamWithRustPlus opens a web server to handle Steam authentication
func linkSteamWithRustPlus() (string, error) {
	// Create a channel to receive the auth token
	tokenChan := make(chan string, 1)

	// Create a channel to signal server shutdown
	shutdownChan := make(chan struct{})

	// Use a mutex and flag to track if the channel has been closed
	var shutdownMutex sync.Mutex
	channelClosed := false

	// Safe channel close function
	safeCloseChannel := func() {
		shutdownMutex.Lock()
		defer shutdownMutex.Unlock()

		if !channelClosed {
			close(shutdownChan)
			channelClosed = true
		}
	}

	// Create HTTP server
	mux := http.NewServeMux()

	// Serve the pair.html page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := pairHTML.ReadFile("pair.html")
		if err != nil {
			http.Error(w, "Failed to read pair.html", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})

	// Handle the callback from the Steam authentication
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		var token string

		// Check if it's a GET or POST request
		if r.Method == "POST" {
			// For POST requests (from axios)

			// Try to parse JSON body
			var requestData struct {
				Token string `json:"token"`
			}

			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&requestData); err == nil && requestData.Token != "" {
				token = requestData.Token
			} else {
				// If JSON parsing fails, try form data
				if err := r.ParseForm(); err != nil {
					http.Error(w, "Failed to parse form data", http.StatusBadRequest)
					return
				}
				token = r.FormValue("token")
			}

			// Set CORS headers for POST requests
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		} else if r.Method == "OPTIONS" {
			// Handle preflight requests
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			return
		} else {
			// For GET requests (from redirects)
			token = r.URL.Query().Get("token")
		}

		if token == "" {
			http.Error(w, "Token missing from request!", http.StatusBadRequest)
			return
		}

		fmt.Printf("Received token: %s\n", token)

		// For POST requests, just send a simple success response
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success": true}`))
		} else {
			// For GET requests, send a simple success message
			w.Write([]byte("Success! You can now close this window."))
		}

		// Send the token to the channel
		select {
		case tokenChan <- token:
			// Signal server to shutdown
			safeCloseChannel()
		default:
			fmt.Println("Warning: Token channel is full, this is unexpected")
		}
	})

	// Create server
	server := &http.Server{
		Addr:    ":3000",
		Handler: mux,
	}

	// Start the server in a goroutine
	go func() {
		fmt.Println("Starting web server on http://localhost:3000")
		fmt.Println("Please open this URL in your browser to link your Steam account with Rust+")

		// Open the browser
		if err := openBrowser("http://localhost:3000"); err != nil {
			fmt.Printf("Failed to open browser: %v\n", err)
			fmt.Println("Please manually open http://localhost:3000 in your browser")
		}

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Wait for token or timeout
	select {
	case token := <-tokenChan:
		// Wait a moment to ensure the browser receives the success message
		time.Sleep(500 * time.Millisecond)

		// Shutdown the server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			fmt.Printf("Server shutdown error: %v\n", err)
		}

		return token, nil

	case <-time.After(5 * time.Minute):
		// Timeout after 5 minutes
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			fmt.Printf("Server shutdown error: %v\n", err)
		}

		// Safely close the channel
		safeCloseChannel()

		return "", fmt.Errorf("timed out waiting for Steam authentication")
	}
}
