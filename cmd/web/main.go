package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/mike-dunton/chronicler/pkg/adding"
	"github.com/mike-dunton/chronicler/pkg/http/rest"
	"github.com/mike-dunton/chronicler/pkg/listing"
	"github.com/mike-dunton/chronicler/pkg/queue/workqueue"
	"github.com/mike-dunton/chronicler/pkg/storage/sqlite"
	"github.com/mike-dunton/chronicler/pkg/validating"
)

const applicationConfigFile string = "/opt/chronicler/config.json"

//AppConfig struct
type AppConfig struct {
	Subfolders []string `json:"subfolders"`
	Redis      struct {
		Host      string `json:"host"`
		Port      string `json:"port"`
		Namespace string `json:"namespace"`
		DebugPort string `json:"httpPort"`
	} `json:"redis"`
	Database struct {
		File string `json:"file"`
	} `json:"database"`
	Web struct {
		Port string `json:"port"`
	} `json:"web"`
}

func main() {
	appConfig := parseOptionFile(applicationConfigFile)

	storage, _ := sqlite.NewStorage(appConfig.Database.File)
	queue, _ := workqueue.NewQueue(appConfig.Redis.Host, appConfig.Redis.Port, appConfig.Redis.Namespace, appConfig.Redis.DebugPort)
	validator, _ := validating.NewValidator(appConfig.Subfolders)
	adder := adding.NewService(storage, queue, validator)
	lister := listing.NewService(storage)

	e := echo.New()
	rest.Handler(adder, lister, e)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.Fatal(e.Start(appConfig.Web.Port))
	queue.StartServer()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	<-c
	queue.StopServer()
}

func parseOptionFile(pathToConfig string) (config AppConfig) {
	// Open our jsonFile
	jsonFile, err := os.Open(pathToConfig)
	// if we os.Open returns an error then handle it
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &config)
	return config
}
