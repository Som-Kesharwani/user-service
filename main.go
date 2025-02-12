package main

import (
	"flag"
	"time"

	_ "github.com/Som-Kesharwani/shared-service/database"
	"github.com/Som-Kesharwani/shared-service/logger"
	_ "github.com/Som-Kesharwani/shared-service/logger"
	"github.com/Som-Kesharwani/user-service/routers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	port := *flag.String("port", "8080", "Application Port")
	flag.Parse()
	if port == "" {
		logger.Error.Println("Provide Correct Application Port !!")
	}
	router := gin.New()
	router.Use(gin.Logger())

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Allow your frontend origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, // Allow cookies to be sent
		MaxAge:           12 * time.Hour,
	}))

	routers.UserRouter(router)

	router.Run(":" + port)

}
