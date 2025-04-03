package parser

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/chickenfresh/go-rustplus/fcm/constants"
	fcmproto "github.com/chickenfresh/go-rustplus/fcm/proto"
	"google.golang.org/protobuf/proto"
)

// Common errors
var (
	ErrInvalidState  = errors.New("invalid parser state")
	ErrUnknownTag    = errors.New("unknown message tag")
	ErrVarIntTooLong = errors.New("varint too long")
)

// Debug controls whether debug messages are printed
var Debug = false

// debugPrint prints debug messages if Debug is enabled
func debugPrint(format string, args ...interface{}) {
	if Debug {
		log.Printf(format, args...)
	}
}

// MessageHandler processes parsed messages
type MessageHandler func(tag int, message proto.Message)

// ErrorHandler processes parser errors
type ErrorHandler func(error)

// Reader represents a connection that can be read from
type Reader interface {
	Read([]byte) (int, error)
}

// Config contains parser configuration
type Config struct {
	Reader         Reader
	MessageHandler MessageHandler
	ErrorHandler   ErrorHandler
	Debug          bool
}

// Parser parses wire data from FCM/GCM
// This is equivalent to the Parser class in the Node.js implementation
type Parser struct {
	config            Config
	state             int
	data              []byte
	sizePacketSoFar   int
	messageTag        int
	messageSize       int
	handshakeComplete bool
	isWaitingForData  bool
	ctx               context.Context
	cancel            context.CancelFunc
}

// NewParser creates a new Parser instance
func NewParser(config Config) *Parser {
	ctx, cancel := context.WithCancel(context.Background())

	return &Parser{
		config:            config,
		state:             constants.MCSVersionTagAndSize,
		data:              []byte{},
		sizePacketSoFar:   0,
		messageTag:        0,
		messageSize:       0,
		handshakeComplete: false,
		isWaitingForData:  true,
		ctx:               ctx,
		cancel:            cancel,
	}
}

// Start begins reading and parsing data from the connection
func (p *Parser) Start() {
	go p.readLoop()
}

// Stop stops the parser
func (p *Parser) Stop() {
	p.cancel()
}

// readLoop continuously reads data from the connection
func (p *Parser) readLoop() {
	buffer := make([]byte, 8192) // 8KB buffer

	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			n, err := p.config.Reader.Read(buffer)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					p.emitError(fmt.Errorf("read error: %w", err))
				}
				return
			}

			if n > 0 {
				p.onData(buffer[:n])
			}
		}
	}
}

// emitError sends an error to the error handler
func (p *Parser) emitError(err error) {
	if p.config.ErrorHandler != nil {
		p.config.ErrorHandler(err)
	}
}

// onData processes incoming data
func (p *Parser) onData(buffer []byte) {
	p.debugPrint("Got data: %d bytes", len(buffer))
	p.data = append(p.data, buffer...)

	if p.isWaitingForData {
		p.isWaitingForData = false
		p.waitForData()
	}
}

// waitForData processes data based on the current state
func (p *Parser) waitForData() {
	debugPrint("waitForData state: %d", p.state)

	var minBytesNeeded int

	switch p.state {
	case constants.MCSVersionTagAndSize:
		minBytesNeeded = constants.VersionPacketLen + constants.TagPacketLen + constants.SizePacketLenMin
	case constants.MCSTagAndSize:
		minBytesNeeded = constants.TagPacketLen + constants.SizePacketLenMin
	case constants.MCSSize:
		minBytesNeeded = p.sizePacketSoFar + 1
	case constants.MCSProtoBytes:
		minBytesNeeded = p.messageSize
	default:
		p.emitError(fmt.Errorf("unexpected state: %d", p.state))
		return
	}

	if len(p.data) < minBytesNeeded {
		debugPrint("Socket read finished prematurely. Waiting for %d more bytes",
			minBytesNeeded-len(p.data))
		p.isWaitingForData = true
		return
	}

	debugPrint("Processing MCS data: state == %d", p.state)

	switch p.state {
	case constants.MCSVersionTagAndSize:
		p.onGotVersion()
	case constants.MCSTagAndSize:
		p.onGotMessageTag()
	case constants.MCSSize:
		p.onGotMessageSize()
	case constants.MCSProtoBytes:
		p.onGotMessageBytes()
	default:
		p.emitError(fmt.Errorf("unexpected state: %d", p.state))
	}
}

// onGotVersion processes the version byte
func (p *Parser) onGotVersion() {
	version := int(p.data[0])
	p.data = p.data[1:]
	debugPrint("VERSION IS %d", version)

	if version < constants.MCSVersion && version != 38 {
		p.emitError(fmt.Errorf("got wrong version: %d", version))
		return
	}

	// Process the LoginResponse message tag
	p.onGotMessageTag()
}

// onGotMessageTag processes the message tag
func (p *Parser) onGotMessageTag() {
	p.messageTag = int(p.data[0])
	p.data = p.data[1:]
	debugPrint("RECEIVED PROTO OF TYPE %d", p.messageTag)

	p.onGotMessageSize()
}

// onGotMessageSize processes the message size
func (p *Parser) onGotMessageSize() {
	reader := bytes.NewReader(p.data)

	// Try to read the varint for message size
	var size uint64
	var err error
	var bytesRead int

	// Read the varint
	size, bytesRead, err = readVarInt(reader)

	if err != nil {
		// If we couldn't read the full varint, wait for more data
		p.sizePacketSoFar = len(p.data)
		p.state = constants.MCSSize
		p.waitForData()
		return
	}

	p.messageSize = int(size)
	p.data = p.data[bytesRead:]

	debugPrint("Proto size: %d", p.messageSize)
	p.sizePacketSoFar = 0

	if p.messageSize > 0 {
		p.state = constants.MCSProtoBytes
		p.waitForData()
	} else {
		p.onGotMessageBytes()
	}
}

// onGotMessageBytes processes the message bytes
func (p *Parser) onGotMessageBytes() {
	protoMsg, err := p.buildProtobufFromTag(p.messageTag)
	if err != nil {
		p.emitError(err)
		// Reset state and continue
		p.getNextMessage()
		return
	}

	// If protoMsg is nil, we can't unmarshal into it
	if protoMsg == nil {
		p.emitError(fmt.Errorf("nil protobuf message for tag %d", p.messageTag))
		// Reset state and continue
		p.getNextMessage()
		return
	}

	// Messages with no content are valid; just use the default protobuf for that tag
	if p.messageSize == 0 {
		p.config.MessageHandler(p.messageTag, protoMsg)
		p.getNextMessage()
		return
	}

	if len(p.data) < p.messageSize {
		// Continue reading data
		debugPrint("Continuing data read. Buffer size is %d, expecting %d",
			len(p.data), p.messageSize)
		p.state = constants.MCSProtoBytes
		p.waitForData()
		return
	}

	buffer := p.data[:p.messageSize]
	p.data = p.data[p.messageSize:]

	// Debug output
	debugPrint("Unmarshaling %d bytes for tag %d", len(buffer), p.messageTag)

	err = proto.Unmarshal(buffer, protoMsg)
	if err != nil {
		p.emitError(fmt.Errorf("failed to unmarshal protobuf: %w", err))
		// Reset state and continue
		p.getNextMessage()
		return
	}

	p.config.MessageHandler(p.messageTag, protoMsg)

	if p.messageTag == constants.LoginResponseTag {
		if p.handshakeComplete {
			log.Println("Unexpected login response")
		} else {
			p.handshakeComplete = true
			debugPrint("GCM Handshake complete.")
		}
	}

	p.getNextMessage()
}

// getNextMessage prepares to read the next message
func (p *Parser) getNextMessage() {
	p.messageTag = 0
	p.messageSize = 0
	p.state = constants.MCSTagAndSize
	p.waitForData()
}

// buildProtobufFromTag creates the appropriate protobuf message for a tag
func (p *Parser) buildProtobufFromTag(tag int) (proto.Message, error) {
	switch tag {
	case constants.HeartbeatPingTag:
		return &fcmproto.HeartbeatPing{}, nil
	case constants.HeartbeatAckTag:
		return &fcmproto.HeartbeatAck{}, nil
	case constants.LoginRequestTag:
		return &fcmproto.LoginRequest{}, nil
	case constants.LoginResponseTag:
		return &fcmproto.LoginResponse{}, nil
	case constants.CloseTag:
		// Close doesn't have a protobuf in the original code
		// Return an empty struct to avoid nil pointer dereference
		return &fcmproto.Close{}, nil
	case constants.IqStanzaTag:
		// IqStanza not defined in our proto file yet
		// Return an empty struct to avoid nil pointer dereference
		return &fcmproto.IqStanza{}, nil
	case constants.DataMessageStanzaTag:
		return &fcmproto.DataMessageStanza{}, nil
	case constants.StreamErrorStanzaTag:
		// StreamErrorStanza not defined in our proto file yet
		// Return an empty struct to avoid nil pointer dereference
		return &fcmproto.StreamErrorStanza{}, nil
	default:
		return nil, fmt.Errorf("unknown tag: %d", tag)
	}
}

// readVarInt reads a varint from a reader
func readVarInt(r io.Reader) (uint64, int, error) {
	var x uint64
	var s uint
	var i int

	for i = 0; i < 10; i++ { // 10 is max length of varint64
		b := make([]byte, 1)
		_, err := r.Read(b)
		if err != nil {
			return 0, i, err
		}

		if b[0] < 0x80 {
			if i == 9 && b[0] > 1 {
				return 0, i + 1, errors.New("overflow")
			}
			return x | uint64(b[0])<<s, i + 1, nil
		}

		x |= uint64(b[0]&0x7f) << s
		s += 7
	}

	return 0, i, errors.New("varint too long")
}

// debugPrint prints debug messages if Debug is enabled
func (p *Parser) debugPrint(format string, args ...interface{}) {
	if p.config.Debug {
		log.Printf(format, args...)
	}
}
