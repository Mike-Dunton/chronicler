package workqueue

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis/v7"
	"github.com/mike-dunton/chronicler/pkg/adding"
	"github.com/mike-dunton/chronicler/pkg/listing"
	"github.com/mike-dunton/chronicler/pkg/updating"
	"github.com/taylorchu/work"
	"github.com/taylorchu/work/middleware/discard"
	"github.com/taylorchu/work/middleware/logrus"
)

// Queue is the interface that defines interacting with Download Records
type Queue struct {
	queue        work.Queue
	workerPool   *work.Worker
	namespace    string
	logger       log.Logger
	downloadPath string
}

type downloadRequest struct {
	URL            string `json:"url"`
	OutputTemplate string `json:"outputTemplate"`
	RequestID      int64  `json:"requestID"`
}

// NewQueue returns a new Sql DB  storage
func NewQueue(logger *log.Logger, redisHost string, redisPort string, redisNamespace string, pathPrefix string, l listing.Service, u updating.Service) (*Queue, error) {
	q := new(Queue)
	q.namespace = redisNamespace
	q.logger = log.With(*logger, "pkg", "workqueue")
	q.downloadPath = pathPrefix
	opt, err := redis.ParseURL(fmt.Sprintf("redis://%s:%s", redisHost, redisPort))
	if err != nil {
		return nil, err
	}
	redisClient := redis.NewClient(opt)
	q.queue = work.NewRedisQueue(redisClient)
	q.workerPool = work.NewWorker(&work.WorkerOptions{
		Namespace: redisNamespace,
		Queue:     q.queue,
		ErrorFunc: func(err error) {
			level.Error(q.logger).Log("err", err)
		},
	})
	jobOpts := &work.JobOptions{
		MaxExecutionTime: time.Minute,
		IdleWait:         4 * time.Second,
		NumGoroutines:    4,
		HandleMiddleware: []work.HandleMiddleware{
			logrus.HandleFuncLogger,
			discard.After(time.Hour),
		},
	}
	q.workerPool.Register("exec_download", func(job *work.Job, opts *work.DequeueOptions) error {
		return doDownload(q, job, l, u)
	}, jobOpts)

	return q, nil
}

// EnqueueDownload enqueues the download
func (q *Queue) EnqueueDownload(dr *adding.DownloadRecord, recordID int64) error {
	outputTemplate := fmt.Sprintf("/%s/%s/%%(title)s/%%(title)s-%%(id)s.%%(ext)s", q.downloadPath, dr.Subfolder)
	level.Debug(q.logger).Log("outputTemplate", outputTemplate)
	job := work.NewJob()
	err := job.MarshalJSONPayload(downloadRequest{dr.URL, outputTemplate, recordID})
	if err != nil {
		level.Error(q.logger).Log("err", fmt.Sprintf("MarshalJSONPayload Failed %v", err))
		return err
	}
	err = q.queue.Enqueue(job, &work.EnqueueOptions{
		Namespace: q.namespace,
		QueueID:   "exec_download",
	})
	if err != nil {
		level.Error(q.logger).Log("err", fmt.Sprintf("Enqueue Failed %v", err))
		return err
	}
	return nil
}

func doDownload(q *Queue, job *work.Job, l listing.Service, u updating.Service) error {
	// Extract arguments:
	var drArgs downloadRequest
	err := job.UnmarshalJSONPayload(&drArgs)
	if err != nil {
		return err
	}
	dr, err := l.GetDownloadRecord(drArgs.RequestID)
	if err != nil {
		level.Error(q.logger).Log("err", fmt.Sprintf("GetDownloadRecord Failed %v", err))
		return err
	}
	level.Debug(q.logger).Log("RequestID", dr.ID, "URL", dr.URL)
	fileNameCmd := exec.Command("youtube-dl", "--get-title", "--get-filename", "-o", drArgs.OutputTemplate, drArgs.URL)
	var out bytes.Buffer
	var errOut bytes.Buffer
	fileNameCmd.Stdout = &out
	fileNameCmd.Stderr = &errOut
	err = fileNameCmd.Run()
	level.Debug(q.logger).Log("out", out.String())
	level.Debug(q.logger).Log("errOut", errOut.String())
	if err != nil {
		level.Error(q.logger).Log("get filename and title err", err)
		var ur updating.DownloadRecord
		ur.ID = dr.ID
		ur.Finished = "false"
		ur.Errors = errOut.String()
		ur.Output = out.String()
		u.UpdateDownloadRecord(ur)
		return err
	}
	titleAndFilename := strings.Split(out.String(), "\n")
	cmd := exec.Command("youtube-dl", "-o", drArgs.OutputTemplate, drArgs.URL)
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err = cmd.Run()
	if err != nil {
		level.Error(q.logger).Log("cmd_error", errOut.String())
		var ur updating.DownloadRecord
		ur.ID = dr.ID
		ur.Finished = "false"
		ur.Errors = errOut.String()
		ur.Output = out.String()
		ur.Title = titleAndFilename[0]
		ur.Filename = titleAndFilename[1]
		u.UpdateDownloadRecord(ur)
		return err
	}
	level.Debug(q.logger).Log("titleAndFilename[0]", titleAndFilename[0], "titleAndFilename[1]", titleAndFilename[1], "len", len(titleAndFilename))
	level.Debug(q.logger).Log("msg", "Download Successful", "recordID", dr.ID)
	var ur updating.DownloadRecord
	ur.ID = dr.ID
	ur.Finished = "true"
	ur.Errors = errOut.String()
	ur.Output = out.String()
	ur.Title = titleAndFilename[0]
	ur.Filename = titleAndFilename[1]
	u.UpdateDownloadRecord(ur)
	return nil
}

//StartServer starts listening on configured port
func (q *Queue) StartServer() {
	level.Debug(q.logger).Log("msg", "Starting HTTP Work Pool")
	//q.http.Start()
}

//StopServer stops the httpServer
func (q *Queue) StopServer() {
	level.Debug(q.logger).Log("msg", "Stopping HTTP Work Pool")
	//q.http.Stop()
}

//StartWorkerPool does what it says
func (q *Queue) StartWorkerPool() {
	level.Debug(q.logger).Log("msg", "Starting Worker Pool")
	q.workerPool.Start()
}

//StopWorkerPool does what it says
func (q *Queue) StopWorkerPool() {
	level.Debug(q.logger).Log("msg", "Starting Worker Pool")
	q.workerPool.Stop()
}
