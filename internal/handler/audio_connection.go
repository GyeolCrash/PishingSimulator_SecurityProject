package handler

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

// Counter for received audio files saved during testing
var audioFileCounter atomic.Uint64
var mockAudioResponse []byte

const testAudioDir = "testdata/received_files"

// Make directory for audio test
func init() {
	if err := os.MkdirAll(testAudioDir, 0755); err != nil {
		log.Fatalf("Failed to create test audio directory: %v", err)
	}
	log.Printf("Test audio files will be stored in: %s", testAudioDir)
}

func manageAudioSession(conn *websocket.Conn, username string, parentCtx context.Context) {
	defer conn.Close()
	log.Printf("Audio session started for user: %s", username)

	// context
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	// WaitGroup for goroutines
	var wg sync.WaitGroup
	wg.Add(3)

	clientAudioInChannel := make(chan []byte, 128)
	clientAudioOutChannel := make(chan []byte, 128)

	go func() {
		defer wg.Done()
		defer cancel()
		clientReadPump(conn, username, clientAudioInChannel, ctx)
	}()

	go func() {
		defer wg.Done()
		defer cancel()
		clientWritePump(conn, username, clientAudioOutChannel, ctx)
	}()

	go func() {
		defer wg.Done()
		defer cancel()
		runTestSimulationLogic(username, clientAudioInChannel, clientAudioOutChannel, ctx)
	}()
	wg.Wait()
}

func clientReadPump(conn *websocket.Conn, username string, clientAudioInChannel chan<- []byte, ctx context.Context) {
	log.Printf("clientReadPump(): started for user: %s", username)
	defer close(clientAudioInChannel)

	for {
		select {
		case <-ctx.Done():
			log.Printf("clientReadPump(): Canceled with %s", username)
			return
		default:
		}
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("clientReadPump(): Error reading message from user %s: %v", username, err)
			return
		}

		if messageType != websocket.BinaryMessage {
			log.Printf("clientReadPump(): Unsupported message type from user %s: %d", username, messageType)
			continue
		} else {
			log.Printf("clientReadPump(): Received audio message from user %s: %d bytes", username, len(message))
			clientAudioInChannel <- message
		}

	}
}

func clientWritePump(conn *websocket.Conn, username string, clientAudioOutChannel <-chan []byte, ctx context.Context) {
	log.Printf("clientWritePump(): started for user: %s", username)
	for {
		select {
		case <-ctx.Done():
			log.Printf("clientWritePump(): %s", username)
			conn.WriteMessage(websocket.CloseMessage, []byte{})
			return

		case audioData, ok := <-clientAudioOutChannel:
			if !ok {
				log.Printf("clientWritePump(): audio out channel closed for user: %s", username)
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.WriteMessage(websocket.BinaryMessage, audioData); err != nil {
				log.Printf("clientWritePump(): Error sending audio to user %s: %v", username, err)
				return
			}
			log.Printf("clientWritePump(): Sent audio to user %s: %d bytes", username, len(audioData))
		}
	}
}

func runTestSimulationLogic(username string, clientAudioInChannel <-chan []byte, serverAudioOutChannel chan<- []byte, ctx context.Context) {
	log.Printf("runTestSimulationLogic(): started for user: %s", username)
	defer close(serverAudioOutChannel)

	for {
		select {
		case <-ctx.Done():
			log.Printf("runTestSimulationLogic(): Canceled with %s", username)
			return
		case audioData, ok := <-clientAudioInChannel:
			if !ok {
				log.Printf("runTestSimulationLogic(): client audio in channel closed for user: %s", username)
				return
			}
			handleReceiveAudio(audioData, username)

			if mockAudioResponse != nil {
				serverAudioOutChannel <- mockAudioResponse
			}
		}
	}
}

func handleReceiveAudio(message []byte, username string) {
	count := audioFileCounter.Add(1)
	fileName := fmt.Sprintf("%s_audio_%d.wav", username, count)
	filePath := filepath.Join(testAudioDir, fileName)
	if err := os.WriteFile(filePath, message, 0644); err != nil {
		log.Printf("handleReceiveAudio(): Failed to save audio file for user %s: %v", username, err)
	} else {
		log.Printf("handleReceiveAudio(): Saved audio file for user %s: %s", username, filePath)
	}
}
