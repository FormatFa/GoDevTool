package mirror

import (
	"fmt"
	"strings"
)

// PipManager implements PackageManager for pip
type PipManager struct {
	*BaseManager
}

// NewPipManager creates a new pip manager
func NewPipManager() *PipManager {
	return &PipManager{
		BaseManager: NewBaseManager("pip", GetPipConfigPath()),
	}
}

// Detect detects the current pip configuration
func (p *PipManager) Detect() ConfigInfo {
	// Check if pip is installed (optional, could check with "pip --version")

	// Check if config file exists
	if !FileExists(p.configPath) {
		return ConfigInfo{
			Status:     StatusDefault,
			ConfigFile: p.configPath,
		}
	}

	// Read config file
	content, err := ReadFile(p.configPath)
	if err != nil {
		return ConfigInfo{
			Status:       StatusError,
			ConfigFile:   p.configPath,
			ErrorMessage: err.Error(),
		}
	}

	// Check if mirror is configured
	if strings.Contains(content, "index-url") {
		// Extract current mirror URL
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "index-url") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					url := strings.TrimSpace(parts[1])
					// Check if it's a Chinese mirror
					if isChineseMirror(url) {
						return ConfigInfo{
							Status:     StatusMirror,
							ConfigFile: p.configPath,
							CurrentURL: url,
						}
					}
				}
			}
		}
	}

	return ConfigInfo{
		Status:     StatusDefault,
		ConfigFile: p.configPath,
	}
}

// Configure configures pip to use the specified mirror
func (p *PipManager) Configure(mirror MirrorConfig) error {
	// Backup current config
	if err := Backup(p.configPath, p.name); err != nil {
		return fmt.Errorf("failed to backup config: %w", err)
	}

	// Read existing config
	var content string
	if FileExists(p.configPath) {
		content, _ = ReadFile(p.configPath)
	}

	// Build new config
	lines := strings.Split(content, "\n")
	newLines := []string{}
	indexURLFound := false

	// Parse existing config
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "index-url") {
			// Replace existing index-url
			newLines = append(newLines, fmt.Sprintf("index-url = %s", mirror.URL))
			indexURLFound = true
		} else if !strings.HasPrefix(trimmed, "trusted-host") {
			// Keep other lines (except old trusted-host)
			newLines = append(newLines, line)
		}
	}

	// Add index-url if not found
	if !indexURLFound {
		newLines = append(newLines, fmt.Sprintf("index-url = %s", mirror.URL))
	}

	// Add trusted-host for the mirror
	// Extract host from URL
	urlHost := extractHost(mirror.URL)
	if urlHost != "" {
		newLines = append(newLines, fmt.Sprintf("trusted-host = %s", urlHost))
	}

	// Write config
	newContent := strings.Join(newLines, "\n")
	if err := WriteFile(p.configPath, newContent); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Restore restores the configuration from backup
func (p *PipManager) Restore() error {
	return Restore(p.configPath, p.name)
}

// isChineseMirror checks if a URL is a Chinese mirror
func isChineseMirror(url string) bool {
	chineseMirrors := []string{
		"aliyun.com", "aliyun.cn",
		"tencent.com", "qq.com",
		"huaweicloud.com",
		"tsinghua.edu.cn",
		"tuna.tsinghua.edu.cn",
	}

	for _, mirror := range chineseMirrors {
		if strings.Contains(url, mirror) {
			return true
		}
	}
	return false
}

// extractHost extracts the host from a URL
func extractHost(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.")

	// Remove path
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}

	return url
}
