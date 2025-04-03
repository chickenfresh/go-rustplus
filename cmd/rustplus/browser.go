package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// openBrowser opens Chrome with special flags for cross-origin access
func openBrowser(url string) error {
	var cmd *exec.Cmd

	// Chrome/Chromium flags to bypass security restrictions
	chromeFlags := []string{
		"--disable-web-security",                                     // allows us to manipulate rust+ window
		"--disable-popup-blocking",                                   // allows us to open rust+ login from our main window
		"--disable-site-isolation-trials",                            // required for --disable-web-security to work
		"--user-data-dir=/tmp/temporary-chrome-profile-dir-rustplus", // create a new chrome profile
	}

	// Determine Chrome executable path based on OS
	var chromePath string
	switch runtime.GOOS {
	case "windows":
		// Try common Chrome installation paths on Windows
		possiblePaths := []string{
			filepath.Join(os.Getenv("ProgramFiles"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("LocalAppData"), "Google", "Chrome", "Application", "chrome.exe"),
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				chromePath = path
				break
			}
		}

	case "darwin":
		// macOS Chrome path
		possiblePaths := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			filepath.Join(os.Getenv("HOME"), "Applications", "Google Chrome.app", "Contents", "MacOS", "Google Chrome"),
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				chromePath = path
				break
			}
		}

	default:
		// Linux and other OS paths
		possiblePaths := []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				chromePath = path
				break
			}
		}
	}

	if chromePath == "" {
		return fmt.Errorf("could not find Chrome/Chromium browser")
	}

	// Append the URL and create the command
	args := append(chromeFlags, url)
	cmd = exec.Command(chromePath, args...)

	// Start the browser
	return cmd.Start()
}
