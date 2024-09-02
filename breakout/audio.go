package main

import (
	"embed"
	"io"
	"log"
	"time"

	"github.com/braheezy/qoa"
	"github.com/ebitengine/oto/v3"
)

// Embed all sound files from the sounds/ directory
//
//go:embed sounds/*
var soundFiles embed.FS

var audioContext *oto.Context

func initAudio() {
	// Prepare an Oto context (this will use the default audio device)
	ctx, ready, err := oto.NewContext(
		&oto.NewContextOptions{
			// Typically 44100 or 48000, we could get it from a QOA file but we'd have to decode one.
			SampleRate: 44100,
			// only 1 or 2 are supported by oto
			ChannelCount: 2,
			// QOA is always 16 bit
			Format: oto.FormatSignedInt16LE,
		})
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}

	// Wait for the audio context to be ready
	<-ready
	audioContext = ctx
}

func playAudioOnLoop(song string) {
	qoaBytes, err := soundFiles.ReadFile(song)
	if err != nil {
		log.Fatalf("Error reading QOA file: %v", err)
	}
	qoaMetadata, qoaAudioData, err := qoa.Decode(qoaBytes)
	if err != nil {
		log.Fatalf("Error decoding QOA data: %v", err)
	}

	reader := qoa.NewReader(qoaAudioData, int(qoaMetadata.Channels))
	player := audioContext.NewPlayer(reader)

	go func() {
		for {
			player.Play()
			for {
				if !player.IsPlaying() {
					// Rewind the song to the beginning
					reader.Seek(0, io.SeekStart)
					break
				}
				// Sleep briefly to avoid busy-waiting
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func playAudioOnce(song string) {
	qoaBytes, err := soundFiles.ReadFile(song)
	if err != nil {
		log.Fatalf("Error reading QOA file: %v", err)
	}
	qoaMetadata, qoaAudioData, err := qoa.Decode(qoaBytes)
	if err != nil {
		log.Fatalf("Error decoding QOA data: %v", err)
	}

	reader := qoa.NewReader(qoaAudioData, int(qoaMetadata.Channels))
	player := audioContext.NewPlayer(reader)

	player.Play()
}
