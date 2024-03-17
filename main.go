package main

import (
	"hoa-control-app/cmd/store"
	"hoa-control-app/cmd/watcher"
	"log"
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
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

	taskGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "solvedTasks",
		},
	)

	if err := prometheus.Register(taskGauge); err != nil {
		log.Fatal(err)
	}

	e.GET("/metrics", echoprometheus.NewHandler())

	e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		AfterNext: func(c echo.Context, err error) {
			num, _ := store.GetNumOfSolvedTasks()
			taskGauge.Set(float64(num))
		},
	}))

	e.GET("/tasks/:name",
		func(c echo.Context) error {
			name := c.Param("name")
			task, err := store.GetTaskByName(name)
			if err != nil {
				return c.String(http.StatusNotFound, err.Error())
			}

			return c.JSON(http.StatusOK, task)
		})

	e.Logger.Fatal(e.Start(":1323"))
}
