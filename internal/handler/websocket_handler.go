package handler

import (
	"PishingSimulator_SecurityProject/internal/auth"
	"PishingSimulator_SecurityProject/internal/simulation"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Upgrade HTTP connection to WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handle WebSocket connection for client-side simulation
func HandleSimulationConnection(c *gin.Context) {

	// URL Query 파라미터 추출
	tokenString := c.Query("token")
	scenarioKey := c.Query("scenario")
	mode := c.Query("mode")

	// 사용자 토큰 검증
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	username := claims.Username
	log.Printf("User %s connected with scenario key: %s", username, scenarioKey)

	// 시나리오와 모드 검증
	scenario, exists := simulation.GetScenario(scenarioKey)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scenario key"})
		return
	}
	if mode != "text" && mode != "voice" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mode"})
		return
	}
	log.Printf("User: %s, Scenario: %s, Mode: %s", username, scenario.Name, mode)

	// WebSocket 연결 업그레이드과 종료
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("error: Failed to upgrade to WebSocket : User %s with %v", username, err)
		return
	}
	defer conn.Close()
	log.Printf("WebSocket connection established for user: %s", username)

	// 초기 메시지 전송
	initalMessage := fmt.Sprintf("Start Secnario %s: %s", scenario.Name, scenario.Description)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(initalMessage)); err != nil {
		log.Printf("Error sending message to user %s: %v", username, err)
		return
	}

	// 모드에 따른 세션 관리
	switch mode {
	case "text":
		manageTextSession(conn, username)
	case "voice":
		manageAudioSession(conn, username, context.Background())
	}
}
