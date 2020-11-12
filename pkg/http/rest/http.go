package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/mike-dunton/chronicler/pkg/adding"
	"github.com/mike-dunton/chronicler/pkg/listing"
)

// H return arbitrary JSON in our response
type H map[string]interface{}

//Handler builds the function handlers
func Handler(a adding.Service, l listing.Service, e *echo.Echo) *echo.Echo {
	apiRoutes := e.Group("/api")
	apiRoutes.GET("/", getRecords(l))
	apiRoutes.GET("/:id", getRecord(l))
	apiRoutes.POST("/", putRecord(a))
	return e
}

// GetRecords endpoint
func getRecords(l listing.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		downloadRecords, err := l.GetDownloadRecords()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, downloadRecords)
	}
}

func getRecord(l listing.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		i, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, H{})
		}
		downloadRecord, err := l.GetDownloadRecord(i)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, H{})
		}
		return c.JSON(http.StatusOK, downloadRecord)
	}
}

// PutRequest endpoint
func putRecord(a adding.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		var request adding.DownloadRecord
		// bind user data to our Request struct
		if err := c.Bind(&request); err != nil {
			return err
		}
		if a.ValidateDownloadRecord(request) {
			id, err := a.AddDownloadRecord(request)
			if err != nil {
				return c.String(http.StatusInternalServerError, fmt.Sprintf("Error creating record: %v", err))
			}
			return c.JSON(http.StatusCreated, H{
				"created": id,
			})
		}
		return c.String(http.StatusBadRequest, "Request failed to validate")
	}
}
