package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	guuid "github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// DownloadRequest defines the payload request
type DownloadRequest struct {
	URL string `json:"url"`
}

func main() {
	var redisPool = &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "redis:6379")
		},
	}

	// Make an enqueuer with a particular namespace
	var enqueuer = work.NewEnqueuer("chronicler", redisPool)

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Route => handler
	e.GET("/", func(c echo.Context) error {
		cmd := exec.Command("youtube-dl", "--version")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}

		return c.String(http.StatusOK, fmt.Sprintf("youtube-dl version:  %q", out.String()))
	})

	// Route => handler
	e.POST("/", func(c echo.Context) error {
		downloadRequest := new(DownloadRequest)
		if err := c.Bind(downloadRequest); err != nil {
			return err
		}
		id := guuid.New()
		_, err := enqueuer.Enqueue("exec_download", work.Q{"url": downloadRequest.URL, "requestID": id.String()})
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Failed to queue request:  %q", err))
		}

		return c.String(http.StatusOK, "thanks")
	})

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
