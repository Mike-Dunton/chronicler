package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"

	"github.com/gocraft/work"
	"github.com/gocraft/work/webui"
	"github.com/gomodule/redigo/redis"
	guuid "github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const applicationConfig string = "/opt/chronicler/config.json"

// DownloadRequest defines the payload request
type DownloadRequest struct {
	URL       string `json:"url"`
	Subfolder string `json:"subfolder"`
}

//AppOptions struct
type AppOptions struct {
	Subfolders []string `json:"subfolders"`
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

	appConfig := parseOptionFile()
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Route => handler
	e.POST("/", func(c echo.Context) error {
		downloadRequest := new(DownloadRequest)
		if err := c.Bind(downloadRequest); err != nil {
			return err
		}
		if !Contains(appConfig.Subfolders, downloadRequest.Subfolder) {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Please select a valid subfolder: %v", appConfig.Subfolders))
		}
		id := guuid.New()
		outputTemplate := fmt.Sprintf("%v/%%(title)s/%%(title)s-%%(id)s.%%(ext)s", downloadRequest.Subfolder)
		_, err := enqueuer.Enqueue("exec_download", work.Q{"url": downloadRequest.URL, "outputTemplate": outputTemplate, "requestID": id.String()})
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Failed to queue request:  %q", err))
		}

		return c.String(http.StatusOK, "thanks")
	})

	server := webui.NewServer("chronicler", redisPool, ":8181")
	server.Start()
	// Start server
	e.Logger.Fatal(e.Start(":8080"))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	<-c

	server.Stop()
}

func parseOptionFile() (config AppOptions) {
	// Open our jsonFile
	jsonFile, err := os.Open(applicationConfig)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened users.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &config)
	return config
}

// Contains takes a slice and looks for an element in it.
func Contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
