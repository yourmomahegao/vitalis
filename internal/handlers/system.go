package handlers

import (
	"net/http"

	"vitalis/internal/handlers/structs"

	"github.com/gin-gonic/gin"
)

func CpuInformation(c *gin.Context) {
	tokenStatus := CheckToken(c)
	if tokenStatus == false {
		return
	}

	c.JSON(http.StatusOK, structs.Response{
		Status:  true,
		Message: "Access token valid",
	})
}

func RamInformation(c *gin.Context) {
	tokenStatus := CheckToken(c)
	if tokenStatus == false {
		return
	}

	c.JSON(http.StatusOK, structs.Response{
		Status:  true,
		Message: "Access token valid",
	})
}

func NetInformation(c *gin.Context) {
	tokenStatus := CheckToken(c)
	if tokenStatus == false {
		return
	}

	c.JSON(http.StatusOK, structs.Response{
		Status:  true,
		Message: "Access token valid",
	})
}

func FileInformation(c *gin.Context) {
	tokenStatus := CheckToken(c)
	if tokenStatus == false {
		return
	}

	c.JSON(http.StatusOK, structs.Response{
		Status:  true,
		Message: "Access token valid",
	})
}
