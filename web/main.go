package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"

	"github.com/gocraft/work"
	"github.com/gocraft/work/webui"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/mattn/go-sqlite3"
)

const applicationConfig string = "/opt/chronicler/config.json"

// DownloadRequest defines the payload request
type DownloadRequest struct {
	URL       string `json:"url"`
	Subfolder string `json:"subfolder"`
}

// DownloadRecord defines the stored download metadata
type DownloadRecord struct {
	ID        int            `json:"id"`
	URL       string         `json:"url"`
	Subfolder string         `json:"subfolder"`
	Output    sql.NullString `json:"output"`
	Errors    sql.NullString `json:"errors"`
	Finished  string         `json:"finished"`
}

// DownloadCollection collection of download records
type DownloadCollection struct {
	DownloadRecords []DownloadRecord `json:"downloads"`
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

	db, err := sql.Open("sqlite3", "/data/sql.db")
	if err != nil {
		panic(err)
	}
	prepDatabase(db)
	appConfig := parseOptionFile()
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "/usr/share/html")

	apiRoutes := e.Group("/api")
	apiRoutes.GET("/", func(c echo.Context) error {
		downloads := allDownloadRecords(db)
		return c.JSON(http.StatusOK, downloads)
	})
	// Route => handler
	apiRoutes.POST("/", func(c echo.Context) error {
		downloadRequest := new(DownloadRequest)
		if err := c.Bind(downloadRequest); err != nil {
			return err
		}
		if !Contains(appConfig.Subfolders, downloadRequest.Subfolder) {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Please select a valid subfolder: %v", appConfig.Subfolders))
		}
		id, err := newDownloadRecord(db, downloadRequest)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Error creating record: %v", err))
		}
		outputTemplate := fmt.Sprintf("%v/%%(title)s/%%(title)s-%%(id)s.%%(ext)s", downloadRequest.Subfolder)
		_, err = enqueuer.Enqueue("exec_download", work.Q{"url": downloadRequest.URL, "outputTemplate": outputTemplate, "requestID": id})
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Failed to queue request:  %q", err))
		}

		return c.String(http.StatusOK, fmt.Sprintf("sdRequestID : %v ", id))
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

func newDownloadRecord(db *sql.DB, request *DownloadRequest) (int64, error) {
	sql := "INSERT INTO downloads(url, subfolder, finished) VALUES(?,?,?)"
	stmt, err := db.Prepare(sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	result, sqlExecError := stmt.Exec(request.URL, request.Subfolder, "false")
	if sqlExecError != nil {
		return 0, sqlExecError
	}
	return result.LastInsertId()
}

func allDownloadRecords(db *sql.DB) DownloadCollection {
	sql := "SELECT id, url, subfolder, output, error, finished FROM downloads"
	rows, err := db.Query(sql)
	// Exit if the SQL doesn't work for some reason
	if err != nil {
		panic(err)
	}
	// make sure to cleanup when the program exits
	defer rows.Close()

	result := DownloadCollection{}
	for rows.Next() {
		download := DownloadRecord{}
		err2 := rows.Scan(&download.ID, &download.URL, &download.Subfolder, &download.Output, &download.Errors, &download.Finished)
		// Exit if we get an error
		if err2 != nil {
			panic(err2)
		}
		result.DownloadRecords = append(result.DownloadRecords, download)
	}
	return result
}

func prepDatabase(db *sql.DB) {
	downloads := `
	CREATE TABLE IF NOT EXISTS downloads(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT, 
		subfolder TEXT, 
		output TEXT, 
		error TEXT, 
		finished TEXT NOT NULL
		CHECK( typeof("finished") = "text" AND
			"finished" IN ("true","false")
		)
	);
	`
	_, err := db.Exec(downloads)
	if err != nil {
		panic(err)
	}
}
