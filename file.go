package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IsDirectory checks if the given path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetFilesInDirectory returns all files with the specified extensions in a directory and its subdirectories
func GetFilesInDirectory(directory string, extensions []string) ([]string, error) {
	var files []string

	// Check if directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory not found: %s", directory)
	}

	// Walk through the directory
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check file extension
		for _, ext := range extensions {
			if strings.EqualFold(filepath.Ext(path), ext) {
				files = append(files, path)
				break
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return files, nil
}

// EnsureDirectoryExists creates a directory if it doesn't exist
func EnsureDirectoryExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// GetDefaultOutputPath generates a default output path for the given input file
func GetDefaultOutputPath(inputPath, outputDir, newExt string) string {
	// If output directory is not specified, use the same directory as the input file
	if outputDir == "" {
		return filepath.Join(
			filepath.Dir(inputPath),
			filepath.Base(inputPath)[:len(filepath.Base(inputPath))-len(filepath.Ext(inputPath))]+newExt,
		)
	}

	// Ensure output directory exists
	if err := EnsureDirectoryExists(outputDir); err != nil {
		// If we can't create the directory, fall back to input directory
		return filepath.Join(
			filepath.Dir(inputPath),
			filepath.Base(inputPath)[:len(filepath.Base(inputPath))-len(filepath.Ext(inputPath))]+newExt,
		)
	}

	// Use the output directory with the base name of the input file
	return filepath.Join(
		outputDir,
		filepath.Base(inputPath)[:len(filepath.Base(inputPath))-len(filepath.Ext(inputPath))]+newExt,
	)
}
