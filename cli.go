package main

import (
	"flag"
	"fmt"
	"log"
)

// RunCLI handles the command-line interface
func RunCLI() {
	// Define command line flags
	videoPath := flag.String("video", "", "Path to the input video file or directory")
	outputPath := flag.String("output", "", "Path for the output audio file or directory (optional)")
	audioPath := flag.String("audio", "", "Path to the input audio file or directory for transcription")
	transcriptPath := flag.String("transcript", "", "Path for the output transcript file or directory (optional)")

	// Add help text
	flag.Usage = func() {
		fmt.Println("Media Processing Tool - Convert videos to audio and transcribe audio files")
		fmt.Println("\nUsage:")
		fmt.Println("  Process a single video file:       go run *.go -video path/to/video.mp4 [-output path/to/audio.mp3]")
		fmt.Println("  Process all videos in a directory: go run *.go -video path/to/videos/ [-output path/to/audios/]")
		fmt.Println("  Transcribe a single audio file:    go run *.go -audio path/to/audio.mp3 [-transcript path/to/transcript.md]")
		fmt.Println("  Transcribe all audios in a dir:    go run *.go -audio path/to/audios/ [-transcript path/to/transcripts/]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Handle video to audio conversion
	if *videoPath != "" {
		fmt.Printf("Processing video(s) from: %s\n", *videoPath)
		if IsDirectory(*videoPath) {
			fmt.Println("Directory mode: Processing all video files in the directory")
		}

		if err := HandleVideoToAudio(*videoPath, *outputPath); err != nil {
			log.Fatal(err)
		}
	}

	// Handle audio to text transcription
	if *audioPath != "" {
		fmt.Printf("Transcribing audio(s) from: %s\n", *audioPath)
		if IsDirectory(*audioPath) {
			fmt.Println("Directory mode: Processing all audio files in the directory")
		}

		if err := HandleAudioToText(*audioPath, *transcriptPath); err != nil {
			log.Fatal(err)
		}
	}

	// If no valid flags are provided
	if *videoPath == "" && *audioPath == "" {
		flag.Usage()
		log.Fatal("Please provide either -video or -audio flag")
	}
}
