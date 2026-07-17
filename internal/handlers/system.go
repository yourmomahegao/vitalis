package handlers

import (
	"errors"
	"net/http"

	"vitalis/internal/handlers/structs"
	"vitalis/internal/services"

	"github.com/gin-gonic/gin"
)

func getFiltersWithValues(c *gin.Context) ([]services.InfoFilter, []any, error) {
	err := c.Request.ParseMultipartForm(32 << 20)
	if err != nil && !errors.Is(err, http.ErrNotMultipart) && !errors.Is(err, http.ErrMissingBoundary) {
		return nil, nil, err
	}

	filters := []services.InfoFilter{}
	filterValues := []any{}

	for argName, argValue := range c.Request.Form {
		if len(argValue) == 0 {
			continue
		}

		for _, filter := range services.InfoFilters {
			if argName == filter.Name {
				filters = append(filters, filter)
				filterValues = append(filterValues, argValue[0])
			}
		}
	}

	return filters, filterValues, nil
}

func CpuInformation(c *gin.Context) {
	tokenStatus := CheckToken(c)
	if tokenStatus == false {
		return
	}

	filters, filterValues, err := getFiltersWithValues(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.Response{
			Status:  false,
			Message: "Error while parsing form data from request",
		})
		return
	}

	infoData, err := services.GetInfoData(services.InfoTypes.CPU, filters, filterValues)

	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.Response{
			Status:  false,
			Message: "Error while getting CPU information",
		})
		return
	}

	c.JSON(http.StatusOK, structs.Response{
		Status:  true,
		Message: "Access token valid",
		Data:    infoData,
	})
}

func RamInformation(c *gin.Context) {
	tokenStatus := CheckToken(c)
	if tokenStatus == false {
		return
	}

	filters, filterValues, err := getFiltersWithValues(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.Response{
			Status:  false,
			Message: "Error while parsing form data from request",
		})
		return
	}

	infoData, err := services.GetInfoData(services.InfoTypes.RAM, filters, filterValues)

	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.Response{
			Status:  false,
			Message: "Error while getting RAM information",
		})
		return
	}

	c.JSON(http.StatusOK, structs.Response{
		Status:  true,
		Message: "Access token valid",
		Data:    infoData,
	})
}

func NetInformation(c *gin.Context) {
	tokenStatus := CheckToken(c)
	if tokenStatus == false {
		return
	}

	filters, filterValues, err := getFiltersWithValues(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.Response{
			Status:  false,
			Message: "Error while parsing form data from request",
		})
		return
	}

	infoData, err := services.GetInfoData(services.InfoTypes.Net, filters, filterValues)

	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.Response{
			Status:  false,
			Message: "Error while getting network information",
		})
		return
	}

	c.JSON(http.StatusOK, structs.Response{
		Status:  true,
		Message: "Access token valid",
		Data:    infoData,
	})
}

func FileInformation(c *gin.Context) {
	tokenStatus := CheckToken(c)
	if tokenStatus == false {
		return
	}

	filters, filterValues, err := getFiltersWithValues(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.Response{
			Status:  false,
			Message: "Error while parsing form data from request",
		})
		return
	}

	infoData, err := services.GetInfoData(services.InfoTypes.File, filters, filterValues)

	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.Response{
			Status:  false,
			Message: "Error while getting file information",
		})
		return
	}

	c.JSON(http.StatusOK, structs.Response{
		Status:  true,
		Message: "Access token valid",
		Data:    infoData,
	})
}
