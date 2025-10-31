package main

import (
	"PishingSimulator_SecurityProject/internal/handler"
	"PishingSimulator_SecurityProject/internal/middleware"
	"PishingSimulator_SecurityProject/internal/storage"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	storage.InitDB()
	router := gin.Default()

	// CORS 설정
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true // 경고: 실제 배포 시에는 특정 도메인으로 제한해야 함
	config.AllowHeaders = append(config.AllowHeaders, "Authorization")
	router.Use(cors.New(config))

	// 라우트 설정
	router.POST("/signup", handler.Signup)
	router.POST("/login", handler.Login)

	// 보호된 라우트 그룹
	protected := router.Group("/api").Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", handler.Profile)
	}

	// WebSocket 핸들러
	router.GET("/ws/simulation", handler.HandleSimulationConnection)
	log.Fatal(router.Run(":8080"))
}
