package workqueue

import (
	"fmt"

	"github.com/gocraft/work"
	"github.com/gocraft/work/webui"
	"github.com/gomodule/redigo/redis"
	"github.com/mike-dunton/chronicler/pkg/adding"
)

// Queue is the interface that defines interacting with Download Records
type Queue struct {
	queue *work.Enqueuer
	http  *webui.Server
}

// NewQueue returns a new Sql DB  storage
func NewQueue(redisHost string, redisPort string, redisNamespace string, debugPort string) (*Queue, error) {
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

//StartServer starts listening on configured port
func (q *Queue) StartServer() {
	q.http.Start()
}

//StopServer stops the httpServer
func (q *Queue) StopServer() {
	q.http.Stop()
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
