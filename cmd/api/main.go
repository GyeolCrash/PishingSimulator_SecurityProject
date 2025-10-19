package main

import (
	"FishingSimulator_SecurityProject/internal/handler"
	"FishingSimulator_SecurityProject/internal/middleware"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = append(config.AllowHeaders, "Authorization")
	router.Use(cors.New(config))

	router.POST("/signup", handler.Signup)
	router.POST("/login", handler.Login)

	protected := router.Group("/api").Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", handler.Profile)
	}

	router.GET("/ws/simulation", handler.HandleSimulation)
	log.Fatal(router.Run(":8080"))
}
