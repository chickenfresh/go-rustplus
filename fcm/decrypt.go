package fcm

import (
	"encoding/base64"
	"encoding/json"

	"github.com/chickenfresh/go-rustplus/fcm/crypto"
)

// DecryptNotification decrypts an encrypted FCM notification
func DecryptNotification(notification Notification, credentials Credentials) (map[string]interface{}, error) {
	// Extract app data from the notification
	var appData []crypto.AppDataItem

	// In the original Node.js code, app data is extracted from the notification
	// We'll need to adapt this based on how notifications are structured in our Go implementation
	if appDataRaw, ok := notification.Data["appData"]; ok {
		// Parse app data from the notification
		appDataBytes, err := json.Marshal(appDataRaw)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(appDataBytes, &appData); err != nil {
			return nil, err
		}
	}

	// Extract raw data from the notification
	var rawData []byte
	if rawDataStr, ok := notification.Data["rawData"].(string); ok {
		var err error
		rawData, err = base64.StdEncoding.DecodeString(rawDataStr)
		if err != nil {
			return nil, err
		}
	}

	// Create the encrypted message
	message := crypto.EncryptedMessage{
		AppData: appData,
		RawData: rawData,
	}

	// Extract keys from credentials
	keys := crypto.Keys{
		PrivateKey: credentials.FCM.Keys.Private,
		AuthSecret: credentials.FCM.Keys.Auth,
	}

	// Decrypt the message
	return crypto.DecryptMessage(message, keys)
}
