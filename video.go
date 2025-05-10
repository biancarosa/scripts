package main

import (
	"fmt"
	"os"
	"os/exec"
)

// HandleVideoToAudio extracts audio from a video file or all video files in a directory
func HandleVideoToAudio(videoPath, outputPath string) error {
	// Check if path is a directory
	if IsDirectory(videoPath) {
		return handleVideoDirectoryToAudio(videoPath, outputPath)
	}

	// Handle single video file
	return handleSingleVideoToAudio(videoPath, outputPath)
}

// handleSingleVideoToAudio extracts audio from a single video file
func handleSingleVideoToAudio(videoPath, outputPath string) error {
	// Check if video file exists
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return fmt.Errorf("video file not found: %s", videoPath)
	}

	// Set default output path if not provided
	if outputPath == "" {
		outputPath = GetDefaultOutputPath(videoPath, "", ".mp3")
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

	fmt.Printf("Audio extraction completed successfully: %s\n", outputPath)
	return nil
}

// handleVideoDirectoryToAudio extracts audio from all video files in a directory
func handleVideoDirectoryToAudio(directoryPath, outputDir string) error {
	// Get all video files in the directory
	videoFiles, err := GetFilesInDirectory(directoryPath, []string{".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv"})
	if err != nil {
		return err
	}

	if len(videoFiles) == 0 {
		return fmt.Errorf("no video files found in directory: %s", directoryPath)
	}

	// Process each video file
	for _, videoFile := range videoFiles {
		// Generate output path
		var audioOutputPath string
		if outputDir != "" {
			audioOutputPath = GetDefaultOutputPath(videoFile, outputDir, ".mp3")
		} else {
			audioOutputPath = ""
		}

		// Process the video file
		if err := handleSingleVideoToAudio(videoFile, audioOutputPath); err != nil {
			fmt.Printf("Error processing %s: %v\n", videoFile, err)
			// Continue with next file
			continue
		}
	}

	return nil
}
