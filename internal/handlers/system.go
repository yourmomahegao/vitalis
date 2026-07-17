package handlers

import (
	"log"
	"net/http"

	"vitalis/internal/handlers/structs"
	"vitalis/internal/services"

	"github.com/gin-gonic/gin"
)

func CpuInformation(c *gin.Context) {
	tokenStatus := CheckToken(c)
	if tokenStatus == false {
		return
	}

	infoFilters := []services.InfoFilter{}
	filterValues := []any{}

	addFilter := func(filter services.InfoFilter) {
		if value := c.PostForm(filter.Name); value != "" {
			infoFilters = append(infoFilters, filter)
			filterValues = append(filterValues, value)
		}
	}

	addFilter(services.InfoFilters.StartDatetime)
	addFilter(services.InfoFilters.EndDatetime)
	addFilter(services.InfoFilters.Limit)
	addFilter(services.InfoFilters.Offset)
	addFilter(services.InfoFilters.CPUName)
	addFilter(services.InfoFilters.CPUPhysicalCoresMin)
	addFilter(services.InfoFilters.CPUPhysicalCoresMax)
	addFilter(services.InfoFilters.CPULogicalCoresMin)
	addFilter(services.InfoFilters.CPULogicalCoresMax)
	addFilter(services.InfoFilters.CPUUtilizationMin)
	addFilter(services.InfoFilters.CPUUtilizationMax)
	addFilter(services.InfoFilters.CPUCurrentSpeedMHzMin)
	addFilter(services.InfoFilters.CPUCurrentSpeedMHzMax)
	addFilter(services.InfoFilters.CPUBaseSpeedMHzMin)
	addFilter(services.InfoFilters.CPUBaseSpeedMHzMax)
	addFilter(services.InfoFilters.CPUProcessesAmountMin)
	addFilter(services.InfoFilters.CPUProcessesAmountMax)
	addFilter(services.InfoFilters.CPUThreadsAmountMin)
	addFilter(services.InfoFilters.CPUThreadsAmountMax)
	addFilter(services.InfoFilters.CPUHandlesAmountMin)
	addFilter(services.InfoFilters.CPUHandlesAmountMax)
	addFilter(services.InfoFilters.CPUUptimeMin)
	addFilter(services.InfoFilters.CPUUptimeMax)

	data, err := services.GetInfoData(services.InfoTypes.CPU, infoFilters, filterValues)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.Response{
			Status:  false,
			Message: "Error while getting CPU information",
		})

		log.Printf("Error occured in CpuInformation(): %v", err)
		return
	}

	c.JSON(http.StatusOK, structs.Response{
		Status: true,
		Data:   data,
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
