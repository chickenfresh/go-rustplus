package constants

// Processing states for the MCS protocol
const (
	// Processing the version, tag, and size packets (assuming minimum length
	// size packet). Only used during the login handshake.
	MCSVersionTagAndSize = 0
	// Processing the tag and size packets (assuming minimum length size
	// packet). Used for normal messages.
	MCSTagAndSize = 1
	// Processing the size packet alone.
	MCSSize = 2
	// Processing the protocol buffer bytes (for those messages with non-zero
	// sizes).
	MCSProtoBytes = 3
)

// Packet length constants
const (
	// Number of bytes a MCS version packet consumes.
	VersionPacketLen = 1
	// Number of bytes a tag packet consumes.
	TagPacketLen = 1
	// Min number of bytes a length packet consumes.
	SizePacketLenMin = 1
	// Max number of bytes a length packet consumes. A Varint32 can consume up to 5 bytes
	// (the msb in each byte is reserved for denoting whether more bytes follow).
	SizePacketLenMax = 5
)

// Protocol version
const (
	// The current MCS protocol version.
	MCSVersion = 41
)

// MCS Message tags
const (
	// WARNING: the order of these tags must remain the same, as the tag values
	// must be consistent with those used on the server.
	HeartbeatPingTag       = 0
	HeartbeatAckTag        = 1
	LoginRequestTag        = 2
	LoginResponseTag       = 3
	CloseTag               = 4
	MessageStanzaTag       = 5
	PresenceStanzaTag      = 6
	IqStanzaTag            = 7
	DataMessageStanzaTag   = 8
	BatchPresenceStanzaTag = 9
	StreamErrorStanzaTag   = 10
	HttpRequestTag         = 11
	HttpResponseTag        = 12
	BindAccountRequestTag  = 13
	BindAccountResponseTag = 14
	TalkMetadataTag        = 15
	NumProtoTypes          = 16
)

// Connection constants
const (
	// Default MCS endpoint
	MCSEndpoint = "mtalk.google.com:5228"
	// Default connection timeout in seconds
	ConnectionTimeout = 30
	// Default heartbeat interval in seconds
	HeartbeatInterval = 60
)

// Error constants
const (
	// Error codes
	ErrorConnectionFailed = "CONNECTION_FAILED"
	ErrorAuthFailed       = "AUTH_FAILED"
	ErrorInvalidMessage   = "INVALID_MESSAGE"
)
