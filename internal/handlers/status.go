package handlers

import (
	"net/http"

	"vitalis/internal/handlers/structs"

	"github.com/gin-gonic/gin"
)

func WorkerStatus(c *gin.Context) {
	response := structs.Response{
		Status:  true,
		Message: "All systems online",
	}

	c.JSON(http.StatusOK, response)
}
