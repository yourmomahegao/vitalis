package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"vitalis/internal/enviroment"
	"vitalis/internal/handlers/structs"
	"vitalis/internal/services"

	"github.com/gin-gonic/gin"
)

func CheckToken(c *gin.Context) bool {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Error occured in CheckToken(): %v", r)
			c.JSON(http.StatusBadRequest, structs.Response{
				Status:  false,
				Message: "No authorization token present",
			})
			return
		}
	}()

	bearerToken := c.Request.Header.Get("Authorization")
	reqToken := strings.Split(bearerToken, " ")[1]
	tokenValid, err := services.CheckSessionKey(reqToken)

	if err != nil {
		c.JSON(http.StatusBadRequest, structs.Response{
			Status:  false,
			Message: "Error while processing authorization token",
		})

		log.Printf("Error occured in CheckToken(): %v", err)
		return false
	}

	if tokenValid == false {
		c.JSON(http.StatusUnauthorized, structs.Response{
			Status:  false,
			Message: "Invalid access token",
		})

		return false
	}

	return true
}

type AccessTokenData struct {
	AccessToken string    `json:"access_token"`
	ValidUntil  time.Time `json:"valid_until"`
}

func AccessToken(c *gin.Context) {
	secretKey := c.PostForm("secret_key")

	if secretKey == "" {
		c.JSON(http.StatusBadRequest, structs.Response{
			Status:  false,
			Message: "secret_key parameter is not present",
		})
		return
	}

	if secretKey != enviroment.ENV.SECRET_KEY {
		c.JSON(http.StatusUnauthorized, structs.Response{
			Status:  false,
			Message: "Secret key is invalid",
		})
		return
	}

	newAccessToken, err := services.GenerateSessionKey()

	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.Response{
			Status:  false,
			Message: "Error while session key generating",
		})
		return
	}

	c.JSON(http.StatusOK, structs.Response{
		Status:  true,
		Message: "Generated new access-token",
		Data: AccessTokenData{
			AccessToken: newAccessToken,
			ValidUntil:  time.Now(),
		},
	})
}

func AccessTokenCheck(c *gin.Context) {
	tokenStatus := CheckToken(c)
	if tokenStatus == false {
		return
	}

	c.JSON(http.StatusOK, structs.Response{
		Status:  true,
		Message: "Access token valid",
	})
}
