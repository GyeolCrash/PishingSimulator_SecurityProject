package handler

import (
	"FishingSimulator_SecurityProject/internal/auth"
	"FishingSimulator_SecurityProject/internal/simulation"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleSimulation(c *gin.Context) {
	tokenString := c.Query("token")
	scenarioKey := c.Query("scenario")

	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	username := claims.Username
	log.Printf("User %s connected with scenario key: %s", username, scenarioKey)

	scenario, exists := simulation.GetScenario(scenarioKey)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scenario key"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("error: Failed to upgrade to WebSocket : User %s with %v", username, err)
		return
	}

	defer conn.Close()
	log.Printf("WebSocket connection established for user: %s", username)

	initalMessage := fmt.Sprintf("Start Secnario %s: %s", scenario.Name, scenario.Description)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(initalMessage)); err != nil {
		log.Printf("Error sending message to user %s: %v", username, err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from user %s: %v", username, err)
			break
		}
		log.Printf("Received message from user %s: %s", username, message)
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Error sending message to user %s: %v", username, err)
			break
		}
	}
	log.Printf("WebSocket connection closed for user: %s", username)
}
