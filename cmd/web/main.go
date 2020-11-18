package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/mike-dunton/chronicler/internal/config"
	internalLog "github.com/mike-dunton/chronicler/internal/logging"
	"github.com/mike-dunton/chronicler/pkg/adding"
	"github.com/mike-dunton/chronicler/pkg/http/rest"
	"github.com/mike-dunton/chronicler/pkg/listing"
	"github.com/mike-dunton/chronicler/pkg/queue/workqueue"
	"github.com/mike-dunton/chronicler/pkg/storage/sqlite"
	"github.com/mike-dunton/chronicler/pkg/updating"
	"github.com/mike-dunton/chronicler/pkg/validating"
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
	lister := listing.NewService(storage)
	updater := updating.NewService(storage)
	queue, _ := workqueue.NewQueue(logger, appConfig.Redis.Host, appConfig.Redis.Port, appConfig.Redis.Namespace, appConfig.Downloads.PathPrefix, lister, updater)
	validator, _ := validating.NewValidator(appConfig.Subfolders)
	adder := adding.NewService(storage, queue, validator)

	e := echo.New()
	rest.Handler(adder, lister, e)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	queue.StartServer()
	e.Logger.Fatal(e.Start(appConfig.Web.Port))
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	<-c
	queue.StopServer()
}
