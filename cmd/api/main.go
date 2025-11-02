package main

import (
	"PishingSimulator_SecurityProject/internal/handler"
	"PishingSimulator_SecurityProject/internal/middleware"
	"PishingSimulator_SecurityProject/internal/storage"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "PishingSimulator_SecurityProject/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Phising Simulator API
// @version 						0.1
// @description
// @host 							localhost:8080
// @BasePath 						/
// @securityDefinitions.apikey		BearerAuth
// @in header
// @name Authorization
// @description Bearer 토큰 형식, Bearer {token}

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

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler((swaggerFiles.Handler)))
	log.Fatal(router.Run(":8080"))
}
