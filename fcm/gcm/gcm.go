package gcm

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	fcmproto "github.com/chickenfresh/go-rustplus/fcm/proto"
	"github.com/chickenfresh/go-rustplus/fcm/serverkey"
	"google.golang.org/protobuf/proto"
)

const (
	// URLs for GCM/FCM registration
	registerURL = "https://android.clients.google.com/c2dm/register3"
	checkinURL  = "https://android.clients.google.com/checkin"
)

// Credentials represents the GCM registration credentials
type Credentials struct {
	Token         string `json:"token"`
	AndroidID     string `json:"androidId"`
	SecurityToken string `json:"securityToken"`
	AppID         string `json:"appId"`
}

// Register registers with GCM and returns credentials
func Register(androidID, securityToken, appID string) (Credentials, error) {
	// First, check in with GCM
	checkinResp, err := CheckIn(androidID, securityToken)
	if err != nil {
		return Credentials{}, fmt.Errorf("checkin failed: %w", err)
	}

	// Then register with the obtained credentials
	credentials, err := doRegister(checkinResp, appID)
	if err != nil {
		return Credentials{}, fmt.Errorf("registration failed: %w", err)
	}

	return credentials, nil
}

// CheckIn performs a check-in with GCM
func CheckIn(androidID, securityToken string) (*fcmproto.AndroidCheckinResponse, error) {
	// Create the check-in request
	request, err := getCheckinRequest(androidID, securityToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create check-in request: %w", err)
	}

	// Encode the request as protobuf
	data, err := proto.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal check-in request: %w", err)
	}

	// Send the request
	resp, err := http.Post(checkinURL, "application/x-protobuf", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("check-in request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read check-in response: %w", err)
	}

	// Decode the response
	response := &fcmproto.AndroidCheckinResponse{}
	if err := proto.Unmarshal(body, response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal check-in response: %w", err)
	}

	return response, nil
}

// doRegister performs the actual registration with GCM
func doRegister(checkinResp *fcmproto.AndroidCheckinResponse, appID string) (Credentials, error) {
	androidID := fmt.Sprintf("%d", checkinResp.GetAndroidId())
	securityToken := fmt.Sprintf("%d", checkinResp.GetSecurityToken())

	// Prepare the form data
	form := url.Values{}
	form.Add("app", "org.chromium.linux")
	form.Add("X-subtype", appID)
	form.Add("device", androidID)
	form.Add("sender", base64.StdEncoding.EncodeToString(serverkey.Key))

	// Send the registration request
	response, err := postRegister(androidID, securityToken, form, 0)
	if err != nil {
		return Credentials{}, err
	}

	// Parse the response
	parts := strings.Split(response, "=")
	if len(parts) < 2 {
		return Credentials{}, errors.New("invalid registration response format")
	}

	token := parts[1]

	return Credentials{
		Token:         token,
		AndroidID:     androidID,
		SecurityToken: securityToken,
		AppID:         appID,
	}, nil
}

// postRegister sends a registration request to GCM with retry logic
func postRegister(androidID, securityToken string, form url.Values, retry int) (string, error) {
	// Create the request
	req, err := http.NewRequest("POST", registerURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create registration request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("AidLogin %s:%s", androidID, securityToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read registration response: %w", err)
	}

	response := string(body)

	// Check for errors
	if strings.Contains(response, "Error") {
		if retry >= 5 {
			return "", errors.New("GCM register has failed after 5 retries")
		}

		// Wait and retry
		time.Sleep(1 * time.Second)
		return postRegister(androidID, securityToken, form, retry+1)
	}

	return response, nil
}

// getCheckinRequest creates a check-in request protobuf
func getCheckinRequest(androidID, securityToken string) (*fcmproto.AndroidCheckinRequest, error) {
	request := &fcmproto.AndroidCheckinRequest{
		UserSerialNumber: proto.Int32(0),
		Checkin: &fcmproto.AndroidCheckinProto{
			Type: proto.Int32(3), // DEVICE_CHROME_BROWSER
			ChromeBuild: &fcmproto.ChromeBuildProto{
				Platform:      fcmproto.ChromeBuildProto_PLATFORM_MAC.Enum(),
				ChromeVersion: proto.String("63.0.3234.0"),
				Channel:       fcmproto.ChromeBuildProto_CHANNEL_STABLE.Enum(),
			},
		},
		Version: proto.Int32(3),
	}

	// Set ID and security token if provided
	if androidID != "" {
		id, err := parseInt64(androidID)
		if err != nil {
			return nil, fmt.Errorf("invalid android ID: %w", err)
		}
		request.Id = &id
	}

	if securityToken != "" {
		token, err := parseUint64(securityToken)
		if err != nil {
			return nil, fmt.Errorf("invalid security token: %w", err)
		}
		request.SecurityToken = &token
	}

	return request, nil
}

// parseInt64 parses a string as int64
func parseInt64(s string) (int64, error) {
	var value int64
	_, err := fmt.Sscanf(s, "%d", &value)
	return value, err
}

// parseUint64 parses a string as uint64
func parseUint64(s string) (uint64, error) {
	var value uint64
	_, err := fmt.Sscanf(s, "%d", &value)
	return value, err
}
