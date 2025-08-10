// Package fsutil provides filesystem utility functions
package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Default permissions for directories and files
const (
	DirPerm  = 0755 // rwxr-xr-x
	FilePerm = 0644 // rw-r--r--
)

// SafeMkdir creates a directory if it doesn't exist
func SafeMkdir(path string, perm os.FileMode) error {
	// Check if directory already exists
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", path)
		}
		// Directory already exists, that's fine
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check directory: %w", err)
	}

	// Create the directory with all parent directories
	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}

	return nil
}

// SafeCreateFile creates a file if it doesn't exist, returns error if it does
func SafeCreateFile(path string, perm os.FileMode) (*os.File, error) {
	// Use O_EXCL to fail if file already exists
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("file already exists: %s", path)
		}
		return nil, fmt.Errorf("failed to create file %s: %w", path, err)
	}
	return file, nil
}

// PathExists checks if a path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if the path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile checks if the path is a regular file
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

// SanitizePath ensures the path is safe and doesn't escape the target directory
func SanitizePath(basePath, userPath string) (string, error) {
	// Clean the user path first
	cleanUserPath := filepath.Clean(userPath)

	// Check for absolute paths
	if filepath.IsAbs(cleanUserPath) {
		return "", fmt.Errorf("absolute paths not allowed: %s", userPath)
	}

	// Check for path traversal attempts
	// Only reject if path contains "../" or "/.." or is exactly ".."
	if cleanUserPath == ".." || strings.Contains(cleanUserPath, "../") || strings.Contains(cleanUserPath, "/..") || strings.HasSuffix(cleanUserPath, "/..") {
		return "", fmt.Errorf("path traversal not allowed: %s", userPath)
	}

	// Join with base path and clean again
	fullPath := filepath.Join(basePath, cleanUserPath)
	cleanFullPath := filepath.Clean(fullPath)

	// Convert to absolute paths for comparison
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base path: %w", err)
	}

	absPath, err := filepath.Abs(cleanFullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve full path: %w", err)
	}

	// Ensure the resolved path is within the base directory
	if !strings.HasPrefix(absPath, absBase) {
		return "", fmt.Errorf("path escapes target directory: %s", userPath)
	}

	return cleanFullPath, nil
}

// EnsureDir ensures all parent directories exist for a given path
func EnsureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "/" {
		return nil
	}
	return SafeMkdir(dir, DirPerm)
}

// NormalizePath normalizes a path for the current OS
func NormalizePath(path string) string {
	// Replace forward slashes with OS-specific separator
	normalized := filepath.FromSlash(path)
	// Clean the path
	return filepath.Clean(normalized)
}
