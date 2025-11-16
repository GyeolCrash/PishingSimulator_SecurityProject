/**
* Name: 			auth_handler.go
* Description: 		Gin 프레임워크의 HTTP 핸들러
* Workflow: 		회원가입, 로그인, 프로필 조회
 */
package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"PishingSimulator_SecurityProject/internal/auth"
	"PishingSimulator_SecurityProject/internal/models"
	"PishingSimulator_SecurityProject/internal/storage"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

/*
// [테스트용 코드]
var users = make(map[string]models.User)
*/

// /Signup 요청 바디
type SignupRequest struct {
	Username string             `json:"username" example:"new_user"`
	Password string             `json:"password" example:"password123"`
	Profile  models.UserProfile `json:"profile"`
}

// /Login 요청 바디
type LoginRequest struct {
	Username string `json:"username" example:"my_user"`
	Password string `json:"password" example:"password123"`
}

type SuccessResponse struct {
	Message string `json:"message" example:"User created successfully"`
}
type ErrorResponse struct {
	Error string `json:"error" example:"Username already exists"`
}
type LoginSuccessResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// Signup godoc
// @Summary      회원가입 (Signup)
// @Description  새로운 사용자 계정을 생성합니다.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body handler.SignupRequest true "회원가입 요청 정보"
// @Success      200 {object} handler.SuccessResponse
// @Failure      400 {object} handler.ErrorResponse
// @Failure      500 {object} handler.ErrorResponse
// @Router       /signup [post]
func Signup(c *gin.Context) {
	var credentials SignupRequest

	// sqlite 드라이버와 ShouldBindJSON의 호환성 문제로 인한 우회 코드
	rawData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := json.Unmarshal(rawData, &credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// " "으로 입력되는 케이스 방지
	if strings.TrimSpace(credentials.Username) == "" || strings.TrimSpace(credentials.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and Password cannot be empty"})
		return
	}
	if credentials.Profile.Age <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Age must be a positive number"})
		return
	}

	// password 해싱
	HashedPassword, err := bcrypt.GenerateFromPassword([]byte(credentials.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	// DB에 사용자 생성
	if err := storage.CreateUser(credentials.Username, string(HashedPassword), credentials.Profile); err != nil {
		if errors.Is(err, storage.ErrUsernameExists) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		} else {
			log.Printf("[ERROR] Failed to create user (database error): %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user (database error)"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})

}

// Login godoc
// @Summary      로그인 (Login)
// @Description  사용자명과 비밀번호로 로그인하고 JWT 토큰을 발급받습니다.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body handler.LoginRequest true "로그인 요청 정보"
// @Success      200 {object} handler.LoginSuccessResponse
// @Failure      400 {object} handler.ErrorResponse "잘못된 요청"
// @Failure      401 {object} handler.ErrorResponse "인증 실패 (자격 증명 오류)"
// @Failure      500 {object} handler.ErrorResponse "서버 내부 오류"
// @Router       /login [post]
func Login(c *gin.Context) {
	var credentials LoginRequest

	rawData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}
	if err := json.Unmarshal(rawData, &credentials); err != nil {
		log.Printf("[ERROR] Login: json.Unmarshal failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON parsing error: " + err.Error()})
		return
	}

	if credentials.Username == "" || credentials.Password == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	user, err := storage.GetUserByUsername(credentials.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		log.Printf("[ERROR] GetUserByUsername failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	tokenString, err := auth.GenerateToken(credentials.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// Profile godoc
// @Summary      프로필 조회 (Profile)
// @Description  인증된 사용자의 프로필 정보를 조회합니다. (JWT 필요)
// @Tags         API (Protected)
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} object{message=string, username=string}
// @Failure      401 {object} handler.ErrorResponse "인증 토큰 누락 또는 만료"
// @Router       /api/profile [get]
func Profile(c *gin.Context) {
	username, _ := c.Get("username")
	c.JSON(http.StatusOK, gin.H{"message": "this is a protected profile", "username": username})
}
