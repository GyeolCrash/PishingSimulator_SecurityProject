package handler

import (
	"PishingSimulator_SecurityProject/internal/storage"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// GetCallHistory godoc
// @Summary      사용자 통화 기록 조회
// @Description  사용자의 과거 시뮬레이션(통화/채팅) 기록 목록을 최신순으로 반환합니다.
// @Tags         History
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  map[string][]models.Recording "history: [기록 배열]"
// @Failure      401      {object}  map[string]string "인증 실패"
// @Failure      500      {object}  map[string]string "서버 내부 오류"
// @Router       /api/history [get]
func GetCallHistory(c *gin.Context) {
	username := c.GetString("username")

	userID, err := storage.GetUserIDByUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	records, err := storage.GetRecordsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch records"})
	}
	c.JSON(http.StatusOK, gin.H{"history": records})
}

// StreamAudio godoc
// @Summary      녹음된 오디오 파일 재생 (스트리밍)
// @Description  특정 통화 기록의 오디오 파일(.mp3)을 스트리밍합니다.
// @Description  <br>
// @Description  **인증 방법:**
// @Description  1. **Header:** `Authorization: Bearer {token}` (안드로이드 권장)
// @Description  2. **Query:** `?token={token}` (웹/HTML 오디오 태그용)
// @Tags         History
// @Produce      audio/mpeg
// @Security     BearerAuth
// @Param        filename path      string  true  "오디오 파일명 (예: session_uuid.mp3)"
// @Param        token    query     string  false "JWT 토큰 (헤더 사용 시 생략 가능)"
// @Success      200      {file}    file    "오디오 파일 스트림"
// @Failure      401      {object}  map[string]string "인증 실패"
// @Failure      404      {object}  map[string]string "파일을 찾을 수 없음"
// @Router       /api/audio/{filename} [get]
func StreamAudio(c *gin.Context) {
	username := c.GetString("username")
	filename := c.Param("filename")

	filePath := filepath.Join("data", "records", username, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audio file not found"})
		return
	}

	c.File(filePath)
}
