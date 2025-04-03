package parser

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/chickenfresh/go-rustplus/fcm/constants"
	fcmproto "github.com/chickenfresh/go-rustplus/fcm/proto"
	"google.golang.org/protobuf/proto"
)

// MockReader implements the Reader interface for testing
type MockReader struct {
	data  [][]byte
	index int
	err   error
}

func (m *MockReader) Read(p []byte) (int, error) {
	if m.index >= len(m.data) {
		return 0, m.err
	}

	n := copy(p, m.data[m.index])
	m.index++
	return n, nil
}

// TestParser_LoginResponse tests parsing a login response message
func TestParser_LoginResponse(t *testing.T) {
	// Create test data for a login response message
	loginResponse := &fcmproto.LoginResponse{
		Id: proto.String("test-id"),
	}

	// Marshal the login response
	data, err := proto.Marshal(loginResponse)
	if err != nil {
		t.Fatalf("Failed to marshal login response: %v", err)
	}

	// Create the full message with version, tag, and size
	message := []byte{
		byte(constants.MCSVersion),       // Version
		byte(constants.LoginResponseTag), // Tag
		byte(len(data)),                  // Size (assuming small message)
	}
	message = append(message, data...)

	// Create a mock reader that returns our test data
	mockReader := &MockReader{
		data: [][]byte{message},
		err:  errors.New("EOF"),
	}

	// Create a channel to receive the parsed message
	messageChan := make(chan proto.Message, 1)

	// Create the parser
	parser := NewParser(Config{
		Reader: mockReader,
		MessageHandler: func(tag int, message proto.Message) {
			if tag == constants.LoginResponseTag {
				messageChan <- message
			}
		},
		ErrorHandler: func(err error) {
			t.Errorf("Unexpected error: %v", err)
		},
		Debug: false,
	})

	// Start the parser
	parser.Start()

	// Wait for the message or timeout
	select {
	case msg := <-messageChan:
		// Check that we got the expected message
		response, ok := msg.(*fcmproto.LoginResponse)
		if !ok {
			t.Fatalf("Expected LoginResponse, got %T", msg)
		}

		if response.GetId() != "test-id" {
			t.Errorf("Expected ID 'test-id', got '%s'", response.GetId())
		}

	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for message")
	}
}

// TestParser_DataMessageStanza tests parsing a data message stanza
func TestParser_DataMessageStanza(t *testing.T) {
	// Create test data for a data message stanza
	dataMessage := &fcmproto.DataMessageStanza{
		Id:           proto.String("test-message-id"),
		From:         proto.String("test-sender"),
		PersistentId: proto.String("test-persistent-id"),
		Category:     proto.String("test-category"),
	}

	// Marshal the data message
	data, err := proto.Marshal(dataMessage)
	if err != nil {
		t.Fatalf("Failed to marshal data message: %v", err)
	}

	// Create the full message with version, tag, and size
	message := []byte{
		byte(constants.MCSVersion),           // Version
		byte(constants.DataMessageStanzaTag), // Tag
	}

	// Add varint size
	size := len(data)
	var sizeBytes []byte
	for size >= 0x80 {
		sizeBytes = append(sizeBytes, byte(size)|0x80)
		size >>= 7
	}
	sizeBytes = append(sizeBytes, byte(size))

	message = append(message, sizeBytes...)
	message = append(message, data...)

	// Create a mock reader that returns our test data
	mockReader := &MockReader{
		data: [][]byte{message},
		err:  errors.New("EOF"),
	}

	// Create a channel to receive the parsed message
	messageChan := make(chan proto.Message, 1)

	// Create the parser
	parser := NewParser(Config{
		Reader: mockReader,
		MessageHandler: func(tag int, message proto.Message) {
			if tag == constants.DataMessageStanzaTag {
				messageChan <- message
			}
		},
		ErrorHandler: func(err error) {
			t.Errorf("Unexpected error: %v", err)
		},
		Debug: false,
	})

	// Start the parser
	parser.Start()

	// Wait for the message or timeout
	select {
	case msg := <-messageChan:
		// Check that we got the expected message
		stanza, ok := msg.(*fcmproto.DataMessageStanza)
		if !ok {
			t.Fatalf("Expected DataMessageStanza, got %T", msg)
		}

		if stanza.GetId() != "test-message-id" {
			t.Errorf("Expected ID 'test-message-id', got '%s'", stanza.GetId())
		}
		if stanza.GetFrom() != "test-sender" {
			t.Errorf("Expected From 'test-sender', got '%s'", stanza.GetFrom())
		}
		if stanza.GetPersistentId() != "test-persistent-id" {
			t.Errorf("Expected PersistentId 'test-persistent-id', got '%s'", stanza.GetPersistentId())
		}
		if stanza.GetCategory() != "test-category" {
			t.Errorf("Expected Category 'test-category', got '%s'", stanza.GetCategory())
		}

	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for message")
	}
}

// TestParser_ErrorHandling tests the parser's error handling
func TestParser_ErrorHandling(t *testing.T) {
	// Create invalid message data (wrong version)
	message := []byte{
		0x99,                             // Invalid version
		byte(constants.LoginResponseTag), // Tag
		0x01,                             // Size
		0x00,                             // Data (empty)
	}

	// Create a mock reader that returns our test data
	mockReader := &MockReader{
		data: [][]byte{message},
		err:  errors.New("EOF"),
	}

	// Create a channel to receive errors
	errorChan := make(chan error, 1)

	// Create the parser
	parser := NewParser(Config{
		Reader: mockReader,
		MessageHandler: func(tag int, message proto.Message) {
			t.Errorf("Unexpected message: %v", message)
		},
		ErrorHandler: func(err error) {
			errorChan <- err
		},
		Debug: false,
	})

	// Start the parser
	parser.Start()

	// Wait for the error or timeout
	select {
	case err := <-errorChan:
		// We expect an error about invalid version
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		// Check that the error is about invalid version
		if !bytes.Contains([]byte(err.Error()), []byte("version")) {
			t.Errorf("Expected error about invalid version, got: %v", err)
		}

	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for error")
	}
}
