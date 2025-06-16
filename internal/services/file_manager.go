package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/francknouama/image-recognition-webapp/internal/config"
	"github.com/sirupsen/logrus"
)

// FileManager handles file operations and cleanup
type FileManager struct {
	config      *config.Config
	logger      *logrus.Logger
	tempDir     string
	uploadsDir  string
	cleanupAge  time.Duration
}

// NewFileManager creates a new file manager
func NewFileManager(cfg *config.Config) (*FileManager, error) {
	tempDir := "./temp"
	uploadsDir := "./uploads"
	cleanupAge := 24 * time.Hour // Default: clean files older than 24 hours

	// Create directories if they don't exist
	if err := os.MkdirAll(tempDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	if err := os.MkdirAll(uploadsDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create uploads directory: %w", err)
	}

	return &FileManager{
		config:     cfg,
		logger:     logrus.New(),
		tempDir:    tempDir,
		uploadsDir: uploadsDir,
		cleanupAge: cleanupAge,
	}, nil
}

// SetCleanupAge sets the age threshold for cleanup
func (fm *FileManager) SetCleanupAge(age time.Duration) {
	fm.cleanupAge = age
}

// CleanupTempFiles removes old temporary files
func (fm *FileManager) CleanupTempFiles() error {
	return fm.cleanupDirectory(fm.tempDir)
}

// CleanupUploads removes old uploaded files
func (fm *FileManager) CleanupUploads() error {
	return fm.cleanupDirectory(fm.uploadsDir)
}

// CleanupAll performs cleanup on all managed directories
func (fm *FileManager) CleanupAll() error {
	var lastErr error

	if err := fm.CleanupTempFiles(); err != nil {
		fm.logger.Errorf("Failed to cleanup temp files: %v", err)
		lastErr = err
	}

	if err := fm.CleanupUploads(); err != nil {
		fm.logger.Errorf("Failed to cleanup uploads: %v", err)
		lastErr = err
	}

	return lastErr
}

// cleanupDirectory removes files older than cleanupAge from a directory
func (fm *FileManager) cleanupDirectory(dir string) error {
	cutoff := time.Now().Add(-fm.cleanupAge)
	
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is older than cutoff
		if info.ModTime().Before(cutoff) {
			fm.logger.Infof("Removing old file: %s (age: %v)", path, time.Since(info.ModTime()))
			
			if err := os.Remove(path); err != nil {
				fm.logger.Errorf("Failed to remove file %s: %v", path, err)
				return err
			}
		}

		return nil
	})
}

// StartPeriodicCleanup starts a background cleanup routine
func (fm *FileManager) StartPeriodicCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	
	go func() {
		defer ticker.Stop()
		
		for range ticker.C {
			fm.logger.Debug("Running periodic cleanup")
			if err := fm.CleanupAll(); err != nil {
				fm.logger.Errorf("Periodic cleanup failed: %v", err)
			}
		}
	}()
	
	fm.logger.Infof("Started periodic cleanup with interval: %v", interval)
}

// GetTempDir returns the temporary directory path
func (fm *FileManager) GetTempDir() string {
	return fm.tempDir
}

// GetUploadsDir returns the uploads directory path
func (fm *FileManager) GetUploadsDir() string {
	return fm.uploadsDir
}

// CreateTempFile creates a temporary file and returns its path
func (fm *FileManager) CreateTempFile(prefix string) (*os.File, error) {
	return os.CreateTemp(fm.tempDir, prefix)
}

// EnsureDirectories creates all necessary directories
func (fm *FileManager) EnsureDirectories() error {
	dirs := []string{
		fm.tempDir,
		fm.uploadsDir,
		"./logs",
		"./models",
		"./cache",
		"./cache/models",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return err
		}
		fm.logger.Debugf("Ensured directory exists: %s", dir)
	}

	return nil
}

// GetDirectorySize calculates the total size of files in a directory
func (fm *FileManager) GetDirectorySize(dir string) (int64, error) {
	var size int64

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			size += info.Size()
		}

		return nil
	})

	return size, err
}

// GetDirectoryStats returns statistics about a directory
func (fm *FileManager) GetDirectoryStats(dir string) (DirectoryStats, error) {
	stats := DirectoryStats{
		Path: dir,
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			stats.Directories++
		} else {
			stats.Files++
			stats.TotalSize += info.Size()
			
			if stats.OldestFile.IsZero() || info.ModTime().Before(stats.OldestFile) {
				stats.OldestFile = info.ModTime()
			}
			
			if stats.NewestFile.IsZero() || info.ModTime().After(stats.NewestFile) {
				stats.NewestFile = info.ModTime()
			}
		}

		return nil
	})

	return stats, err
}

// DirectoryStats contains statistics about a directory
type DirectoryStats struct {
	Path        string    `json:"path"`
	Files       int       `json:"files"`
	Directories int       `json:"directories"`
	TotalSize   int64     `json:"total_size"`
	OldestFile  time.Time `json:"oldest_file"`
	NewestFile  time.Time `json:"newest_file"`
}