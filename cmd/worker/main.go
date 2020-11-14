package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/mike-dunton/chronicler/internal/config"
	internalLog "github.com/mike-dunton/chronicler/internal/logging"
	"github.com/mike-dunton/chronicler/pkg/listing"
	"github.com/mike-dunton/chronicler/pkg/queue/workqueue"
	"github.com/mike-dunton/chronicler/pkg/storage/sqlite"
	"github.com/mike-dunton/chronicler/pkg/updating"
)

const applicationConfigFile string = "/opt/chronicler/config.json"

func main() {
	configService := config.NewService(applicationConfigFile)
	appConfig, err := configService.LoadConfig()
	if err != nil {
		fmt.Print(err)
		panic("Failed To Load Application Config")
	}


	logger := internalLog.NewLogger("debug", "web")

	storage, _ := sqlite.NewStorage(logger, appConfig.Database.File)
	updater := updating.NewService(storage)
	lister := listing.NewService(storage)
	queue, _ := workqueue.NewQueue(logger, appConfig.Redis.Host, appConfig.Redis.Port, appConfig.Redis.Namespace, lister, updater)

	// Start processing jobs
	queue.StartWorkerPool()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	<-signalChan

	queue.StopWorkerPool()
}
