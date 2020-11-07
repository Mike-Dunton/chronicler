package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
)

// Context is worker context
type Context struct {
	requestID int64
}

func main() {
	// Make a redis pool
	var redisPool = &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "redis:6379")
		},
	}

	pool := work.NewWorkerPool(Context{}, 5, "chronicler", redisPool)

	pool.Middleware((*Context).Log)

	// Map the name of jobs to handler functions
	pool.Job("exec_download", (*Context).Download)

	// Start processing jobs
	pool.Start()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	pool.Stop()

}

// Log is job middle ware that logs the job name.
func (c *Context) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	log.Println("Starting job: ", job.Name)
	return next()
}

// Download is this workers download function
func (c *Context) Download(job *work.Job) error {
	// Extract arguments:
	URL := job.ArgString("url")
	requestID := job.ArgString("requestID")
	if err := job.ArgError(); err != nil {
		return err
	}
	job.Checkin(fmt.Sprintf("Request: %v URL: %v", requestID, URL))
	cmd := exec.Command("youtube-dl", URL)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	if err != nil {
		return err
		//return c.String(http.StatusBadRequest, fmt.Sprintf("err:  %q", errOut.String()))
	}

	return nil
}
