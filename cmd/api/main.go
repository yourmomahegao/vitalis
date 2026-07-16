package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"vitalis/internal/database"
	"vitalis/internal/enviroment"
	"vitalis/internal/handlers"
	"vitalis/internal/tasks"
)

func main() {
	// ========== DOTENV ==========
	enviroment.Preload()

	// ========== TASKS ==========
	tasks.CompileTasks()

	// ========== SQLITE INITIALIZATION ==========
	err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		return
	}

	err = database.Initialize()
	if err != nil {
		log.Fatalf("Failed to intialize PostgreSQL: %v", err)
		return
	}

	// ========== GIN INITIALIZATION ==========
	ginEngine := gin.Default()

	// ========== GIN URLS ==========
	ginEngine.GET("/auth/token/", handlers.AccessToken)
	ginEngine.GET("/auth/token/check/", handlers.AccessTokenCheck)
	ginEngine.POST("/worker/status/", handlers.WorkerStatus)

	// ========== GIN RUN ==========
	if err := ginEngine.Run(enviroment.Env.RunAddress); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
