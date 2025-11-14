package handler

import (
	"PishingSimulator_SecurityProject/internal/archiver"
	"PishingSimulator_SecurityProject/internal/llm"
	"PishingSimulator_SecurityProject/internal/models"
	"strings"
	"sync"
	"time"

	"context"
	"log"

	"github.com/google/uuid"
)

func orchestrateAudioSession(user models.User, scenarioKey string, clientChannel <-chan []byte, serverChannel chan<- []byte, ctx context.Context) {

	llmSessionID := uuid.New().String()
	username := user.Username
	log.Printf("orchestrateAudioSession(): started for user: %s, session: %s", username, llmSessionID)
	defer close(serverChannel)

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
		close(serverChannel)
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
	serverChannel <- responseAudio

	sttResultChannel := make(chan string, 10)
	sttErrorChannel := make(chan error, 1)

	go sttRecognizer.ReceiveTranslatedText(sttResultChannel, sttErrorChannel)

	for {
		select {
		case <-ctx.Done(): // 세션 취소 및 정리
			log.Printf("orchestrateAudioSession(): Canceled with %s", username)
			return
		case audioChunk, ok := <-clientChannel: // 클라이언트 오디오 수신
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

			serverChannel <- responseAudio

		case err := <-sttErrorChannel: // 치명적 오류 수신
			log.Printf("orchestrateAudioSession(): STT error for user %s: %v", username, err)
			return
		}
	}
}

func orchestrateVoiceEchoTest(user models.User, scenarioKey string, sessionStartTime time.Time,
	clientChan <-chan []byte, serverChan chan<- []byte, archiveC2SChan chan<- []byte,
	archiveS2CChan chan<- archiver.ArchiveS2CJob, sessionContext context.Context) {

	username := user.Username
	log.Printf("orchestrateVoiceSession(): [STT->TTS Echo Test Mode] started for user: %s", username)

	// WritePump 및 Archiver(G4) 고루틴에 종료 신호 전송
	defer close(serverChan)
	defer close(archiveC2SChan)
	defer close(archiveS2CChan)

	sttRecognizer, err := llm.NewStreamingRecognizer(sessionContext)
	if err != nil { /* ... (에러 처리) ... */
		return
	}

	ttsClient, err := llm.NewTTSClient(sessionContext)
	if err != nil { /* ... (에러 처리) ... */
		return
	}

	defer func() {
		log.Printf("orchestrateVoiceSession(): Cleaning up resources for %s", username)
		if err := sttRecognizer.Close(); err != nil {
			log.Printf("orchestrateVoiceSession(): Error closing STT: %v", err)
		}
		if err := ttsClient.Close(); err != nil {
			log.Printf("orchestrateVoiceSession(): Error closing TTS: %v", err)
		}
	}()

	var isListening = true
	var stateMutex sync.Mutex
	var lastFinalText string = ""

	// 첫 인사말 전송을 위한 Goroutine (TTS / 아카이빙 / 전송)
	go func() {
		initialUtterance := "STT TTS 에코 테스트 모드입니다. 좀 돼라 망할."
		log.Printf("orchestrateVoiceSession(): Sending initial test message...")
		responseAudio, err := ttsClient.ConvertTextToAudio(initialUtterance)
		if err == nil {
			startTime := time.Since(sessionStartTime)

			archiveS2CChan <- archiver.ArchiveS2CJob{Data: responseAudio, StartTime: startTime}
			serverChan <- responseAudio
		} else {
			log.Printf("orchestrateVoiceSession(): Failed to convert initial TTS: %v", err)
		}
	}()

	sttResultChan := make(chan string, 10)
	sttErrChan := make(chan error, 1)

	// STT 응답 수신을 위한 별도 Goroutine 시작
	go sttRecognizer.ReceiveTranslatedText(sttResultChan, sttErrChan) // (ReceiveTranscribedText)

	// 메인 처리 루프 (State Machine)
	for {
		select {
		case <-sessionContext.Done():
			log.Printf("orchestrateVoiceSession(): Canceled with %s", username)
			return

		case audioChunk, ok := <-clientChan:
			if !ok {
				log.Printf("orchestrateVoiceSession(): client audio in channel closed for user: %s", username)
				return
			}

			archiveC2SChan <- audioChunk

			stateMutex.Lock()
			currentListeningState := isListening
			stateMutex.Unlock()

			if currentListeningState {
				if err := sttRecognizer.SendAudio(audioChunk); err != nil {
					log.Printf("Failed to send audio to STT: %v", err)
				}
			}

		case userText := <-sttResultChan:
			sttFinalTime := time.Since(sessionStartTime)
			cleanedText := strings.TrimSpace(userText)

			stateMutex.Lock()
			if !isListening || cleanedText == "" || cleanedText == lastFinalText {
				continue
			}

			isListening = false
			lastFinalText = userText
			stateMutex.Unlock()
			log.Printf("orchestrateVoiceSession(): STT [FINAL] -> %s", userText)
			log.Printf("... (State change: NOW RESPONDING. Discarding audio input)")

			// [수정] TTS 변환 및 전송을 *별도 Goroutine*에서 처리
			go func(textToSpeak string, sttTimestamp time.Duration) {
				log.Printf("orchestrateVoiceSession(): Calling TTS for: %s", textToSpeak)
				responseAudio, err := ttsClient.ConvertTextToAudio(textToSpeak)

				if err != nil {
					log.Printf("orchestrateVoiceSession(): Failed to convert chat TTS: %v", err)
				} else {
					archiveS2CChan <- archiver.ArchiveS2CJob{Data: responseAudio, StartTime: sttTimestamp} // S->C 아카이빙
					serverChan <- responseAudio                                                            // S->C 전송
				}

				stateMutex.Lock()
				isListening = true
				log.Printf("... (State change: NOW LISTENING)")
				stateMutex.Unlock()

			}(cleanedText, sttFinalTime)

		case err := <-sttErrChan:
			log.Printf("orchestrateVoiceSession(): STT stream error: %v", err)
			return
		}
	}
}

func archiveAudioConversation(username string, archiver *archiver.Archiver, c2sIn <-chan []byte,
	s2cIn <-chan archiver.ArchiveS2CJob, ctx context.Context) {

	log.Printf("archiveAudioConversation(): started for user: %s", username)
	defer archiver.CloseBaseTrack()

	for {
		select {
		case <-ctx.Done():
			log.Printf("runArchivingLogic(): Canceled for %s", username)
			return
		case chunk := <-c2sIn:
			/*
				if !ok {
					// archiver.WriteC2S(chunk)
					return
				}*/
			archiver.WriteC2S(chunk)
		case job, ok := <-s2cIn:
			if !ok {
				return
			}
			archiver.WriteS2C(job)
		}
	}

}
