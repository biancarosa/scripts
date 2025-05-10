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

// HandleAudioToText processes an audio file or all audio files in a directory and generates transcripts
func HandleAudioToText(audioPath, transcriptPath string) error {
	// Check if path is a directory
	if IsDirectory(audioPath) {
		return handleAudioDirectoryToText(audioPath, transcriptPath)
	}

	// Handle single audio file
	return handleSingleAudioToText(audioPath, transcriptPath)
}

// handleSingleAudioToText processes a single audio file and generates a transcript
func handleSingleAudioToText(audioPath, transcriptPath string) error {
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
		transcriptPath = GetDefaultOutputPath(audioPath, "", ".md")
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
			Encoding:                            speechpb.RecognitionConfig_MP3,
			SampleRateHertz:                     16000,
			LanguageCode:                        "pt-BR",
			EnableAutomaticPunctuation:          true,
			UseEnhanced:                         true,
			Model:                               "latest_long",
			AudioChannelCount:                   1,
			EnableWordTimeOffsets:               true,
			EnableWordConfidence:                true,
			ProfanityFilter:                     false,
			EnableSeparateRecognitionPerChannel: false,
			DiarizationConfig: &speechpb.SpeakerDiarizationConfig{
				EnableSpeakerDiarization: false,
			},
			Adaptation: &speechpb.SpeechAdaptation{
				PhraseSets: []*speechpb.PhraseSet{
					{
						Phrases: []*speechpb.PhraseSet_Phrase{
							{Value: "git push"},
							{Value: "git pull"},
							{Value: "git commit"},
							{Value: "git clone"},
							{Value: "git status"},
						},
						Boost: 20,
					},
				},
			},
			Metadata: &speechpb.RecognitionMetadata{
				InteractionType:          speechpb.RecognitionMetadata_PRESENTATION,
				IndustryNaicsCodeOfAudio: 611420, // Computer training
				OriginalMediaType:        speechpb.RecognitionMetadata_VIDEO,
				RecordingDeviceType:      speechpb.RecognitionMetadata_PC,
				RecordingDeviceName:      "Video Course",
				OriginalMimeType:         "audio/mp3",
			},
			SpeechContexts: []*speechpb.SpeechContext{
				{
					Phrases: []string{
						// Comandos Git e ferramentas de controle de versão
						"git push", "git pull", "git commit", "git clone", "git checkout", "git branch",
						"git merge", "git rebase", "git status", "git add", "git init", "git log",
						"pull request", "merge request", "GitHub", "GitLab", "Bitbucket", "git fetch",

						// Linguagem Go e termos específicos
						"Go", "Golang", "struct", "interface", "goroutine", "channel", "defer",
						"slice", "map", "package", "import", "function", "método", "ponteiro",
						"concorrência", "API", "HTTP", "JSON", "variável", "constante",
						"erro", "servidor", "cliente", "biblioteca", "framework",

						// Ferramentas de desenvolvimento
						"VSCode", "IDE", "terminal", "compilador", "runtime", "Docker", "Kubernetes",
						"CLI", "command line", "bash", "shell", "terminal", "console",

						// Conceitos de programação
						"código", "programação", "desenvolvimento", "aplicação", "debug", "debugar",
						"bug", "feature", "deploy", "deployment", "CI/CD", "pipeline",

						// Outros termos técnicos em inglês
						"string", "integer", "float", "boolean", "array", "hash", "null", "nil",
						"pointer", "reference", "stack", "heap", "memory", "buffer", "cache",
						"async", "sync", "thread", "process", "request", "response",
					},
					Boost: 20, // Aumentar o boost para priorizar o reconhecimento desses termos técnicos
				},
				{
					// Comandos Git específicos com boost muito alto para evitar erros de reconhecimento
					Phrases: []string{
						"git push", "git pull", "git commit", "git clone",
						"git status", "git add", "git init", "git log",
						"git checkout", "git branch", "git merge", "git rebase",
					},
					Boost: 30, // Boost muito alto para comandos Git específicos
				},
			},
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

// handleAudioDirectoryToText processes all audio files in a directory and generates transcripts
func handleAudioDirectoryToText(directoryPath, outputDir string) error {
	// Get all audio files in the directory
	audioFiles, err := GetFilesInDirectory(directoryPath, []string{".mp3"})
	if err != nil {
		return err
	}

	if len(audioFiles) == 0 {
		return fmt.Errorf("no MP3 files found in directory: %s", directoryPath)
	}

	// Process each audio file
	for _, audioFile := range audioFiles {
		// Generate output path
		var transcriptOutputPath string
		if outputDir != "" {
			transcriptOutputPath = GetDefaultOutputPath(audioFile, outputDir, ".md")
		} else {
			transcriptOutputPath = ""
		}

		// Process the audio file
		if err := handleSingleAudioToText(audioFile, transcriptOutputPath); err != nil {
			fmt.Printf("Error processing %s: %v\n", audioFile, err)
			// Continue with next file
			continue
		}
	}

	return nil
}
