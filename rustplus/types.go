package rustplus

// CameraButtons represents the possible buttons that can be sent to the server
type CameraButtons int32

// Camera button constants
const (
	ButtonNone          CameraButtons = 0
	ButtonForward       CameraButtons = 2
	ButtonBackward      CameraButtons = 4
	ButtonLeft          CameraButtons = 8
	ButtonRight         CameraButtons = 16
	ButtonJump          CameraButtons = 32
	ButtonDuck          CameraButtons = 64
	ButtonSprint        CameraButtons = 128
	ButtonUse           CameraButtons = 256
	ButtonFirePrimary   CameraButtons = 1024
	ButtonFireSecondary CameraButtons = 2048
	ButtonReload        CameraButtons = 8192
	ButtonFireThird     CameraButtons = 134217728
)

// CameraControlFlags represents the possible control flags for cameras
type CameraControlFlags int32

// Camera control flag constants
const (
	ControlNone          CameraControlFlags = 0
	ControlMovement      CameraControlFlags = 1
	ControlMouse         CameraControlFlags = 2
	ControlSprintAndDuck CameraControlFlags = 4
	ControlFire          CameraControlFlags = 8
	ControlReload        CameraControlFlags = 16
	ControlCrosshair     CameraControlFlags = 32
)

// EventType represents the type of event emitted by the client
type EventType string

// Event types
const (
	EventConnecting   EventType = "connecting"
	EventConnected    EventType = "connected"
	EventDisconnected EventType = "disconnected"
	EventMessage      EventType = "message"
	EventRequest      EventType = "request"
	EventError        EventType = "error"
)

// Event represents an event emitted by the client
type Event struct {
	Type  EventType
	Data  interface{}
	Error error
}

// Helper functions to create protocol buffer values

// Uint32 creates a uint32 pointer for protocol buffer fields
func Uint32(v uint32) *uint32 {
	return &v
}

// Int32 creates an int32 pointer for protocol buffer fields
func Int32(v int32) *int32 {
	return &v
}

// Uint64 creates a uint64 pointer for protocol buffer fields
func Uint64(v uint64) *uint64 {
	return &v
}

// Float32 creates a float32 pointer for protocol buffer fields
func Float32(v float32) *float32 {
	return &v
}

// String creates a string pointer for protocol buffer fields
func String(v string) *string {
	return &v
}

// Bool creates a bool pointer for protocol buffer fields
func Bool(v bool) *bool {
	return &v
}
