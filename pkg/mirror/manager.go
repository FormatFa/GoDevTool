package mirror

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigStatus represents the status of a package manager configuration
type ConfigStatus string

const (
	StatusNotInstalled ConfigStatus = "未安装"
	StatusDefault      ConfigStatus = "默认配置"
	StatusMirror       ConfigStatus = "已配置镜像"
	StatusError        ConfigStatus = "检测错误"
)

// MirrorConfig represents a mirror source configuration
type MirrorConfig struct {
	Name string // Mirror name (e.g., "阿里云", "腾讯云")
	URL  string // Mirror URL
}

// ConfigInfo represents the detected configuration information
type ConfigInfo struct {
	Status       ConfigStatus // Current configuration status
	ConfigFile   string       // Configuration file path
	CurrentURL   string       // Current mirror URL (if configured)
	ErrorMessage string       // Error message (if StatusError)
}

// PackageManager defines the interface for package manager mirror configuration
type PackageManager interface {
	// Name returns the package manager name (e.g., "pip", "npm", "maven")
	Name() string

	// Detect detects the current configuration status
	Detect() ConfigInfo

	// Configure configures the package manager to use the specified mirror
	Configure(mirror MirrorConfig) error

	// Backup creates a backup of the current configuration
	Backup() error

	// Restore restores the configuration from the most recent backup
	Restore() error

	// GetConfigPath returns the configuration file path
	GetConfigPath() string
}

// BaseManager provides common functionality for package managers
type BaseManager struct {
	name       string
	configPath string
}

// NewBaseManager creates a new base manager
func NewBaseManager(name, configPath string) *BaseManager {
	return &BaseManager{
		name:       name,
		configPath: configPath,
	}
}

// Name returns the package manager name
func (b *BaseManager) Name() string {
	return b.name
}

// GetConfigPath returns the configuration file path
func (b *BaseManager) GetConfigPath() string {
	return b.configPath
}

// EnsureConfigDir ensures the configuration directory exists
func EnsureConfigDir(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return nil
}
