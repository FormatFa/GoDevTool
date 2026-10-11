package mirror

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Backup creates a backup of the configuration file
func Backup(configPath, managerName string) error {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No config file to backup
		return nil
	}

	// Ensure backup directory exists
	backupDir := GetBackupDir()
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create backup file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupFileName := fmt.Sprintf("%s-%s.bak", managerName, timestamp)
	backupPath := filepath.Join(backupDir, backupFileName)

	// Copy file to backup location
	if err := copyFile(configPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Clean up old backups (keep last 10)
	cleanupOldBackups(backupDir, managerName, 10)

	return nil
}

// Restore restores the configuration from the most recent backup
func Restore(configPath, managerName string) error {
	backupDir := GetBackupDir()
	backups, err := findBackups(backupDir, managerName)
	if err != nil {
		return fmt.Errorf("failed to find backups: %w", err)
	}

	if len(backups) == 0 {
		return fmt.Errorf("no backups found for %s", managerName)
	}

	// Get the most recent backup
	latestBackup := backups[len(backups)-1]

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Copy backup to config location
	if err := copyFile(latestBackup, configPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// findBackups finds all backup files for a package manager, sorted by date
func findBackups(backupDir, managerName string) ([]string, error) {
	var backups []string

	err := filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file matches backup pattern
		filename := info.Name()
		prefix := managerName + "-"
		if strings.HasPrefix(filename, prefix) && strings.HasSuffix(filename, ".bak") {
			backups = append(backups, path)
		}

		return nil
	})

	return backups, err
}

// cleanupOldBackups removes old backups, keeping only the most recent n
func cleanupOldBackups(backupDir, managerName string, keep int) {
	backups, err := findBackups(backupDir, managerName)
	if err != nil {
		return
	}

	// If we have more backups than we want to keep, remove the oldest ones
	if len(backups) > keep {
		for i := 0; i < len(backups)-keep; i++ {
			os.Remove(backups[i])
		}
	}
}

// ReadFile reads the contents of a file
func ReadFile(path string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// WriteFile writes content to a file, creating parent directories if needed
func WriteFile(path, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
