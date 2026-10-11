package mirror

import (
	"fmt"
	"strings"
)

// NpmManager implements PackageManager for npm
type NpmManager struct {
	*BaseManager
}

// NewNpmManager creates a new npm manager
func NewNpmManager() *NpmManager {
	return &NpmManager{
		BaseManager: NewBaseManager("npm", GetNpmConfigPath()),
	}
}

// Detect detects the current npm configuration
func (n *NpmManager) Detect() ConfigInfo {
	// Check if config file exists
	if !FileExists(n.configPath) {
		return ConfigInfo{
			Status:     StatusDefault,
			ConfigFile: n.configPath,
		}
	}

	// Read config file
	content, err := ReadFile(n.configPath)
	if err != nil {
		return ConfigInfo{
			Status:       StatusError,
			ConfigFile:   n.configPath,
			ErrorMessage: err.Error(),
		}
	}

	// Check if registry is configured
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "registry") {
			// Extract registry URL
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				url := strings.TrimSpace(parts[1])
				// Check if it's a Chinese mirror
				if isChineseMirror(url) {
					return ConfigInfo{
						Status:     StatusMirror,
						ConfigFile: n.configPath,
						CurrentURL: url,
					}
				}
				// Non-Chinese registry
				return ConfigInfo{
					Status:     StatusDefault,
					ConfigFile: n.configPath,
					CurrentURL: url,
				}
			}
		}
	}

	return ConfigInfo{
		Status:     StatusDefault,
		ConfigFile: n.configPath,
	}
}

// Configure configures npm to use the specified mirror
func (n *NpmManager) Configure(mirror MirrorConfig) error {
	// Backup current config
	if err := Backup(n.configPath, n.name); err != nil {
		return fmt.Errorf("failed to backup config: %w", err)
	}

	// Read existing config
	var content string
	if FileExists(n.configPath) {
		content, _ = ReadFile(n.configPath)
	}

	// Build new config
	lines := strings.Split(content, "\n")
	newLines := []string{}
	registryFound := false

	// Parse existing config
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "registry") {
			// Replace existing registry
			newLines = append(newLines, fmt.Sprintf("registry=%s", mirror.URL))
			registryFound = true
		} else {
			// Keep other lines
			newLines = append(newLines, line)
		}
	}

	// Add registry if not found
	if !registryFound {
		newLines = append(newLines, fmt.Sprintf("registry=%s", mirror.URL))
	}

	// Write config
	newContent := strings.Join(newLines, "\n")
	if err := WriteFile(n.configPath, newContent); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Restore restores the configuration from backup
func (n *NpmManager) Restore() error {
	return Restore(n.configPath, n.name)
}
