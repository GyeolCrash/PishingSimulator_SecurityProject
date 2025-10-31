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

func Signup(c *gin.Context) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

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

func Login(c *gin.Context) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

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

func Profile(c *gin.Context) {
	username, _ := c.Get("username")
	c.JSON(http.StatusOK, gin.H{"message": "this is a protected profile", "username": username})
}
