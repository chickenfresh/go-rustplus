package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// Supported versions
	VersionAes128Gcm = "aes128gcm"
	VersionAesGcm    = "aesgcm"

	// Constants
	AesGcm       = "aes-128-gcm"
	TagLength    = 16
	KeyLength    = 16
	NonceLength  = 12
	Sha256Length = 32
	ModeEncrypt  = "encrypt"
	ModeDecrypt  = "decrypt"
)

// PadSize defines the padding size for different versions
var PadSize = map[string]int{
	VersionAes128Gcm: 1,
	VersionAesGcm:    2,
}

// ECEParams contains parameters for ECE encryption/decryption
type ECEParams struct {
	Version    string
	Key        []byte
	Salt       []byte
	AuthSecret []byte
	KeyID      string
	DH         []byte
	PrivateKey interface{} // This would be a crypto.PrivateKey in practice
	RS         int
	KeyMap     map[string][]byte
}

// Decrypt decrypts ECE encrypted content
func Decrypt(buffer []byte, params ECEParams) ([]byte, error) {
	header := params
	if header.Version == VersionAes128Gcm {
		headerLength, err := readHeader(buffer, &header)
		if err != nil {
			return nil, err
		}
		buffer = buffer[headerLength:]
	}

	key, err := deriveKeyAndNonce(header, ModeDecrypt, nil)
	if err != nil {
		return nil, err
	}

	result := []byte{}
	start := 0
	chunkSize := header.RS

	if header.Version != VersionAes128Gcm {
		chunkSize += TagLength
	}

	for i := 0; start < len(buffer); i++ {
		end := start + chunkSize
		if header.Version != VersionAes128Gcm && end == len(buffer) {
			return nil, errors.New("truncated payload")
		}
		end = min(end, len(buffer))
		if end-start <= TagLength {
			return nil, fmt.Errorf("invalid block: too small at %d", i)
		}

		block, err := decryptRecord(key, i, buffer[start:end], header, end >= len(buffer))
		if err != nil {
			return nil, err
		}
		result = append(result, block...)
		start = end
	}

	return result, nil
}

// Encrypt encrypts content using ECE
func Encrypt(buffer []byte, params ECEParams) ([]byte, error) {
	if len(buffer) == 0 {
		return nil, errors.New("buffer argument must not be empty")
	}

	header := params
	if header.Salt == nil {
		salt := make([]byte, KeyLength)
		if _, err := rand.Read(salt); err != nil {
			return nil, err
		}
		header.Salt = salt
	}

	var result []byte
	if header.Version == VersionAes128Gcm {
		// Save the DH public key in the header unless keyid is set
		if header.PrivateKey != nil && header.KeyID == "" {
			// In a real implementation, we would get the public key from the private key
			// header.KeyID = header.PrivateKey.PublicKey
		}
		headerBytes, err := writeHeader(header)
		if err != nil {
			return nil, err
		}
		result = headerBytes
	} else {
		result = []byte{}
	}

	key, err := deriveKeyAndNonce(header, ModeEncrypt, nil)
	if err != nil {
		return nil, err
	}

	start := 0
	padSize := PadSize[header.Version]
	overhead := padSize
	if header.Version == VersionAes128Gcm {
		overhead += TagLength
	}
	pad := 0 // This would be params.Pad in a real implementation

	counter := 0
	for {
		// Pad so that at least one data byte is in a block
		recordPad := min(header.RS-overhead-1, pad)
		if header.Version != VersionAes128Gcm {
			recordPad = min((1<<(padSize*8))-1, recordPad)
		}
		if pad > 0 && recordPad == 0 {
			recordPad++ // Deal with perverse case of rs=overhead+1 with padding
		}
		pad -= recordPad

		end := start + header.RS - overhead - recordPad
		var last bool
		if header.Version != VersionAes128Gcm {
			// The > here ensures that we write out a padding-only block at the end of a buffer
			last = end > len(buffer)
		} else {
			last = end >= len(buffer)
		}
		last = last && pad <= 0

		var blockData []byte
		if end > len(buffer) {
			blockData = buffer[start:]
		} else {
			blockData = buffer[start:end]
		}

		block, err := encryptRecord(key, counter, blockData, recordPad, header, last)
		if err != nil {
			return nil, err
		}
		result = append(result, block...)

		if last {
			break
		}

		start = end
		counter++
	}

	return result, nil
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func hmacHash(key, input []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(input)
	return h.Sum(nil)
}

// HKDF as defined in RFC5869, using SHA-256
func hkdfExtract(salt, ikm []byte) []byte {
	return hmacHash(salt, ikm)
}

func hkdfExpand(prk, info []byte, length int) []byte {
	output := []byte{}
	t := []byte{}
	counter := byte(0)

	for len(output) < length {
		counter++
		data := append(t, info...)
		data = append(data, counter)
		t = hmacHash(prk, data)
		output = append(output, t...)
	}

	return output[:length]
}

func hkdf(salt, ikm, info []byte, length int) []byte {
	prk := hkdfExtract(salt, ikm)
	return hkdfExpand(prk, info, length)
}

func createInfo(base string, context []byte) []byte {
	prefix := []byte("Content-Encoding: " + base + "\x00")
	return append(prefix, context...)
}

func deriveKeyAndNonce(header ECEParams, mode string, lookupKeyCallback func(string) []byte) (map[string][]byte, error) {
	if len(header.Salt) == 0 {
		return nil, errors.New("must include a salt parameter")
	}

	var keyInfo, nonceInfo []byte
	var secret []byte
	var err error

	if header.Version == VersionAesGcm {
		// Old version
		s, err := extractSecretAndContext(header, mode)
		if err != nil {
			return nil, err
		}
		keyInfo = createInfo("aesgcm", s.Context)
		nonceInfo = createInfo("nonce", s.Context)
		secret = s.Secret
	} else if header.Version == VersionAes128Gcm {
		// Latest version
		keyInfo = []byte("Content-Encoding: aes128gcm\x00")
		nonceInfo = []byte("Content-Encoding: nonce\x00")
		secret, err = extractSecret(header, mode, lookupKeyCallback)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unsupported version: %s", header.Version)
	}

	prk := hkdfExtract(header.Salt, secret)
	result := map[string][]byte{
		"key":   hkdfExpand(prk, keyInfo, KeyLength),
		"nonce": hkdfExpand(prk, nonceInfo, NonceLength),
	}

	return result, nil
}

type secretContext struct {
	Secret  []byte
	Context []byte
}

func extractSecretAndContext(header ECEParams, mode string) (secretContext, error) {
	result := secretContext{
		Secret:  nil,
		Context: []byte{},
	}

	if header.Key != nil {
		result.Secret = header.Key
		if len(result.Secret) != KeyLength {
			return result, fmt.Errorf("an explicit key must be %d bytes", KeyLength)
		}
	} else if header.DH != nil {
		// This would extract the DH secret in a real implementation
		// For now, we'll just return an error
		return result, errors.New("DH extraction not implemented")
	} else if header.KeyID != "" {
		result.Secret = header.KeyMap[header.KeyID]
	}

	if result.Secret == nil {
		return result, errors.New("unable to determine key")
	}

	if header.AuthSecret != nil {
		result.Secret = hkdf(header.AuthSecret, result.Secret, createInfo("auth", []byte{}), Sha256Length)
	}

	return result, nil
}

func extractSecret(header ECEParams, mode string, keyLookupCallback func(string) []byte) ([]byte, error) {
	if header.Key != nil {
		if len(header.Key) != KeyLength {
			return nil, fmt.Errorf("an explicit key must be %d bytes", KeyLength)
		}
		return header.Key, nil
	}

	if header.PrivateKey == nil {
		// Lookup based on keyid
		var key []byte
		if keyLookupCallback != nil {
			key = keyLookupCallback(header.KeyID)
		} else {
			key = header.KeyMap[header.KeyID]
		}
		if key == nil {
			return nil, fmt.Errorf("no saved key (keyid: %s)", header.KeyID)
		}
		return key, nil
	}

	// This would compute the WebPush secret in a real implementation
	// For now, we'll just return an error
	return nil, errors.New("WebPush secret extraction not implemented")
}

func generateNonce(base []byte, counter int) []byte {
	nonce := make([]byte, len(base))
	copy(nonce, base)

	// The original implementation does some XOR operations on the last 6 bytes
	// For simplicity, we'll just increment the last byte by the counter
	if len(nonce) > 0 {
		nonce[len(nonce)-1] ^= byte(counter)
	}

	return nonce
}

func readHeader(buffer []byte, header *ECEParams) (int, error) {
	if len(buffer) < 21 {
		return 0, errors.New("buffer too short for header")
	}

	header.Salt = buffer[:KeyLength]
	header.RS = int(binary.BigEndian.Uint32(buffer[KeyLength : KeyLength+4]))
	idSize := int(buffer[20])

	if len(buffer) < 21+idSize {
		return 0, errors.New("buffer too short for key ID")
	}

	if idSize > 0 {
		header.KeyID = string(buffer[21 : 21+idSize])
	} else {
		header.KeyID = ""
	}

	return 21 + idSize, nil
}

func writeHeader(header ECEParams) ([]byte, error) {
	keyID := []byte(header.KeyID)
	if len(keyID) > 255 {
		return nil, errors.New("keyid is too large")
	}

	result := make([]byte, KeyLength+5+len(keyID))
	copy(result[:KeyLength], header.Salt)
	binary.BigEndian.PutUint32(result[KeyLength:KeyLength+4], uint32(header.RS))
	result[KeyLength+4] = byte(len(keyID))
	copy(result[KeyLength+5:], keyID)

	return result, nil
}

func unpadLegacy(data []byte, version string) ([]byte, error) {
	padSize := PadSize[version]
	if len(data) < padSize {
		return nil, errors.New("data too short for padding")
	}

	pad := 0
	for i := 0; i < padSize; i++ {
		pad = (pad << 8) | int(data[i])
	}

	if pad+padSize > len(data) {
		return nil, errors.New("padding exceeds block size")
	}

	// Check that padding bytes are all zeros
	for i := padSize; i < padSize+pad; i++ {
		if data[i] != 0 {
			return nil, errors.New("invalid padding")
		}
	}

	return data[padSize+pad:], nil
}

func unpad(data []byte, last bool) ([]byte, error) {
	i := len(data) - 1
	for i >= 0 {
		if data[i] != 0 {
			if last {
				if data[i] != 2 {
					return nil, errors.New("last record needs to start padding with a 2")
				}
			} else {
				if data[i] != 1 {
					return nil, errors.New("non-last record needs to start padding with a 1")
				}
			}
			return data[:i], nil
		}
		i--
	}
	return nil, errors.New("all zero plaintext")
}

func decryptRecord(key map[string][]byte, counter int, buffer []byte, header ECEParams, last bool) ([]byte, error) {
	nonce := generateNonce(key["nonce"], counter)

	block, err := aes.NewCipher(key["key"])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := buffer[:len(buffer)-TagLength]
	tag := buffer[len(buffer)-TagLength:]

	// In Go, the tag is appended to the ciphertext for AEAD decryption
	ciphertextWithTag := append(ciphertext, tag...)

	plaintext, err := gcm.Open(nil, nonce, ciphertextWithTag, nil)
	if err != nil {
		return nil, err
	}

	if header.Version != VersionAes128Gcm {
		return unpadLegacy(plaintext, header.Version)
	}
	return unpad(plaintext, last)
}

func encryptRecord(key map[string][]byte, counter int, buffer []byte, pad int, header ECEParams, last bool) ([]byte, error) {
	nonce := generateNonce(key["nonce"], counter)

	block, err := aes.NewCipher(key["key"])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	var plaintext []byte
	padSize := PadSize[header.Version]
	padding := make([]byte, pad+padSize)

	if header.Version != VersionAes128Gcm {
		// Legacy padding
		// Write pad size to the first padSize bytes
		for i := 0; i < padSize; i++ {
			padding[i] = byte((pad >> (8 * (padSize - i - 1))) & 0xff)
		}
		plaintext = append(padding, buffer...)
	} else {
		// New padding
		plaintext = buffer
		if last {
			padding[0] = 2
		} else {
			padding[0] = 1
		}
		plaintext = append(plaintext, padding...)
	}

	// In Go, the tag is appended to the ciphertext by the AEAD interface
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	return ciphertext, nil
}
