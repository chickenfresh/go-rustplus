package crypto

import (
	"crypto/ecdh"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

// AppDataItem represents a key-value pair in the app data
type AppDataItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// EncryptedMessage represents an encrypted FCM message
type EncryptedMessage struct {
	AppData []AppDataItem `json:"appData"`
	RawData []byte        `json:"rawData"`
}

// Keys contains the keys needed for decryption
type Keys struct {
	PrivateKey string `json:"privateKey"`
	AuthSecret string `json:"authSecret"`
}

// Decrypt decrypts an encrypted FCM message using the provided keys
// https://tools.ietf.org/html/draft-ietf-webpush-encryption-03
func DecryptMessage(message EncryptedMessage, keys Keys) (map[string]interface{}, error) {
	// Find crypto-key in app data
	var cryptoKeyValue string
	var saltValue string

	for _, item := range message.AppData {
		if item.Key == "crypto-key" {
			cryptoKeyValue = item.Value
		} else if item.Key == "encryption" {
			saltValue = item.Value
		}
	}

	if cryptoKeyValue == "" {
		return nil, errors.New("crypto-key is missing")
	}

	if saltValue == "" {
		return nil, errors.New("salt is missing")
	}

	// Extract the DH value from crypto-key (removing the "dh=" prefix)
	dhValue := strings.TrimPrefix(cryptoKeyValue, "dh=")
	if dhValue == cryptoKeyValue {
		// If no change, try another common format
		dhValue = strings.TrimPrefix(cryptoKeyValue, "p256ecdsa=")
	}

	// Extract the salt value (removing the "salt=" prefix)
	salt := strings.TrimPrefix(saltValue, "salt=")
	if salt == saltValue {
		// If no change, try another format
		salt = strings.TrimPrefix(saltValue, "keyid=")
	}

	// Decode the private key
	privateKeyBytes, err := base64.StdEncoding.DecodeString(keys.PrivateKey)
	if err != nil {
		return nil, err
	}

	// Parse the private key
	// Note: In the original Node.js code, they use crypto.createECDH('prime256v1')
	// In Go, we need to parse the private key from the bytes
	parsedKey, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	// Convert to ECDH private key
	privateKey, ok := parsedKey.(*ecdh.PrivateKey)
	if !ok {
		return nil, errors.New("invalid private key type")
	}

	// Decode the auth secret
	authSecret, err := base64.StdEncoding.DecodeString(keys.AuthSecret)
	if err != nil {
		return nil, err
	}

	// Decode the DH value
	dhBytes, err := base64.StdEncoding.DecodeString(dhValue)
	if err != nil {
		return nil, err
	}

	// Decode the salt
	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return nil, err
	}

	// Set up the ECE parameters
	params := ECEParams{
		Version:    VersionAesGcm,
		AuthSecret: authSecret,
		DH:         dhBytes,
		PrivateKey: privateKey,
		Salt:       saltBytes,
		RS:         4096, // Default record size
	}

	// Decrypt the message
	decrypted, err := Decrypt(message.RawData, params)
	if err != nil {
		return nil, err
	}

	// Parse the JSON result
	var result map[string]interface{}
	if err := json.Unmarshal(decrypted, &result); err != nil {
		return nil, err
	}

	return result, nil
}
