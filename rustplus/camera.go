package rustplus

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"

	"github.com/chickenfresh/go-rustplus/rustplus/proto"
	protobuf "google.golang.org/protobuf/proto"
)

// Camera represents a CCTV camera in Rust
type Camera struct {
	client         *Client
	identifier     string
	isSubscribed   bool
	subscribeInfo  *proto.AppCameraInfo
	cameraRays     []*proto.AppCameraRays
	subscribeTimer *time.Timer
	mutex          sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	eventChan      chan CameraEvent
}

// CameraEventType represents the type of camera event
type CameraEventType string

const (
	// CameraEventSubscribing is emitted when the camera is subscribing
	CameraEventSubscribing CameraEventType = "subscribing"
	// CameraEventSubscribed is emitted when the camera is subscribed
	CameraEventSubscribed CameraEventType = "subscribed"
	// CameraEventUnsubscribing is emitted when the camera is unsubscribing
	CameraEventUnsubscribing CameraEventType = "unsubscribing"
	// CameraEventUnsubscribed is emitted when the camera is unsubscribed
	CameraEventUnsubscribed CameraEventType = "unsubscribed"
	// CameraEventRender is emitted when a camera frame is rendered
	CameraEventRender CameraEventType = "render"
	// CameraEventError is emitted when an error occurs
	CameraEventError CameraEventType = "error"
)

// CameraEvent represents an event emitted by the camera
type CameraEvent struct {
	Type  CameraEventType
	Data  interface{}
	Error error
}

// NewCamera creates a new camera instance
func NewCamera(client *Client, identifier string) *Camera {
	ctx, cancel := context.WithCancel(context.Background())

	return &Camera{
		client:     client,
		identifier: identifier,
		ctx:        ctx,
		cancel:     cancel,
		eventChan:  make(chan CameraEvent, 100),
	}
}

// Events returns a channel of camera events
func (c *Camera) Events() <-chan CameraEvent {
	return c.eventChan
}

// emitEvent emits an event to the event channel
func (c *Camera) emitEvent(eventType CameraEventType, data interface{}, err error) {
	select {
	case c.eventChan <- CameraEvent{Type: eventType, Data: data, Error: err}:
		// Event sent successfully
	default:
		// Channel is full, log the error
		fmt.Printf("Warning: Camera event channel is full, event %s dropped\n", eventType)
	}
}

// Subscribe subscribes to the camera
func (c *Camera) Subscribe() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isSubscribed {
		return errors.New("already subscribed to camera")
	}

	c.emitEvent(CameraEventSubscribing, nil, nil)

	// Create the subscribe request
	request := &proto.AppRequest{
		CameraSubscribe: &proto.AppCameraSubscribe{
			CameraId: &c.identifier,
		},
	}

	// Send the request
	response, err := c.client.SendRequestAsync(request, 10*time.Second)
	if err != nil {
		c.emitEvent(CameraEventError, nil, fmt.Errorf("failed to subscribe to camera: %w", err))
		return err
	}

	// Check if we got a valid response
	if response.Response == nil || response.Response.CameraSubscribeInfo == nil {
		err := errors.New("invalid camera subscribe response")
		c.emitEvent(CameraEventError, nil, err)
		return err
	}

	// Store the camera info
	c.subscribeInfo = response.Response.CameraSubscribeInfo

	// Set up a handler for camera rays
	c.client.AddMessageHandler(func(msg *proto.AppMessage) bool {
		if msg.Broadcast != nil && msg.Broadcast.CameraRays != nil {
			c.handleCameraRays(msg.Broadcast.CameraRays)
			return true
		}
		return false
	})

	// Start the subscription timer
	c.startSubscriptionTimer()

	c.isSubscribed = true
	c.emitEvent(CameraEventSubscribed, c.subscribeInfo, nil)

	return nil
}

// Unsubscribe unsubscribes from the camera
func (c *Camera) Unsubscribe() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isSubscribed {
		return errors.New("not subscribed to camera")
	}

	c.emitEvent(CameraEventUnsubscribing, nil, nil)

	// Stop the subscription timer
	if c.subscribeTimer != nil {
		c.subscribeTimer.Stop()
		c.subscribeTimer = nil
	}

	// Create the unsubscribe request
	request := &proto.AppRequest{
		CameraUnsubscribe: &proto.AppEmpty{},
	}

	// Send the request
	_, err := c.client.SendRequestAsync(request, 10*time.Second)
	if err != nil {
		c.emitEvent(CameraEventError, nil, fmt.Errorf("failed to unsubscribe from camera: %w", err))
		return err
	}

	c.isSubscribed = false
	c.emitEvent(CameraEventUnsubscribed, nil, nil)

	return nil
}

// Move sends camera movement input to the server
func (c *Camera) Move(buttons CameraButtons, x, y float32) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.isSubscribed {
		return errors.New("not subscribed to camera")
	}

	// Create the camera input request
	request := &proto.AppRequest{
		CameraInput: &proto.AppCameraInput{
			Buttons: protobuf.Int32(int32(buttons)),
			MouseDelta: &proto.Vector2{
				X: protobuf.Float32(x),
				Y: protobuf.Float32(y),
			},
		},
	}

	// Send the request
	_, err := c.client.SendRequestAsync(request, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to send camera input: %w", err)
	}

	return nil
}

// LookUp looks up
func (c *Camera) LookUp() error {
	return c.Move(ButtonNone, 0, -10)
}

// LookDown looks down
func (c *Camera) LookDown() error {
	return c.Move(ButtonNone, 0, 10)
}

// LookLeft looks left
func (c *Camera) LookLeft() error {
	return c.Move(ButtonNone, -10, 0)
}

// LookRight looks right
func (c *Camera) LookRight() error {
	return c.Move(ButtonNone, 10, 0)
}

// MoveForward moves forward
func (c *Camera) MoveForward() error {
	return c.Move(ButtonForward, 0, 0)
}

// MoveBackward moves backward
func (c *Camera) MoveBackward() error {
	return c.Move(ButtonBackward, 0, 0)
}

// MoveLeft strafes left
func (c *Camera) MoveLeft() error {
	return c.Move(ButtonLeft, 0, 0)
}

// MoveRight strafes right
func (c *Camera) MoveRight() error {
	return c.Move(ButtonRight, 0, 0)
}

// Shoot shoots a PTZ controllable Auto Turret
func (c *Camera) Shoot() error {
	// Press left mouse button to shoot
	if err := c.Move(ButtonFirePrimary, 0, 0); err != nil {
		return err
	}

	// Release all mouse buttons
	return c.Move(ButtonNone, 0, 0)
}

// Reload reloads a PTZ controllable Auto Turret
func (c *Camera) Reload() error {
	// Press reload button
	if err := c.Move(ButtonReload, 0, 0); err != nil {
		return err
	}

	// Release all buttons
	return c.Move(ButtonNone, 0, 0)
}

// IsAutoTurret checks if the camera is an auto turret
func (c *Camera) IsAutoTurret() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.subscribeInfo == nil {
		return false
	}

	return (CameraControlFlags(*c.subscribeInfo.ControlFlags) & ControlCrosshair) == ControlCrosshair
}

// startSubscriptionTimer starts a timer to periodically resubscribe to the camera
func (c *Camera) startSubscriptionTimer() {
	// Cancel any existing timer
	if c.subscribeTimer != nil {
		c.subscribeTimer.Stop()
	}

	// Create a new timer that resubscribes every 30 seconds
	c.subscribeTimer = time.AfterFunc(30*time.Second, func() {
		// Resubscribe
		request := &proto.AppRequest{
			CameraSubscribe: &proto.AppCameraSubscribe{
				CameraId: &c.identifier,
			},
		}

		_, err := c.client.SendRequestAsync(request, 10*time.Second)
		if err != nil {
			c.emitEvent(CameraEventError, nil, fmt.Errorf("failed to resubscribe to camera: %w", err))
			return
		}

		// Restart the timer
		c.startSubscriptionTimer()
	})
}

// handleCameraRays processes camera rays data
func (c *Camera) handleCameraRays(rays *proto.AppCameraRays) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Add the rays to our collection
	c.cameraRays = append(c.cameraRays, rays)

	// If we have enough rays, render the image
	if len(c.cameraRays) >= 10 {
		// Render the image
		img, err := c.renderImage()
		if err != nil {
			c.emitEvent(CameraEventError, nil, fmt.Errorf("failed to render image: %w", err))
			return
		}

		// Emit the render event with the image
		c.emitEvent(CameraEventRender, img, nil)

		// Clear the rays
		c.cameraRays = nil
	}
}

// renderImage renders an image from the camera rays
func (c *Camera) renderImage() (image.Image, error) {
	if c.subscribeInfo == nil || len(c.cameraRays) == 0 {
		return nil, errors.New("no camera data available")
	}

	width := int(*c.subscribeInfo.Width)
	height := int(*c.subscribeInfo.Height)

	// Create a new RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with black
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}

	// Process each ray packet
	for _, rays := range c.cameraRays {
		// Process ray data
		rayData := rays.GetRayData()
		sampleOffset := int(rays.GetSampleOffset())

		// Create an index generator with the sample offset
		indexGen := newIndexGenerator(uint32(sampleOffset))

		// Calculate the number of rays
		rayCount := len(rayData) / 4 // Each ray is 4 bytes

		// Process each ray
		for i := 0; i < rayCount; i++ {
			// Get a random position based on the index generator
			x := indexGen.nextInt(width)
			y := indexGen.nextInt(height)

			// Get the ray data
			rayIndex := i * 4
			if rayIndex+3 >= len(rayData) {
				break
			}

			// Extract the ray data
			r := rayData[rayIndex]
			g := rayData[rayIndex+1]
			b := rayData[rayIndex+2]
			a := rayData[rayIndex+3]

			// Set the pixel
			img.Set(x, y, color.RGBA{r, g, b, a})
		}

		// Process entities
		for _, entity := range rays.GetEntities() {
			// Project 3D entity position to 2D screen coordinates
			// This is a simplified projection that doesn't account for camera rotation
			// For a more accurate projection, we would need the camera's view matrix

			// Get entity position
			pos := entity.GetPosition()

			// Get entity size for drawing
			size := entity.GetSize()
			entitySize := int(size.GetX()+size.GetY()+size.GetZ()) / 3
			if entitySize < 5 {
				entitySize = 5 // Minimum size
			} else if entitySize > 20 {
				entitySize = 20 // Maximum size
			}

			// Simple projection - this is approximate and won't be perfect
			// We're assuming the camera is at origin looking in the negative Z direction
			// A proper projection would use the camera's view and projection matrices
			// fov := 60.0 // Approximate field of view in degrees
			// aspectRatio := float64(width) / float64(height)

			// Convert to screen space
			depth := -pos.GetZ() // Depth is negative Z in view space
			if depth <= 0 {
				continue // Behind the camera
			}

			// Calculate screen position
			screenX := int((pos.GetX()/depth)*float32(width)*0.5 + float32(width)*0.5)
			screenY := int((-pos.GetY()/depth)*float32(height)*0.5 + float32(height)*0.5)

			// Check if on screen
			if screenX < 0 || screenX >= width || screenY < 0 || screenY >= height {
				continue
			}

			// Draw the entity as a colored circle
			var entityColor color.RGBA

			switch entity.GetType() {
			case proto.AppCameraRays_Player:
				// Red for players
				entityColor = color.RGBA{255, 0, 0, 255}
			case proto.AppCameraRays_Tree:
				// Green for trees
				entityColor = color.RGBA{0, 255, 0, 255}
			default:
				// Yellow for other entities
				entityColor = color.RGBA{255, 255, 0, 255}
			}

			// Draw a simple circle for the entity
			drawCircle(img, screenX, screenY, entitySize, entityColor)

			// If it's a player, draw their name
			if entity.GetType() == proto.AppCameraRays_Player && entity.Name != nil {
				// TODO: Draw player name
				// This would require a font rendering library
				// For simplicity, we'll skip this part
			}
		}
	}

	return img, nil
}

// IndexGenerator is a deterministic random number generator
// This is a direct port of the JavaScript implementation
type indexGenerator struct {
	state uint32
}

// newIndexGenerator creates a new index generator
func newIndexGenerator(seed uint32) *indexGenerator {
	gen := &indexGenerator{state: seed}
	gen.nextState()
	return gen
}

// nextInt generates a random integer in the range [0, max)
func (g *indexGenerator) nextInt(max int) int {
	state := g.nextState()
	result := int((uint64(state) * uint64(max)) / 4294967295)
	if result < 0 {
		result = max + result - 1
	}
	return result
}

// nextState advances the generator state
func (g *indexGenerator) nextState() uint32 {
	e := g.state
	t := e

	e = e ^ (e << 13)
	e = e ^ (e >> 17)
	e = e ^ (e << 5)

	g.state = e

	if t >= 0 {
		return t
	}
	return 4294967295 + t - 1
}

// drawCircle draws a circle on the image
func drawCircle(img *image.RGBA, x, y, radius int, c color.RGBA) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			// Check if the point is within the circle
			if dx*dx+dy*dy <= radius*radius {
				// Set the pixel if it's within the image bounds
				px, py := x+dx, y+dy
				if px >= 0 && px < img.Bounds().Max.X && py >= 0 && py < img.Bounds().Max.Y {
					img.Set(px, py, c)
				}
			}
		}
	}
}

// Zoom zooms the camera by simulating a click of the primary fire button
func (c *Camera) Zoom() error {
	// Press fire button
	if err := c.Move(ButtonFirePrimary, 0, 0); err != nil {
		return err
	}

	// Release all buttons
	return c.Move(ButtonNone, 0, 0)
}
