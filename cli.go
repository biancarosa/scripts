package main

import (
	"flag"
	"log"
)

// RunCLI handles the command-line interface
func RunCLI() {
	// Define command line flags
	videoPath := flag.String("video", "", "Path to the input video file")
	outputPath := flag.String("output", "", "Path for the output audio file (optional)")
	audioPath := flag.String("audio", "", "Path to the input audio file for transcription")
	transcriptPath := flag.String("transcript", "", "Path for the output transcript file (optional)")
	flag.Parse()

	// Handle video to audio conversion
	if *videoPath != "" {
		if err := HandleVideoToAudio(*videoPath, *outputPath); err != nil {
			log.Fatal(err)
		}
	}

	// Handle audio to text transcription
	if *audioPath != "" {
		if err := HandleAudioToText(*audioPath, *transcriptPath); err != nil {
			log.Fatal(err)
		}
	}

	// If no valid flags are provided
	if *videoPath == "" && *audioPath == "" {
		log.Fatal("Please provide either -video or -audio flag")
	}
}
