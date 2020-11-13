package workqueue

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"github.com/gocraft/work"
	"github.com/gocraft/work/webui"
	"github.com/gomodule/redigo/redis"
	"github.com/mike-dunton/chronicler/pkg/adding"
	"github.com/mike-dunton/chronicler/pkg/listing"
	"github.com/mike-dunton/chronicler/pkg/updating"
)

// Queue is the interface that defines interacting with Download Records
type Queue struct {
	queue      *work.Enqueuer
	http       *webui.Server
	workerPool *work.WorkerPool
}

type context struct {
	l listing.Service
	u updating.Service
}

// NewQueue returns a new Sql DB  storage
func NewQueue(redisHost string, redisPort string, redisNamespace string, debugPort string, l listing.Service, u updating.Service) (*Queue, error) {
	q := new(Queue)

	var redisPool = &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
		},
	}
	err := ping(redisPool)
	if err != nil {
		return nil, err
	}

	// Make an enqueuer with a particular namespace
	q.queue = work.NewEnqueuer(redisNamespace, redisPool)
	q.workerPool = work.NewWorkerPool(context{l, u}, 5, redisNamespace, redisPool)
	q.workerPool.Middleware((*context).Log)
	q.workerPool.Job("exec_download", (*context).DoDownload)
	q.http = webui.NewServer(redisNamespace, redisPool, debugPort)
	return q, nil
}

// EnqueueDownload enqueues the download
func (q *Queue) EnqueueDownload(dr *adding.DownloadRecord, recordID int64) error {
	outputTemplate := fmt.Sprintf("%v/%%(title)s/%%(title)s-%%(id)s.%%(ext)s", dr.Subfolder)
	_, err := q.queue.Enqueue("exec_download", work.Q{"url": dr.URL, "outputTemplate": outputTemplate, "requestID": recordID})
	if err != nil {
		return err
	}
	return nil
}

func (c *context) DoDownload(job *work.Job) error {
	// Extract arguments:
	URL := job.ArgString("url")
	outputTemplate := job.ArgString(("outputTemplate"))
	requestID := job.ArgInt64("requestID")
	if err := job.ArgError(); err != nil {
		return err
	}
	fmt.Println("1")
	fmt.Printf("c %v", c)
	dr, err := c.l.GetDownloadRecord(requestID)
	if err != nil {
		fmt.Println("2 ")
		fmt.Printf("err %v", err)
		job.Checkin(fmt.Sprintf("Failed to retreive RequestID %d from storage.", requestID))
		return err
	}
	job.Checkin(fmt.Sprintf("Got Request: %d URL: %s", dr.ID, dr.URL))
	cmd := exec.Command("youtube-dl", fmt.Sprintf("-o %v", outputTemplate), URL)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	if err != nil {
		fmt.Printf("3 err %v", err)
		job.Checkin(fmt.Sprintf("Download Failed: %v URL: %v err: %v", requestID, URL, errOut.String()))
		var ur updating.DownloadRecord
		ur.ID = dr.ID
		ur.Finished = "false"
		ur.Errors = errOut.String()
		ur.Output = out.String()
		c.u.UpdateDownloadRecord(ur)
		return err
	}
	job.Checkin(fmt.Sprintf("Successful Download: %v URL: %v output: %v", requestID, URL, out.String()))
	var ur updating.DownloadRecord
	ur.ID = dr.ID
	ur.Finished = "true"
	ur.Errors = errOut.String()
	ur.Output = out.String()
	c.u.UpdateDownloadRecord(ur)
	return nil
}

func (c *context) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	log.Printf("Starting job: %s %v \n", job.Name, job.Args)
	return next()
}

//StartServer starts listening on configured port
func (q *Queue) StartServer() {
	fmt.Println("Starting HTTP Work Pool")
	q.http.Start()
}

//StopServer stops the httpServer
func (q *Queue) StopServer() {
	fmt.Println("Stopping HTTP Work Pool")
	q.http.Stop()
}

//StartWorkerPool does what it says
func (q *Queue) StartWorkerPool() {
	fmt.Println("Starting Worker Pool")
	q.workerPool.Start()
}

//StopWorkerPool does what it says
func (q *Queue) StopWorkerPool() {
	fmt.Println("Stopping Worker Pool")
	q.workerPool.Stop()
}

func ping(pool *redis.Pool) error {
	conn, err := pool.Dial()
	defer conn.Close()
	if err != nil {
		return fmt.Errorf("ERROR: fail to connect to redis: %s", err.Error())
	}
	_, err = redis.String(conn.Do("PING"))
	if err != nil {
		return fmt.Errorf("ERROR: fail ping redis conn: %s", err.Error())
	}
	return nil
}
