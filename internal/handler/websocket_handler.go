package handler

import (
	"PishingSimulator_SecurityProject/internal/auth"
	"PishingSimulator_SecurityProject/internal/models"
	"PishingSimulator_SecurityProject/internal/storage"
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

// HandleSimulationConnection godoc
// @Summary      보이스피싱 시뮬레이션 시작 (WebSocket)
// @Description  지정된 시나리오와 모드로 실시간 시뮬레이션을 위한 WebSocket 연결을 시작합니다.
// @Description  <br>
// @Description  **[중요]** 이것은 표준 HTTP API가 아닙니다. `ws://` 또는 `wss://` 스킴을 사용해야 합니다.
// @Description  **인증:** WebSocket 연결 시에는 HTTP Header를 사용할 수 없으므로, **Query Parameter(`token`)**로 JWT를 전달해야 합니다.
// @Tags         Simulation (WebSocket)
// @Accept       json
// @Produce      json
// @Param        token    query     string  true  "Bearer 토큰 (접두사 없이 토큰 값만 입력)"
// @Param        scenario query     string  true  "시나리오 키 (예: loan_scam, institution_impersonation)"
// @Param        mode     query     string  true  "모드 선택 (text: 텍스트 채팅, voice: 실시간 음성 통화)"
// @Success      101      {string}  string  "Switching Protocols"
// @Failure      400      {object}  map[string]string "잘못된 파라미터"
// @Failure      401      {object}  map[string]string "인증 실패"
// @Router       /ws/simulation [get]
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
	scenario, exists := models.GetScenario(scenarioKey)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scenario key"})
		return
	}
	if mode != "text" && mode != "voice" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mode"})
		return
	}

	user, err := storage.GetUserByUsername(username)
	if err != nil {
		log.Printf("HandleSimulationConnection(): Failed to get user info for websocket: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	log.Printf("User: %s, %d, %s, Scenario: %s, Mode: %s", user.Profile.Name, user.Profile.Age, user.Profile.Gender, scenario.Name, mode)

	// WebSocket 연결 업그레이드과 종료
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("error: Failed to upgrade to WebSocket : User %s with %v", username, err)
		return
	}
	conn.SetReadLimit(10485760) // DoS 방지용 최대 메시지 크기 제한, 10MB

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
		manageTextSession(conn, user, context.Background(), scenarioKey)
	case "voice":
		manageAudioSession(conn, user, context.Background(), scenarioKey)
	default:
		// add error handling for unsupported mode
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		log.Printf("Unsupported mode for user %s: %s", username, mode)
	}
}
