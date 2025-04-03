package rustplus

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chickenfresh/go-rustplus/rustplus/proto"
	"github.com/gorilla/websocket"
	protobuf "google.golang.org/protobuf/proto"
)

// Client represents a Rust+ client
type Client struct {
	// Configuration
	server            string
	port              int
	playerID          uint64
	playerToken       int
	useFacepunchProxy bool

	// WebSocket connection
	conn              *websocket.Conn
	connMutex         sync.RWMutex
	isConnected       bool
	reconnectAttempts int

	// Message handling
	seq              uint32
	seqCallbacks     map[uint32]func(*proto.AppMessage) bool
	seqCallbackMutex sync.Mutex
	messageHandlers  []func(*proto.AppMessage) bool
	handlerMutex     sync.RWMutex

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Event handling
	eventChan chan Event
}

// NewClient creates a new Rust+ client
func NewClient(server string, port int, playerID uint64, playerToken int, useFacepunchProxy bool) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		server:            server,
		port:              port,
		playerID:          playerID,
		playerToken:       playerToken,
		useFacepunchProxy: useFacepunchProxy,
		seqCallbacks:      make(map[uint32]func(*proto.AppMessage) bool),
		ctx:               ctx,
		cancel:            cancel,
		eventChan:         make(chan Event, 100),
	}
}

// Events returns a channel of client events
func (c *Client) Events() <-chan Event {
	return c.eventChan
}

// emitEvent emits an event to the event channel
func (c *Client) emitEvent(eventType EventType, data interface{}, err error) {
	select {
	case c.eventChan <- Event{Type: eventType, Data: data, Error: err}:
		// Event sent successfully
	default:
		// Channel is full, log the error
		fmt.Printf("Warning: Client event channel is full, event %s dropped\n", eventType)
	}
}

// Connect connects to the Rust+ server
func (c *Client) Connect() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if c.isConnected {
		return nil
	}

	c.emitEvent(EventConnecting, nil, nil)

	// Determine the WebSocket URL
	var wsURL string
	if c.useFacepunchProxy {
		wsURL = fmt.Sprintf("wss://companion-rust.facepunch.com/game/%s/%d", c.server, c.port)
	} else {
		wsURL = fmt.Sprintf("ws://%s:%d", c.server, c.port)
	}

	// Add URL parameters
	u, err := url.Parse(wsURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	q := u.Query()
	q.Set("playerid", fmt.Sprintf("%d", c.playerID))
	q.Set("playertoken", fmt.Sprintf("%d", c.playerToken))
	q.Set("protocol", "2")
	q.Set("app", "companion")
	u.RawQuery = q.Encode()

	// Connect to the WebSocket
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.isConnected = true
	c.reconnectAttempts = 0

	// Start the message reader
	go c.readMessages()

	c.emitEvent(EventConnected, nil, nil)
	return nil
}

// Disconnect disconnects from the Rust+ server
func (c *Client) Disconnect() error {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if !c.isConnected {
		return nil
	}

	// Close the WebSocket connection
	err := c.conn.Close()
	c.isConnected = false
	c.emitEvent(EventDisconnected, nil, nil)

	return err
}

// Close completely closes the client and releases resources
func (c *Client) Close() error {
	err := c.Disconnect()
	c.cancel()
	close(c.eventChan)
	return err
}

// readMessages reads messages from the WebSocket connection
func (c *Client) readMessages() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// Read the next message
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				c.handleDisconnect(err)
				return
			}

			// Parse the message
			appMessage := &proto.AppMessage{}
			if err := protobuf.Unmarshal(message, appMessage); err != nil {
				c.emitEvent(EventError, nil, fmt.Errorf("failed to unmarshal message: %w", err))
				continue
			}

			// Emit the message event
			c.emitEvent(EventMessage, appMessage, nil)

			// Handle the message
			c.handleMessage(appMessage)
		}
	}
}

// handleDisconnect handles a disconnection
func (c *Client) handleDisconnect(err error) {
	c.connMutex.Lock()
	wasConnected := c.isConnected
	c.isConnected = false
	c.connMutex.Unlock()

	if wasConnected {
		c.emitEvent(EventDisconnected, nil, err)

		// Try to reconnect
		c.reconnectAttempts++
		if c.reconnectAttempts < 5 {
			time.Sleep(time.Duration(c.reconnectAttempts) * time.Second)
			if err := c.Connect(); err != nil {
				c.emitEvent(EventError, nil, fmt.Errorf("failed to reconnect: %w", err))
			}
		}
	}
}

// handleMessage handles an incoming message
func (c *Client) handleMessage(msg *proto.AppMessage) {
	// Check if this is a response to a request
	if msg.Response != nil && msg.Response.Seq != nil {
		seq := msg.Response.GetSeq()

		c.seqCallbackMutex.Lock()
		callback, ok := c.seqCallbacks[seq]
		c.seqCallbackMutex.Unlock()

		if ok {
			if callback(msg) {
				c.seqCallbackMutex.Lock()
				delete(c.seqCallbacks, seq)
				c.seqCallbackMutex.Unlock()
			}
		}
	}

	// Run message handlers
	c.handlerMutex.RLock()
	for _, handler := range c.messageHandlers {
		if handler(msg) {
			break
		}
	}
	c.handlerMutex.RUnlock()
}

// AddMessageHandler adds a message handler
func (c *Client) AddMessageHandler(handler func(*proto.AppMessage) bool) {
	c.handlerMutex.Lock()
	c.messageHandlers = append(c.messageHandlers, handler)
	c.handlerMutex.Unlock()
}

// SendRequest sends a request to the server
func (c *Client) SendRequest(req *proto.AppRequest) error {
	// Increment the sequence number
	seq := atomic.AddUint32(&c.seq, 1)
	req.Seq = Uint32(seq)

	// Add player ID and token to the request
	req.PlayerId = Uint64(c.playerID)
	req.PlayerToken = Int32(int32(c.playerToken))

	// Marshal the message
	data, err := protobuf.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send the message
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()

	if !c.isConnected {
		return fmt.Errorf("not connected")
	}

	if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	c.emitEvent(EventRequest, req, nil)
	return nil
}

// SendRequestAsync sends a request to the server and waits for a response
func (c *Client) SendRequestAsync(req *proto.AppRequest, timeout time.Duration) (*proto.AppMessage, error) {
	// Create a channel for the response
	responseChan := make(chan *proto.AppMessage, 1)
	errChan := make(chan error, 1)

	// Increment the sequence number
	seq := atomic.AddUint32(&c.seq, 1)
	req.Seq = Uint32(seq)

	// Add player ID and token to the request
	req.PlayerId = Uint64(c.playerID)
	req.PlayerToken = Int32(int32(c.playerToken))

	// Register the callback
	c.seqCallbackMutex.Lock()
	c.seqCallbacks[seq] = func(msg *proto.AppMessage) bool {
		select {
		case responseChan <- msg:
		default:
		}
		return true
	}
	c.seqCallbackMutex.Unlock()

	// Send the request
	go func() {
		// Marshal the message
		data, err := protobuf.Marshal(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		// Send the message
		c.connMutex.RLock()
		defer c.connMutex.RUnlock()

		if !c.isConnected {
			errChan <- fmt.Errorf("not connected")
			return
		}

		if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			errChan <- fmt.Errorf("failed to send request: %w", err)
			return
		}

		c.emitEvent(EventRequest, req, nil)
	}()

	// Wait for the response or timeout
	select {
	case msg := <-responseChan:
		return msg, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(timeout):
		c.seqCallbackMutex.Lock()
		delete(c.seqCallbacks, seq)
		c.seqCallbackMutex.Unlock()
		return nil, fmt.Errorf("request timed out")
	}
}

// GetInfo gets information about the server
func (c *Client) GetInfo() (*proto.AppInfo, error) {
	request := &proto.AppRequest{
		GetInfo: &proto.AppEmpty{},
	}

	response, err := c.SendRequestAsync(request, 10*time.Second)
	if err != nil {
		return nil, err
	}

	if response.Response == nil || response.Response.Info == nil {
		return nil, fmt.Errorf("invalid response")
	}

	return response.Response.GetInfo(), nil
}

// GetTime gets the current time on the server
func (c *Client) GetTime() (*proto.AppTime, error) {
	request := &proto.AppRequest{
		GetTime: &proto.AppEmpty{},
	}

	response, err := c.SendRequestAsync(request, 10*time.Second)
	if err != nil {
		return nil, err
	}

	if response.Response == nil || response.Response.Time == nil {
		return nil, fmt.Errorf("invalid response")
	}

	return response.Response.GetTime(), nil
}

// GetMap gets the map from the server
func (c *Client) GetMap() (*proto.AppMap, error) {
	request := &proto.AppRequest{
		GetMap: &proto.AppEmpty{},
	}

	response, err := c.SendRequestAsync(request, 10*time.Second)
	if err != nil {
		return nil, err
	}

	if response.Response == nil || response.Response.Map == nil {
		return nil, fmt.Errorf("invalid response")
	}

	return response.Response.GetMap(), nil
}

// GetTeamInfo gets information about the player's team
func (c *Client) GetTeamInfo() (*proto.AppTeamInfo, error) {
	request := &proto.AppRequest{
		GetTeamInfo: &proto.AppEmpty{},
	}

	response, err := c.SendRequestAsync(request, 10*time.Second)
	if err != nil {
		return nil, err
	}

	if response.Response == nil || response.Response.TeamInfo == nil {
		return nil, fmt.Errorf("invalid response")
	}

	return response.Response.GetTeamInfo(), nil
}

// GetTeamChat gets the team chat
func (c *Client) GetTeamChat() ([]*proto.AppTeamMessage, error) {
	request := &proto.AppRequest{
		GetTeamChat: &proto.AppEmpty{},
	}

	response, err := c.SendRequestAsync(request, 10*time.Second)
	if err != nil {
		return nil, err
	}

	if response.Response == nil || response.Response.TeamChat == nil {
		return nil, fmt.Errorf("invalid response")
	}

	return response.Response.GetTeamChat().GetMessages(), nil
}

// SendTeamMessage sends a message to the team chat
func (c *Client) SendTeamMessage(message string) error {
	request := &proto.AppRequest{
		SendTeamMessage: &proto.AppSendMessage{
			Message: protobuf.String(message),
		},
	}

	_, err := c.SendRequestAsync(request, 10*time.Second)
	return err
}

// GetEntityInfo gets information about an entity
func (c *Client) GetEntityInfo(entityID uint32) (*proto.AppEntityInfo, error) {
	request := &proto.AppRequest{
		EntityId: protobuf.Uint32(entityID),
	}

	response, err := c.SendRequestAsync(request, 10*time.Second)
	if err != nil {
		return nil, err
	}

	if response.Response == nil || response.Response.EntityInfo == nil {
		return nil, fmt.Errorf("invalid response")
	}

	return response.Response.GetEntityInfo(), nil
}

// SetEntityValue sets a value on an entity
func (c *Client) SetEntityValue(entityID uint32, value bool) error {
	request := &proto.AppRequest{
		EntityId: protobuf.Uint32(entityID),
		SetEntityValue: &proto.AppSetEntityValue{
			Value: protobuf.Bool(value),
		},
	}

	_, err := c.SendRequestAsync(request, 10*time.Second)
	return err
}

// GetCamera gets a camera instance for controlling CCTV cameras
func (c *Client) GetCamera(identifier string) *Camera {
	return NewCamera(c, identifier)
}
