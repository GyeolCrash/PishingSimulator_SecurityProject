package handler

import (
	"log"

	"github.com/gorilla/websocket"
)

func manageTextSession(conn *websocket.Conn, username string) {
	defer conn.Close()
	log.Printf("Text session started for user: %s", username)

ReadLoop:
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from user %s: %v", username, err)
			break ReadLoop
		}

		if messageType != websocket.TextMessage {
			log.Printf("Unsupported message type from user %s: %d", username, messageType)
			continue
		} else {
			log.Printf("Received text message from user %s: %s", username, string(message))
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error sending message to user %s: %v", username, err)
				break ReadLoop
			}
		}
	}
	log.Printf("Text session ended for user: %s", username)
}
