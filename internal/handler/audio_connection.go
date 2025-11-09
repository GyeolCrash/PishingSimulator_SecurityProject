package handler

import (
	"PishingSimulator_SecurityProject/internal/llm"
	"PishingSimulator_SecurityProject/internal/models"

	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
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

func manageAudioSession(conn *websocket.Conn, user models.User, parentCtx context.Context, scenarioKey string) {
	defer conn.Close()
	log.Printf("Audio session started for user: %s", user.Username)

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
		clientReadPump(conn, user.Username, clientAudioInChannel, ctx)
	}()

	go func() {
		defer wg.Done()
		defer cancel()
		clientWritePump(conn, user.Username, clientAudioOutChannel, ctx)
	}()

	go func() {
		defer wg.Done()
		defer cancel()
		orchestrateAudioSession(user, scenarioKey, clientAudioInChannel, clientAudioOutChannel, ctx)
		//runBinaryEchoLogic(username, clientAudioInChannel, clientAudioOutChannel, ctx)
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

func orchestrateAudioSession(user models.User, scenarioKey string, clientAudioInChannel <-chan []byte, serverAudioOutChannel chan<- []byte, ctx context.Context) {

	llmSessionID := uuid.New().String()
	username := user.Username
	log.Printf("orchestrateAudioSession(): started for user: %s, session: %s", username, llmSessionID)
	defer close(serverAudioOutChannel)

	// STT 스트리밍 인식기 생성
	sttRecognizer, err := llm.NewStreamingRecognizer(ctx)
	if err != nil {
		log.Printf("orchestrateAudioSession(): Failed to create STT recognizer for user %s: %v", username, err)
		return
	}

	ttsCleint, err := llm.NewTTSClient(ctx)
	if err != nil {
		log.Printf("orchestrateAudioSession(): Failed to create TTS client for user %s: %v", username, err)
		// 이미 생성된 STT Recognizer 종료
		sttRecognizer.Close()
		return
	}

	// LLM 세션 종료 시 Recongizer, TTS 클라이언트 종료
	defer func() {
		log.Printf("orchestrateAudioSession(): Ending LLM session %s for user: %s", llmSessionID, username)

		if err := sttRecognizer.Close(); err != nil {
			log.Printf("orchestrateAudioSession(): Failed to close STT recognizer for user %s: %v", username, err)
		}
		if err := ttsCleint.Close(); err != nil {
			log.Printf("orchestrateAudioSession(): Failed to close TTS client for user %s: %v", username, err)
		}

		log.Printf("orchestrateAudioSession(): LLM session %s ended for user: %s", llmSessionID, username)
		llm.ClearSession(llmSessionID)
		close(serverAudioOutChannel)
	}()

	// LLM 세션 초기화
	initalUtterance, err := llm.InitSession(llmSessionID, scenarioKey, user.Profile)
	if err != nil {
		log.Printf("orchestrateAudioSession(): Failed to initialize LLM session for user %s: %v", username, err)
		return
	}
	log.Printf("orchestrateAudioSession(): LLM initial utterance for user %s: %s", username, initalUtterance)

	responseAudio, err := ttsCleint.ConvertTextToAudio(initalUtterance)
	if err != nil {
		log.Printf("orchestrateAudioSession(): TTS conversion failed for user %s: %v", username, err)
		return
	}
	serverAudioOutChannel <- responseAudio

	sttResultChannel := make(chan string, 10)
	sttErrorChannel := make(chan error, 1)

	go sttRecognizer.ReceiveTranslatedText(sttResultChannel, sttErrorChannel)

	for {
		select {
		case <-ctx.Done(): // 세션 취소 및 정리
			log.Printf("orchestrateAudioSession(): Canceled with %s", username)
			return
		case audioChunk, ok := <-clientAudioInChannel: // 클라이언트 오디오 수신
			if !ok {
				log.Printf("orchestrateAudioSession(): client audio in channel closed for user: %s", username)
				return
			}

			// Todo: C->S 아카이빙 로직 추가
			// handleReceiveAudio(audioChunk, username)

			if err := sttRecognizer.SendAudio(audioChunk); err != nil {
				log.Printf("orchestrateAudioSession(): STT send audio failed for user %s: %v", username, err)
			}

		case userText := <-sttResultChannel: // is final 텍스트 수신
			log.Printf("orchestrateAudioSession(): STT [FINAL] %s with %s", userText, username)
			chatResp, err := llm.Chat(llmSessionID, userText)
			if err != nil {
				log.Printf("orchestrateAudioSession(): LLM chat failed for user %s: %v", username, err)
				continue
			}

			// Todo: S->C 아카이빙 로직 추가

			log.Printf("orchestrateAudioSession(): LLM response for user %s: %s", username, chatResp.Utterance)
			responseAudio, err := ttsCleint.ConvertTextToAudio(chatResp.Utterance)
			if err != nil {
				log.Printf("orchestrateAudioSession(): TTS conversion failed for user %s: %v", username, err)
				continue
			}

			serverAudioOutChannel <- responseAudio

		case err := <-sttErrorChannel: // 치명적 오류 수신
			log.Printf("orchestrateAudioSession(): STT error for user %s: %v", username, err)
			return
		}
	}
}

// 오디오 파일 저장
func handleReceiveAudio(message []byte, username string) string {
	count := audioFileCounter.Add(1)
	fileName := fmt.Sprintf("%s_audio_%d.wav", username, count)
	filePath := filepath.Join(testAudioDir, fileName)
	if err := os.WriteFile(filePath, message, 0644); err != nil {
		log.Printf("handleReceiveAudio(): Failed to save audio file for user %s: %v", username, err)
	} else {
		log.Printf("handleReceiveAudio(): Saved audio file for user %s: %s", username, filePath)
	}
	return fmt.Sprintf("Received audio chunk %d (%d bytes)", count, len(message))
}

// 오디오 에코, 저장
// 제대로 동작하지 않는다면 handleReceiveAudio() 주석 처리
func runBinaryEchoLogic(username string, clientAudioInChannel <-chan []byte, serverAudioOutChannel chan<- []byte, ctx context.Context) {
	log.Printf("runBinaryEchoLogic(): started for user: %s", username)
	defer close(serverAudioOutChannel)

	for {
		select {
		case <-ctx.Done():
			log.Printf("runBinaryEchoLogic(): Canceled with %s", username)
			return
		case audioData, ok := <-clientAudioInChannel:
			if !ok {
				log.Printf("runBinaryEchoLogic(): client audio in channel closed for user: %s", username)
				return
			}
			handleReceiveAudio(audioData, username)
			if mockAudioResponse != nil {
				serverAudioOutChannel <- mockAudioResponse
			}
			log.Printf("runBinaryEchoLogic(): Echoing audio data for user: %s, %d bytes", username, len(audioData))
			serverAudioOutChannel <- audioData
		}

	}
}
