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

// HandleSimulation godoc
// @Summary      시뮬레이션 WebSocket 연결
// @Description  지정된 시나리오와 모드로 실시간 시뮬레이션을 위한 WebSocket 연결을 시작합니다.
// @Description  <br>
// @Description  **참고: 이것은 표준 HTTP API가 아닙니다.**
// @Description  클라이언트는 `ws://` 또는 `wss://` 스킴을 사용하여 이 엔드포인트에 연결해야 합니다.
// @Description  인증은 HTTP Header가 아닌 **쿼리 파라미터('token')**를 통해 수행됩니다.
// @Tags         WebSocket (Simulation)
// @Param        token    query     string  true  "로그인 시 발급받은 JWT 토큰"
// @Param        scenario query     string  true  "시나리오 키 (예: loan_scam)"
// @Param        mode     query     string  true  "시뮬레이션 모드 (text 또는 voice)"
// @Success      101      {string}  string  "101 Switching Protocols (WebSocket으로 프로토콜 전환 성공)"
// @Failure      400      {object}  handler.ErrorResponse "잘못된 파라미터 (scenario, mode)"
// @Failure      401      {object}  handler.ErrorResponse "토큰 누락 또는 유효하지 않은 토큰"
// @Failure      500      {object}  handler.ErrorResponse "WebSocket 업그레이드 실패"
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
		manageTextSession(conn, user, context.Background(), scenarioKey)
	case "voice":
		manageAudioSession(conn, user, context.Background(), scenarioKey)
	default:
		// add error handling for unsupported mode
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		log.Printf("Unsupported mode for user %s: %s", username, mode)
	}
}
