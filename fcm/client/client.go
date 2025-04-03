package client

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chickenfresh/go-rustplus/fcm/constants"
	"github.com/chickenfresh/go-rustplus/fcm/crypto"
	"github.com/chickenfresh/go-rustplus/fcm/gcm"
	"github.com/chickenfresh/go-rustplus/fcm/parser"
	fcmproto "github.com/chickenfresh/go-rustplus/fcm/proto"
	"google.golang.org/protobuf/proto"
)

const (
	host            = "mtalk.google.com"
	port            = 5228
	maxRetryTimeout = 15 // seconds
)

// Notification represents a received FCM notification
type Notification struct {
	Message      map[string]interface{} `json:"notification"`
	PersistentID string                 `json:"persistentId"`
	Object       *fcmproto.DataMessageStanza
}

// Client represents an FCM client
type Client struct {
	androidID      string
	securityToken  string
	persistentIDs  []string
	retryCount     int
	conn           net.Conn
	parser         *parser.Parser
	retryTimer     *time.Timer
	credentials    *crypto.Keys
	mutex          sync.Mutex
	onConnect      func()
	onDisconnect   func()
	onNotification func(Notification)
	stopChan       chan struct{}
	onError        func(error)
}

// NewClient creates a new FCM client
func NewClient(androidID, securityToken string, persistentIDs []string) *Client {
	return &Client{
		androidID:     androidID,
		securityToken: securityToken,
		persistentIDs: persistentIDs,
		stopChan:      make(chan struct{}),
	}
}

// SetCredentials sets the crypto credentials for decrypting messages
func (c *Client) SetCredentials(privateKey, authSecret string) {
	c.credentials = &crypto.Keys{
		PrivateKey: privateKey,
		AuthSecret: authSecret,
	}
}

// OnConnect sets the callback for connection events
func (c *Client) OnConnect(callback func()) {
	c.onConnect = callback
}

// OnDisconnect sets the callback for disconnection events
func (c *Client) OnDisconnect(callback func()) {
	c.onDisconnect = callback
}

// OnNotification sets the callback for notification events
func (c *Client) OnNotification(callback func(Notification)) {
	c.onNotification = callback
}

// OnError sets the callback for error events
func (c *Client) OnError(callback func(error)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.onError = callback
}

// Connect establishes a connection to FCM
func (c *Client) Connect() error {
	// First check in with GCM
	if err := c.checkIn(); err != nil {
		return fmt.Errorf("check-in failed: %w", err)
	}

	// Then connect to FCM
	if err := c.connect(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	return nil
}

// Stop closes the connection and stops the client
func (c *Client) Stop() {
	c.destroy()
	close(c.stopChan)
}

// checkIn performs a check-in with GCM
func (c *Client) checkIn() error {
	_, err := gcm.CheckIn(c.androidID, c.securityToken)
	return err
}

// connect establishes a TLS connection to FCM
func (c *Client) connect() error {
	// Create TLS connection
	dialer := &net.Dialer{
		Timeout: time.Duration(constants.ConnectionTimeout) * time.Second,
	}
	conn, err := tls.DialWithDialer(dialer, "tcp", fmt.Sprintf("%s:%d", host, port), &tls.Config{
		InsecureSkipVerify: false,
	})
	if err != nil {
		return fmt.Errorf("failed to establish TLS connection: %w", err)
	}

	c.conn = conn

	// Send login request
	loginBuffer, err := c.createLoginBuffer()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to create login buffer: %w", err)
	}

	if _, err := c.conn.Write(loginBuffer); err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to send login request: %w", err)
	}

	// Create parser with Config struct
	c.parser = parser.NewParser(parser.Config{
		Reader:         c.conn,
		MessageHandler: c.onMessage,
		ErrorHandler:   c.onParserError,
		Debug:          false,
	})
	c.parser.Start()

	// Reset retry count
	c.retryCount = 0

	// Trigger connect callback
	if c.onConnect != nil {
		c.onConnect()
	}

	return nil
}

// destroy cleans up resources
func (c *Client) destroy() {
	if c.retryTimer != nil {
		c.retryTimer.Stop()
		c.retryTimer = nil
	}

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	if c.parser != nil {
		c.parser.Stop()
		c.parser = nil
	}
}

// retry attempts to reconnect after a delay
func (c *Client) retry() {
	c.destroy()

	// Calculate retry timeout
	c.retryCount++
	if c.retryCount > maxRetryTimeout {
		c.retryCount = maxRetryTimeout
	}
	timeout := time.Duration(c.retryCount) * time.Second

	// Schedule retry
	c.retryTimer = time.AfterFunc(timeout, func() {
		if err := c.Connect(); err != nil {
			log.Printf("Retry failed: %v", err)
			c.retry()
		}
	})
}

// onParserError handles parser errors
func (c *Client) onParserError(err error) {
	log.Printf("Parser error: %v", err)
	c.retry()
}

// onMessage handles incoming messages
func (c *Client) onMessage(tag int, message proto.Message) {
	switch tag {
	case constants.LoginResponseTag:
		// Clear persistent IDs as we just sent them to the server
		c.mutex.Lock()
		c.persistentIDs = []string{}
		c.mutex.Unlock()

	case constants.DataMessageStanzaTag:
		// Process data message
		msg, ok := message.(*fcmproto.DataMessageStanza)
		if !ok {
			log.Printf("Invalid message type for DataMessageStanza")
			return
		}
		c.processDataMessage(msg)
	}
}

// processDataMessage processes a data message stanza
func (c *Client) processDataMessage(msg *fcmproto.DataMessageStanza) {
	// Check if we've already seen this message
	c.mutex.Lock()
	for _, id := range c.persistentIDs {
		if id == *msg.PersistentId {
			c.mutex.Unlock()
			return
		}
	}
	c.mutex.Unlock()

	// Check if the message has crypto keys
	hasCryptoKey := false
	for _, appData := range msg.AppData {
		if appData.Key == proto.String("crypto-key") {
			hasCryptoKey = true
			break
		}
	}

	// If no crypto keys, just emit the raw message
	if !hasCryptoKey {
		c.mutex.Lock()
		c.persistentIDs = append(c.persistentIDs, *msg.PersistentId)
		c.mutex.Unlock()

		// Emit raw data message
		if c.onNotification != nil {
			c.onNotification(Notification{
				Message:      map[string]interface{}{"raw": true},
				PersistentID: *msg.PersistentId,
				Object:       msg,
			})
		}
		return
	}

	// If we have credentials, try to decrypt
	if c.credentials != nil {
		// Convert to format expected by decrypt function
		encryptedMsg := crypto.EncryptedMessage{
			AppData: make([]crypto.AppDataItem, 0, len(msg.AppData)),
			RawData: msg.RawData,
		}

		for _, appData := range msg.AppData {
			encryptedMsg.AppData = append(encryptedMsg.AppData, crypto.AppDataItem{
				Key:   *appData.Key,
				Value: *appData.Value,
			})
		}

		// Try to decrypt
		decrypted, err := crypto.DecryptMessage(encryptedMsg, *c.credentials)
		if err != nil {
			// Handle specific errors silently
			if isIgnorableDecryptError(err) {
				log.Printf("Message dropped as it could not be decrypted: %v", err)
				c.mutex.Lock()
				c.persistentIDs = append(c.persistentIDs, *msg.PersistentId)
				c.mutex.Unlock()
				return
			}
			log.Printf("Failed to decrypt message: %v", err)
		} else {
			// Successfully decrypted
			c.mutex.Lock()
			c.persistentIDs = append(c.persistentIDs, *msg.PersistentId)
			c.mutex.Unlock()

			// Emit notification
			if c.onNotification != nil {
				c.onNotification(Notification{
					Message:      decrypted,
					PersistentID: *msg.PersistentId,
					Object:       msg,
				})
			}
			return
		}
	}

	// If we get here, either we don't have credentials or decryption failed
	// Add to persistent IDs and emit raw message
	c.mutex.Lock()
	c.persistentIDs = append(c.persistentIDs, *msg.PersistentId)
	c.mutex.Unlock()

	if c.onNotification != nil {
		c.onNotification(Notification{
			Message:      map[string]interface{}{"raw": true, "decryptFailed": true},
			PersistentID: *msg.PersistentId,
			Object:       msg,
		})
	}
}

// createLoginBuffer creates the login request buffer
func (c *Client) createLoginBuffer() ([]byte, error) {
	// Convert Android ID to hex
	androidID, err := strconv.ParseInt(c.androidID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid Android ID: %w", err)
	}
	hexAndroidID := fmt.Sprintf("%x", androidID)

	// Create login request
	loginRequest := &fcmproto.LoginRequest{
		AdaptiveHeartbeat: proto.Bool(false),
		AuthService:       fcmproto.LoginRequest_ANDROID_ID.Enum(),
		AuthToken:         proto.String(c.securityToken),
		Id:                proto.String("chrome-63.0.3234.0"),
		Domain:            proto.String("mcs.android.com"),
		DeviceId:          proto.String(fmt.Sprintf("android-%s", hexAndroidID)),
		NetworkType:       proto.Int32(1),
		Resource:          proto.String(c.androidID),
		User:              proto.String(c.androidID),
		UseRmq2:           proto.Bool(true),
		Setting: []*fcmproto.Setting{
			{Name: proto.String("new_vc"), Value: proto.String("1")},
		},
		ClientEvent:          []*fcmproto.ClientEvent{},
		ReceivedPersistentId: c.persistentIDs,
	}

	// Encode login request
	data, err := proto.Marshal(loginRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	// Add varint length prefix
	var lenBuf [10]byte
	n := binary.PutUvarint(lenBuf[:], uint64(len(data)))

	// Combine version, tag, length, and data
	result := make([]byte, 2+n+len(data))
	result[0] = byte(constants.MCSVersion)
	result[1] = byte(constants.LoginRequestTag)
	copy(result[2:], lenBuf[:n])
	copy(result[2+n:], data)

	return result, nil
}

// isIgnorableDecryptError checks if a decryption error can be ignored
func isIgnorableDecryptError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "Unsupported state or unable to authenticate data") ||
		strings.Contains(errMsg, "crypto-key is missing") ||
		strings.Contains(errMsg, "salt is missing")
}
