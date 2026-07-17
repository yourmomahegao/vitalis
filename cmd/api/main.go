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

	if enviroment.ENV.GIN_DEBUG {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// ========== GIN URLS ==========
	ginEngine.GET("/worker/status/", handlers.WorkerStatus)

	// ========== GIN URLS AUTH ==========
	ginEngine.GET("/auth/token/", handlers.AccessToken)
	ginEngine.GET("/auth/token/check/", handlers.AccessTokenCheck)

	// ========== GIN URLS SYSTEM ==========
	ginEngine.POST("/info/system/cpu/", handlers.CpuInformation)
	ginEngine.POST("/info/system/ram/", handlers.RamInformation)
	ginEngine.POST("/info/system/net/", handlers.NetInformation)
	ginEngine.POST("/info/system/file/", handlers.FileInformation)

	// ========== GIN RUN ==========
	if err := ginEngine.Run(enviroment.ENV.RUN_ADDRESS); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
