package fcm

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/chickenfresh/go-rustplus/fcm/android"
	"github.com/google/uuid"
)

const (
	FCMSubscribe = "https://fcm.googleapis.com/fcm/connect/subscribe"
	FCMEndpoint  = "https://fcm.googleapis.com/fcm/send"
)

type GCMCredentials struct {
	Token         string `json:"token"`
	AndroidId     string `json:"androidId"`
	SecurityToken string `json:"securityToken"`
	AppId         string `json:"appId"`
}

type FCMCredentials struct {
	Token   string `json:"token"`
	PushSet string `json:"pushSet"`
	Keys    struct {
		Private string `json:"private"`
		Public  string `json:"public"`
		Secret  string `json:"secret"`
		Auth    string `json:"auth"`
	} `json:"keys"`
}

// Credentials represents the authentication credentials for FCM
type Credentials struct {
	GCM GCMCredentials `json:"gcm"`
	FCM FCMCredentials `json:"fcm"`
}

// Keys represents the cryptographic keys used for FCM
type Keys struct {
	Private string `json:"privateKey"`
	Public  string `json:"publicKey"`
	Auth    string `json:"authSecret"`
}

// FCMResponse represents the response from FCM registration
type FCMResponse struct {
	Token string `json:"token"`
}

// Register registers with FCM and returns credentials
func Register(senderId string) (Credentials, error) {
	var credentials Credentials

	// Generate a unique Android ID
	androidId, err := generateAndroidId()
	if err != nil {
		return credentials, fmt.Errorf("failed to generate Android ID: %w", err)
	}

	// Generate security token
	securityToken, err := generateSecurityToken()
	if err != nil {
		return credentials, fmt.Errorf("failed to generate security token: %w", err)
	}

	// Register with GCM
	err = registerGCM(senderId, androidId, securityToken, &credentials)
	if err != nil {
		return credentials, fmt.Errorf("GCM registration failed: %w", err)
	}

	// Generate keys for FCM
	err = generateKeys(&credentials)
	if err != nil {
		return credentials, fmt.Errorf("failed to generate keys: %w", err)
	}

	// Register with FCM
	err = registerFCM(senderId, &credentials)
	if err != nil {
		return credentials, fmt.Errorf("FCM registration failed: %w", err)
	}

	return credentials, nil
}

// generateAndroidId generates a unique Android ID
func generateAndroidId() (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	// Convert UUID to Android ID format (remove hyphens and use only first 16 chars)
	androidId := strings.Replace(uuid.String(), "-", "", -1)
	return androidId[:16], nil
}

// generateSecurityToken generates a security token
func generateSecurityToken() (string, error) {
	// In the original implementation, this is a long integer
	// For simplicity, we'll use a random number
	token := make([]byte, 8)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	// Convert to hex string
	return fmt.Sprintf("%x", token), nil
}

// registerGCM registers with GCM service
func registerGCM(senderId, androidId, securityToken string, credentials *Credentials) error {
	// TODO: Implement actual GCM registration
	// This is a placeholder implementation

	credentials.GCM.AndroidId = androidId
	credentials.GCM.SecurityToken = securityToken
	credentials.GCM.AppId = fmt.Sprintf("wp:receiver.push.com#%s", uuid.New().String())
	credentials.GCM.Token = fmt.Sprintf("APA91b%s", uuid.New().String())

	return nil
}

// generateKeys generates the necessary keys for FCM
func generateKeys(credentials *Credentials) error {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Encode private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	credentials.FCM.Keys.Private = base64.StdEncoding.EncodeToString(privateKeyBytes)

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}
	credentials.FCM.Keys.Public = base64.StdEncoding.EncodeToString(publicKeyBytes)

	// Generate auth secret
	authSecret := make([]byte, 16)
	_, err = rand.Read(authSecret)
	if err != nil {
		return err
	}
	credentials.FCM.Keys.Auth = base64.StdEncoding.EncodeToString(authSecret)

	return nil
}

// registerFCM registers with FCM service
func registerFCM(senderId string, credentials *Credentials) error {
	// TODO: Implement actual FCM registration
	// This is a placeholder implementation

	credentials.FCM.Token = fmt.Sprintf("fcm-%s", uuid.New().String())
	credentials.FCM.PushSet = uuid.New().String()

	return nil
}

// RegisterFCM registers with FCM using the provided sender ID and token
func RegisterFCM(senderID, token string) (Keys, FCMResponse, error) {
	// Create keys
	keys, err := createKeys()
	if err != nil {
		return Keys{}, FCMResponse{}, fmt.Errorf("failed to create keys: %w", err)
	}

	// Prepare form data
	form := url.Values{}
	form.Add("authorized_entity", senderID)
	form.Add("endpoint", fmt.Sprintf("%s/%s", FCMEndpoint, token))
	form.Add("encryption_key", formatURLBase64(keys.Public))
	form.Add("encryption_auth", formatURLBase64(keys.Auth))

	// Create request
	req, err := http.NewRequest("POST", FCMSubscribe, strings.NewReader(form.Encode()))
	if err != nil {
		return Keys{}, FCMResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Keys{}, FCMResponse{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Keys{}, FCMResponse{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var fcmResponse FCMResponse
	if err := json.Unmarshal(body, &fcmResponse); err != nil {
		return Keys{}, FCMResponse{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return keys, fcmResponse, nil
}

// createKeys generates the cryptographic keys needed for FCM
func createKeys() (Keys, error) {
	// Generate ECDH key pair
	curve := ecdh.P256()
	privateKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return Keys{}, fmt.Errorf("failed to generate ECDH key: %w", err)
	}

	// Generate auth secret
	authSecret := make([]byte, 16)
	if _, err := rand.Read(authSecret); err != nil {
		return Keys{}, fmt.Errorf("failed to generate auth secret: %w", err)
	}

	// Encode keys
	keys := Keys{
		Private: base64.StdEncoding.EncodeToString(privateKey.Bytes()),
		Public:  base64.StdEncoding.EncodeToString(privateKey.PublicKey().Bytes()),
		Auth:    base64.StdEncoding.EncodeToString(authSecret),
	}

	return keys, nil
}

// formatURLBase64 formats a base64 string for use in URLs
func formatURLBase64(input string) string {
	return strings.TrimRight(
		strings.ReplaceAll(
			strings.ReplaceAll(input, "+", "-"),
			"/", "_",
		),
		"=",
	)
}

// RegisterAndroid registers with FCM using the Android method
func RegisterAndroid(apiKey, projectID, gcmSenderID, gmsAppID, androidPackageName, androidPackageCert string) (Credentials, error) {
	var credentials Credentials

	// Create Android FCM client
	androidFCM := &android.AndroidFCM{}

	// Register with FCM
	resp, err := androidFCM.Register(android.RegisterOptions{
		APIKey:             apiKey,
		ProjectID:          projectID,
		GCMSenderID:        gcmSenderID,
		GMSAppID:           gmsAppID,
		AndroidPackageName: androidPackageName,
		AndroidPackageCert: androidPackageCert,
	})
	if err != nil {
		return credentials, fmt.Errorf("Android FCM registration failed: %w", err)
	}

	// Set credentials
	credentials.GCM.AndroidId = resp.GCM.AndroidID
	credentials.GCM.SecurityToken = resp.GCM.SecurityToken
	credentials.FCM.Token = resp.FCM.Token

	// Generate keys for FCM
	err = generateKeys(&credentials)
	if err != nil {
		return credentials, fmt.Errorf("failed to generate keys: %w", err)
	}

	return credentials, nil
}
