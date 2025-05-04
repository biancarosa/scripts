package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// HandleVideoToAudio extracts audio from a video file
func HandleVideoToAudio(videoPath, outputPath string) error {
	// Check if video file exists
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return fmt.Errorf("video file not found: %s", videoPath)
	}

	// Set default output path if not provided
	if outputPath == "" {
		outputPath = filepath.Join(filepath.Dir(videoPath), filepath.Base(videoPath)[:len(filepath.Base(videoPath))-len(filepath.Ext(videoPath))]+".mp3")
	}

	// Check if ffmpeg is installed
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg is not installed. Please install it first")
	}

	// Execute ffmpeg command
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vn", "-acodec", "libmp3lame", "-ab", "192k", outputPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	fmt.Printf("Extracting audio from %s to %s...\n", videoPath, outputPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error extracting audio: %v", err)
	}

	fmt.Printf("Audio extraction completed successfully!\n")
	return nil
}
