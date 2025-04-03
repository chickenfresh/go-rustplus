package android

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chickenfresh/go-rustplus/fcm/gcm"
)

// AndroidFCM handles Firebase Cloud Messaging registration for Android
type AndroidFCM struct{}

// FirebaseInstallationResponse represents the response from Firebase Installation API
type FirebaseInstallationResponse struct {
	AuthToken struct {
		Token string `json:"token"`
	} `json:"authToken"`
}

// RegisterOptions contains the options for FCM registration
type RegisterOptions struct {
	APIKey             string
	ProjectID          string
	GCMSenderID        string
	GMSAppID           string
	AndroidPackageName string
	AndroidPackageCert string
}

// RegisterResponse contains the response from FCM registration
type RegisterResponse struct {
	GCM struct {
		AndroidID     string `json:"androidId"`
		SecurityToken string `json:"securityToken"`
	} `json:"gcm"`
	FCM struct {
		Token string `json:"token"`
	} `json:"fcm"`
}

// Register registers with FCM and returns credentials
func (a *AndroidFCM) Register(opts RegisterOptions) (*RegisterResponse, error) {
	// Create Firebase installation
	installationAuthToken, err := a.installRequest(
		opts.APIKey,
		opts.ProjectID,
		opts.GMSAppID,
		opts.AndroidPackageName,
		opts.AndroidPackageCert,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firebase installation: %w", err)
	}

	// Check in with GCM
	checkInResp, err := gcm.CheckIn("", "")
	if err != nil {
		return nil, fmt.Errorf("GCM check-in failed: %w", err)
	}

	// Register with GCM
	androidID := fmt.Sprintf("%d", checkInResp.GetAndroidId())
	securityToken := fmt.Sprintf("%d", checkInResp.GetSecurityToken())

	fcmToken, err := a.registerRequest(
		androidID,
		securityToken,
		installationAuthToken,
		opts.APIKey,
		opts.GCMSenderID,
		opts.GMSAppID,
		opts.AndroidPackageName,
		opts.AndroidPackageCert,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("FCM registration failed: %w", err)
	}

	// Create response
	response := &RegisterResponse{}
	response.GCM.AndroidID = androidID
	response.GCM.SecurityToken = securityToken
	response.FCM.Token = fcmToken

	return response, nil
}

// installRequest creates a Firebase installation
func (a *AndroidFCM) installRequest(apiKey, projectID, gmsAppID, androidPackage, androidCert string) (string, error) {
	// Generate Firebase FID
	fid, err := a.generateFirebaseFID()
	if err != nil {
		return "", fmt.Errorf("failed to generate Firebase FID: %w", err)
	}

	// Create request body
	requestBody := map[string]interface{}{
		"fid":         fid,
		"appId":       gmsAppID,
		"authVersion": "FIS_v2",
		"sdkVersion":  "a:17.0.0",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request
	url := fmt.Sprintf("https://firebaseinstallations.googleapis.com/v1/projects/%s/installations", projectID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Android-Package", androidPackage)
	req.Header.Set("X-Android-Cert", androidCert)
	req.Header.Set("x-firebase-client", "android-min-sdk/23 fire-core/20.0.0 device-name/a21snnxx device-brand/samsung device-model/a21s android-installer/com.android.vending fire-android/30 fire-installations/17.0.0 fire-fcm/22.0.0 android-platform/ kotlin/1.9.23 android-target-sdk/34")
	req.Header.Set("x-firebase-client-log-type", "3")
	req.Header.Set("x-goog-api-key", apiKey)
	req.Header.Set("User-Agent", "Dalvik/2.1.0 (Linux; U; Android 11; SM-A217F Build/RP1A.200720.012)")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var installationResp FirebaseInstallationResponse
	if err := json.Unmarshal(body, &installationResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Ensure auth token received
	if installationResp.AuthToken.Token == "" {
		return "", fmt.Errorf("failed to get Firebase installation AuthToken: %s", string(body))
	}

	return installationResp.AuthToken.Token, nil
}

// registerRequest registers with GCM
func (a *AndroidFCM) registerRequest(androidID, securityToken, installationAuthToken, apiKey, gcmSenderID, gmsAppID, androidPackageName, androidPackageCert string, retry int) (string, error) {
	// Prepare form data
	form := url.Values{}
	form.Add("device", androidID)
	form.Add("app", androidPackageName)
	form.Add("cert", androidPackageCert)
	form.Add("app_ver", "1")
	form.Add("X-subtype", gcmSenderID)
	form.Add("X-app_ver", "1")
	form.Add("X-osv", "29")
	form.Add("X-cliv", "fiid-21.1.1")
	form.Add("X-gmsv", "220217001")
	form.Add("X-scope", "*")
	form.Add("X-Goog-Firebase-Installations-Auth", installationAuthToken)
	form.Add("X-gms_app_id", gmsAppID)
	form.Add("X-Firebase-Client", "android-min-sdk/23 fire-core/20.0.0 device-name/a21snnxx device-brand/samsung device-model/a21s android-installer/com.android.vending fire-android/30 fire-installations/17.0.0 fire-fcm/22.0.0 android-platform/ kotlin/1.9.23 android-target-sdk/34")
	form.Add("X-Firebase-Client-Log-Type", "1")
	form.Add("X-app_ver_name", "1")
	form.Add("target_ver", "31")
	form.Add("sender", gcmSenderID)

	// Create request
	req, err := http.NewRequest("POST", "https://android.clients.google.com/c2dm/register3", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("AidLogin %s:%s", androidID, securityToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	response := string(body)

	// Retry if needed
	if strings.Contains(response, "Error") {
		if retry >= 5 {
			return "", fmt.Errorf("GCM register has failed after 5 retries: %s", response)
		}

		// Wait and retry
		time.Sleep(1 * time.Second)
		return a.registerRequest(androidID, securityToken, installationAuthToken, apiKey, gcmSenderID, gmsAppID, androidPackageName, androidPackageCert, retry+1)
	}

	// Extract FCM token from response
	parts := strings.Split(response, "=")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid response format: %s", response)
	}

	return parts[1], nil
}

// generateFirebaseFID generates a Firebase Installation ID
func (a *AndroidFCM) generateFirebaseFID() (string, error) {
	// Generate random bytes
	buf := make([]byte, 17)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Replace the first 4 bits with the constant FID header of 0b0111
	buf[0] = 0b01110000 | (buf[0] & 0b00001111)

	// Encode to base64 and remove padding
	return strings.TrimRight(base64.StdEncoding.EncodeToString(buf), "="), nil
}
