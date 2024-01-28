package main

import (
	"hoa-control-app/cmd/store"
	"hoa-control-app/cmd/watcher"
	"log"
	"net/http"

	"github.com/labstack/echo"
)

func main() {

	log.Println("starting the control app for house of apps")
	var store = store.NewInMemStorage()
	watcher := watcher.NewTaskPopulator(store, &watcher.TaskWatcherService{})
	go watcher.StartWatching()

	e := echo.New()
	e.GET("/tasks", func(c echo.Context) error {
		tasks, err := store.GetAllTasks()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, tasks)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
