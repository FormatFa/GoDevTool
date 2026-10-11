package mirror

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetPipConfigPath returns the pip configuration file path for the current platform
func GetPipConfigPath() string {
	homeDir, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\pip\pip.ini
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "pip", "pip.ini")
		}
		return filepath.Join(homeDir, "AppData", "Roaming", "pip", "pip.ini")
	case "darwin":
		// macOS: ~/Library/Application Support/pip/pip.conf
		return filepath.Join(homeDir, "Library", "Application Support", "pip", "pip.conf")
	default: // linux, etc.
		// Linux: ~/.config/pip/pip.conf or /etc/pip.conf
		configPath := filepath.Join(homeDir, ".config", "pip", "pip.conf")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
		return "/etc/pip.conf"
	}
}

// GetMavenConfigPath returns the Maven configuration file path for the current platform
func GetMavenConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	// Maven uses the same path on all platforms: ~/.m2/settings.xml
	return filepath.Join(homeDir, ".m2", "settings.xml")
}

// GetNpmConfigPath returns the npm configuration file path for the current platform
func GetNpmConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	// npm uses the same path on all platforms: ~/.npmrc
	return filepath.Join(homeDir, ".npmrc")
}

// GetAppDataDir returns the application data directory for storing backups
func GetAppDataDir() string {
	homeDir, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\godevkit\
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "godevkit")
		}
		return filepath.Join(homeDir, "AppData", "Roaming", "godevkit")
	case "darwin":
		// macOS: ~/Library/Application Support/godevkit/
		return filepath.Join(homeDir, "Library", "Application Support", "godevkit")
	default: // linux, etc.
		// Linux: ~/.config/godevkit/
		return filepath.Join(homeDir, ".config", "godevkit")
	}
}

// GetBackupDir returns the backup directory for package manager configurations
func GetBackupDir() string {
	return filepath.Join(GetAppDataDir(), "backups")
}
