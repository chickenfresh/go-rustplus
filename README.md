# Rust+ CLI

A command-line interface for interacting with the Rust+ companion app API. This tool allows you to register with FCM (Firebase Cloud Messaging), link your Steam account with Rust+, and listen for notifications from Rust servers.

## Features

- **FCM Registration**: Register with Firebase Cloud Messaging to receive push notifications
- **Steam Authentication**: Link your Steam account with Rust+ using a secure web-based flow
- **Notification Listening**: Listen for and display notifications from Rust servers
- **Expo Push Token**: Generate an Expo Push Token for use with the Rust+ API
- **Rust+ API Integration**: Register with the Rust+ API to receive server notifications

## Installation

### Prerequisites

- Go 1.16 or higher
- Google Chrome or Chromium browser

### Building from Source

1. Clone the repository:
   ```
   git clone https://github.com/chickenfresh/go-rustplus.git
   cd rustplus-cli
   ```

2. Build the project:
   ```
   go build -o rustplus ./cmd/rustplus
   ```

3. (Optional) Install the binary to your PATH:
   ```
   go install ./cmd/rustplus
   ```

## Usage

### Basic Commands

```
rustplus [options] <command>
```

### Available Commands

- `fcm-register`: Register with FCM, generate an Expo Push Token, and link your Steam account with Rust+
- `fcm-listen`: Listen for notifications from FCM, such as Rust+ pairing notifications
- `help`: Display usage information

### Options

- `--config-file`: Path to the configuration file (default: `./rustplus.config.json`)

### Examples

#### Register with FCM and Link Steam Account

```
rustplus fcm-register
```

This command will:
1. Register with Firebase Cloud Messaging
2. Generate an Expo Push Token
3. Open a browser window for Steam authentication
4. Link your Steam account with Rust+
5. Register with the Rust+ API
6. Save all credentials to the configuration file

#### Listen for Notifications

```
rustplus fcm-listen
```

This command will start listening for notifications from FCM, including Rust+ pairing notifications and server alerts.

## Authentication Flow

The authentication process works as follows:

1. The CLI starts a local web server on port 3000
2. A Chrome/Chromium browser window opens to `http://localhost:3000`
3. You click the "Start Steam Authentication" button
4. A popup window opens to the Rust+ login page
5. You log in with your Steam account
6. After successful authentication, the popup window closes
7. The main window shows a success message and automatically closes
8. The CLI receives your authentication token and continues the setup process

## Configuration

The configuration is stored in a JSON file (default: `rustplus.config.json`) with the following structure:

```json
{
  "fcm_credentials": {
    "fcm": {
      "token": "your-fcm-token"
    },
    "gcm": {
      "androidId": "your-android-id",
      "securityToken": "your-security-token"
    }
  },
  "expo_push_token": "your-expo-push-token",
  "rustplus_auth_token": "your-rustplus-auth-token"
}
```

## Security Considerations

- The configuration file contains sensitive tokens and should be kept secure
- The CLI uses a temporary Chrome profile to handle the authentication flow
- All communication with the Rust+ API is done over HTTPS

## Troubleshooting

### Browser Issues

If the browser fails to open automatically, you can manually navigate to `http://localhost:3000` in Chrome or Chromium.

### Authentication Failures

If authentication fails:
1. Make sure you're using a valid Steam account that owns Rust
2. Check that you have allowed popups in your browser
3. Try running the command again

### Notification Issues

If you're not receiving notifications:
1. Verify that your FCM registration was successful
2. Check that your Steam account is properly linked with Rust+
3. Ensure that the server you're trying to receive notifications from has paired with your Rust+ app

## Development

### Project Structure

```
rustplus-cli/
├── cmd/
│   └── rustplus/       # Main CLI application
├── fcm/                # FCM client implementation
│   ├── android/        # Android-specific FCM implementation
│   ├── client/         # FCM client
│   ├── constants/      # Constants used throughout the project
│   ├── crypto/         # Cryptography utilities
│   ├── gcm/            # GCM implementation
│   ├── parser/         # Message parser
│   ├── proto/          # Protocol buffer definitions
│   └── serverkey/      # Server key storage
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
└── README.md           # This file
```

### Running Tests

```
go test ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Facepunch Studios](https://facepunch.com/) for creating Rust and the Rust+ companion app
- [push-receiver](https://github.com/MatthieuLemoine/push-receiver) for the original Node.js implementation
- The Rust community for their support and feedback

## Disclaimer

This project is not affiliated with or endorsed by Facepunch Studios. Use at your own risk.
