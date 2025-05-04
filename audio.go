package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	speech "cloud.google.com/go/speech/apiv1"
	"cloud.google.com/go/speech/apiv1/speechpb"
	"cloud.google.com/go/storage"
)

// HandleAudioToText processes an audio file and generates a transcript
func HandleAudioToText(audioPath, transcriptPath string) error {
	// Check if audio file exists
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return fmt.Errorf("audio file not found: %s", audioPath)
	}

	// Check if file is MP3
	if filepath.Ext(audioPath) != ".mp3" {
		return fmt.Errorf("only .mp3 files are supported for audio transcription")
	}

	// Set default transcript path if not provided
	if transcriptPath == "" {
		transcriptPath = filepath.Join(filepath.Dir(audioPath), filepath.Base(audioPath)[:len(filepath.Base(audioPath))-len(filepath.Ext(audioPath))]+".md")
	}

	// Initialize Google Cloud clients
	ctx := context.Background()

	// Create Speech client
	speechClient, err := speech.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create speech client: %v", err)
	}
	defer speechClient.Close()

	// Create Storage client
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}
	defer storageClient.Close()

	// Upload audio file to GCS
	bucketName := "platformlabs-audios"
	objectName := fmt.Sprintf("audio-%d.mp3", time.Now().Unix())

	// Open the audio file
	audioFile, err := os.Open(audioPath)
	if err != nil {
		return fmt.Errorf("failed to open audio file: %v", err)
	}
	defer audioFile.Close()

	// Upload to GCS
	bucket := storageClient.Bucket(bucketName)
	bucket.Attrs(ctx) // This will ensure the bucket exists and is in the correct region
	obj := bucket.Object(objectName)
	wc := obj.NewWriter(ctx)
	if _, err := io.Copy(wc, audioFile); err != nil {
		return fmt.Errorf("failed to upload to GCS: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %v", err)
	}

	// Configure the request with GCS URI
	gcsURI := fmt.Sprintf("gs://%s/%s", bucketName, objectName)
	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        speechpb.RecognitionConfig_MP3,
			SampleRateHertz: 16000,
			LanguageCode:    "pt-BR",
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Uri{Uri: gcsURI},
		},
	}

	// Perform the transcription
	op, err := speechClient.LongRunningRecognize(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to start transcription: %v", err)
	}

	resp, err := op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to complete transcription: %v", err)
	}

	// Generate markdown content
	var sb strings.Builder
	sb.WriteString("# Audio Transcription\n\n")
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			sb.WriteString(alt.Transcript + "\n\n")
		}
	}

	// Write to markdown file
	if err := ioutil.WriteFile(transcriptPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write transcript: %v", err)
	}

	// Clean up: delete the temporary GCS object
	if err := obj.Delete(ctx); err != nil {
		log.Printf("Warning: Failed to delete temporary GCS object: %v", err)
	}

	fmt.Printf("Transcription completed successfully! Output saved to %s\n", transcriptPath)
	return nil
}
