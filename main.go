package main

import (
	"bytes"
	"context"
	"fmt"
	"hoa-control-app/cmd/store"
	"hoa-control-app/cmd/watcher"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {

	log.Println("starting the control app for house of apps")
	var store = store.NewInMemStorage()
	watcher := watcher.NewTaskPopulator(store, &watcher.TaskWatcherService{})
	go watcher.StartWatchingChecks()
	go watcher.StartWatchingImpls()

	e := echo.New()

	e.GET("/tasks", func(c echo.Context) error {
		tasks, err := store.GetAllTasks()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, tasks)
	})

	if solved_task_gauge, err := registerGauge("hoa_solved_tasks"); err == nil {
		e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
			AfterNext: func(c echo.Context, err error) {
				num, _ := store.GetNumOfSolvedTasks()
				solved_task_gauge.Set(float64(num))
			},
		}))
	}
	/*
		if tasks_total_gauge, err := registerGauge("hoa_tasks_total"); err == nil {
			e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
				AfterNext: func(c echo.Context, err error) {
					tasks, _ := store.GetAllTasks()
					tasks_total_gauge.Set(float64(len(tasks)))
				},
			}))
		}
	*/
	e.GET("/metrics", echoprometheus.NewHandler())

	e.GET("/tasks/:name",
		func(c echo.Context) error {
			name := c.Param("name")
			task, err := store.GetTaskByName(name)
			if err != nil {
				return c.String(http.StatusNotFound, err.Error())
			}

			return c.JSON(http.StatusOK, task)
		})

	e.GET("/tasks/:name/logs",
		func(c echo.Context) error {
			name := c.Param("name")

			labels := make(map[string]string)
			labels["type"] = "impl"
			labels["task"] = name
			logs := getPodLogsBylabel(labels)
			return c.String(http.StatusOK, logs)
		})

	e.Logger.Fatal(e.Start(":1323"))
}

func getPodLogsBylabel(labelsMap map[string]string) string {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Sprintf("an error occured: %s", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Sprintf("an error occured: %s", err.Error())
	}

	labelSelector := labels.FormatLabels(labelsMap)

	log.Default().Printf("start listing pods with labels: %s in the namespace %s", labelSelector, "K8S_NAMESPACE")

	pods, err := clientset.CoreV1().Pods(watcher.K8S_NAMESPACE).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})

	if err != nil {
		return fmt.Sprintf("an error occured: %s", err.Error())
	}

	log.Default().Printf("pods found %d", len(pods.Items))
	logString := "log output\n"

	for _, p := range pods.Items {
		logString += fmt.Sprintf("------------ %s ------------- \n", p.Name)
		func() {
			logReq := clientset.CoreV1().Pods(watcher.K8S_NAMESPACE).GetLogs(p.Name, &v1.PodLogOptions{})
			podLogs, err := logReq.Stream(context.TODO())
			if err != nil {
				logString = fmt.Sprintf("an error occured: %s", err.Error())
				return
			}
			defer podLogs.Close()

			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, podLogs)
			if err != nil {
				logString += "error in copy information from podLogs to buf"
			}

			logString += buf.String()
		}()

	}

	return logString
}

func registerGauge(name string) (prometheus.Gauge, error) {
	gauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: name,
		},
	)

	if err := prometheus.Register(gauge); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return gauge, nil
}
