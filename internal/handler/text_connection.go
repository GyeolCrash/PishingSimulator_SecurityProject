package handler

import (
	"PishingSimulator_SecurityProject/internal/models"
	"context"
	"log"

	"github.com/gorilla/websocket"
)

func manageTextSession(conn *websocket.Conn, user models.User, parentCtx context.Context, scenarioKey string) {
	defer conn.Close()
	log.Printf("Text session started for user: %s", user.Username)

	// llm.InitSession 호출 등 세션 초기화 로직 추가 가능

ReadLoop:
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from user %s: %v", user.Username, err)
			break ReadLoop
		}

		if messageType != websocket.TextMessage {
			log.Printf("Unsupported message type from user %s: %d", user.Username, messageType)
			continue
		} else {
			log.Printf("Received text message from user %s: %s", user.Username, string(message))
			// Todo: llm.Chat 호출 등 메시지 처리 로직 추가 가능

			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error sending message to user %s: %v", user.Username, err)
				break ReadLoop
			}
		}
	}
	log.Printf("Text session ended for user: %s", user.Username)
}
