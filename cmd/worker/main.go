package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/mike-dunton/chronicler/internal/config"
	"github.com/mike-dunton/chronicler/pkg/listing"
	"github.com/mike-dunton/chronicler/pkg/queue/workqueue"
	"github.com/mike-dunton/chronicler/pkg/storage/sqlite"
	"github.com/mike-dunton/chronicler/pkg/updating"
)

const applicationConfigFile string = "/opt/chronicler/config.json"

// Context is worker context
type Context struct{}

func main() {
	configService := config.NewService(applicationConfigFile)
	appConfig, err := configService.LoadConfig()
	if err != nil {
		fmt.Print(err)
		panic("Failed To Load Application Config")
	}

	storage, _ := sqlite.NewStorage(appConfig.Database.File)
	updater := updating.NewService(storage)
	lister := listing.NewService(storage)
	queue, _ := workqueue.NewQueue(appConfig.Redis.Host, appConfig.Redis.Port, appConfig.Redis.Namespace, appConfig.Redis.DebugPort, lister, updater)
	// validator, _ := validating.NewValidator(appConfig.Subfolders)
	// adder := adding.NewService(storage, queue, validator)

	// Start processing jobs
	queue.StartWorkerPool()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	queue.StopWorkerPool()
}
