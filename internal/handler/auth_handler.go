package handler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"PishingSimulator_SecurityProject/internal/auth"
	"PishingSimulator_SecurityProject/internal/storage"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// var users = make(map[string]models.User)

type SignupRequest struct {
	Username string `json:"username" example:"new_user"`
	Password string `json:"password" example:"password123"`
}

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

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	/*
		if _, exist := users[credentials.Username]; exist {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
			return
		} */

	if credentials.Password == "" || credentials.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and Password cannot be empty"})
		return
	}

	HashedPassword, err := bcrypt.GenerateFromPassword([]byte(credentials.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	/* 회원 정보 검증용, 메모리에 저장
	users[credentials.Username] = models.User{
		Username:     credentials.Username,
		PasswordHash: string(HashedPassword),
	} */

	if err := storage.CreateUser(credentials.Username, string(HashedPassword)); err != nil {
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

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	/*
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		user, exist := users[credentials.Username]
		if !exist {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		tokenString, err := auth.GenerateToken(credentials.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
	*/

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
