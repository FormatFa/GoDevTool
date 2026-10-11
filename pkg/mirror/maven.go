package mirror

import (
	"fmt"
	"strings"
)

// MavenManager implements PackageManager for Maven
type MavenManager struct {
	*BaseManager
}

// NewMavenManager creates a new Maven manager
func NewMavenManager() *MavenManager {
	return &MavenManager{
		BaseManager: NewBaseManager("maven", GetMavenConfigPath()),
	}
}

// Detect detects the current Maven configuration
func (m *MavenManager) Detect() ConfigInfo {
	// Check if config file exists
	if !FileExists(m.configPath) {
		return ConfigInfo{
			Status:     StatusDefault,
			ConfigFile: m.configPath,
		}
	}

	// Read config file
	content, err := ReadFile(m.configPath)
	if err != nil {
		return ConfigInfo{
			Status:       StatusError,
			ConfigFile:   m.configPath,
			ErrorMessage: err.Error(),
		}
	}

	// Check if mirror is configured
	if strings.Contains(content, "<mirror>") {
		// Try to extract mirror URL
		if strings.Contains(content, "<url>") {
			start := strings.Index(content, "<url>")
			if start != -1 {
				start += 5
				end := strings.Index(content[start:], "</url>")
				if end != -1 {
					url := strings.TrimSpace(content[start : start+end])
					return ConfigInfo{
						Status:     StatusMirror,
						ConfigFile: m.configPath,
						CurrentURL: url,
					}
				}
			}
		}
		return ConfigInfo{
			Status:     StatusMirror,
			ConfigFile: m.configPath,
		}
	}

	return ConfigInfo{
		Status:     StatusDefault,
		ConfigFile: m.configPath,
	}
}

// Configure configures Maven to use the specified mirror
func (m *MavenManager) Configure(mirror MirrorConfig) error {
	// Backup current config
	if err := Backup(m.configPath, m.name); err != nil {
		return fmt.Errorf("failed to backup config: %w", err)
	}

	// Read existing config
	content := ""
	if FileExists(m.configPath) {
		content, _ = ReadFile(m.configPath)
	}

	// Build mirror configuration
	mirrorConfig := fmt.Sprintf(`
    <mirror>
      <id>%s</id>
      <mirrorOf>*</mirrorOf>
      <name>%s Maven Mirror</name>
      <url>%s</url>
    </mirror>`, mirror.Name, mirror.Name, mirror.URL)

	// Check if mirrors section exists
	if strings.Contains(content, "<mirrors>") {
		// Replace existing mirrors section
		start := strings.Index(content, "<mirrors>")
		if start != -1 {
			start += len("<mirrors>")
			end := strings.Index(content[start:], "</mirrors>")
			if end != -1 {
				// Replace content between <mirrors> and </mirrors>
				newContent := content[:start] + mirrorConfig + content[start+end:]
				if err := WriteFile(m.configPath, newContent); err != nil {
					return fmt.Errorf("failed to write config: %w", err)
				}
				return nil
			}
		}
	}

	// If no mirrors section exists, we need to add it
	// This is a simplified version - a full implementation would handle XML properly
	// For now, append to file
	newContent := content
	if !strings.Contains(content, "</settings>") {
		// Create new settings file
		newContent = `<?xml version="1.0" encoding="UTF-8"?>
<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0"
          xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
          xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0
          http://maven.apache.org/xsd/settings-1.0.0.xsd">
  <mirrors>` + mirrorConfig + `
  </mirrors>
</settings>`
	} else {
		// Insert before </settings>
		newContent = strings.Replace(content,
			"</settings>",
			"  <mirrors>"+mirrorConfig+`
  </mirrors>
</settings>`, 1)
	}

	if err := WriteFile(m.configPath, newContent); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Restore restores the configuration from backup
func (m *MavenManager) Restore() error {
	return Restore(m.configPath, m.name)
}
