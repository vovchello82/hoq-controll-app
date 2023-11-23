package watcher

import (
	"context"
	"hoa-control-app/cmd/store"
	"hoa-control-app/cmd/task"
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	toolsWatch "k8s.io/client-go/tools/watch"
)

type TaskWatcher interface {
	WatchTasks(labelMap map[string]string, feedchan chan<- task.Task)
	WatchJobStatus()
}
type TaskWatcherService struct {
}

func (t *TaskWatcherService) WatchJobStatus() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {
		jobs, err := clientset.BatchV1().CronJobs("default").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		log.Printf("There are %d jobs in the cluster\n", len(jobs.Items))

		for _, j := range jobs.Items {

			log.Printf("Job %s with LastSuccessfulTime %s time", j.Name, j.Status.LastSuccessfulTime)
			log.Printf("Job %s with LastScheduleTime %s time", j.Name, j.Status.LastScheduleTime)

		}

		time.Sleep(20 * time.Second)
	}
}
func (t *TaskWatcherService) WatchTasks(labelMap map[string]string, feedchan chan<- task.Task) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	labelSelector := labels.FormatLabels(labelMap)

	log.Default().Printf("start watching pods with labels: %s", labelSelector)
	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		timeOut := int64(60)

		return clientset.CoreV1().Pods("default").Watch(context.Background(), metav1.ListOptions{
			TimeoutSeconds: &timeOut,
			LabelSelector:  labelSelector,
		})
	}

	watcher, _ := toolsWatch.NewRetryWatcher("1", &cache.ListWatch{WatchFunc: watchFunc})

	for event := range watcher.ResultChan() {
		item := event.Object.(*corev1.Pod)

		switch event.Type {
		case watch.Modified:
			log.Default().Printf("new Modified event pod %s", item.Name)

			feedchan <- task.Task{
				Name:   item.Name,
				Labels: item.Labels,
				Status: "OPEN",
			}
		case watch.Bookmark:
		case watch.Error:
			log.Default().Printf("new Error event pod %s", item.Name)
		case watch.Deleted:
			log.Default().Printf("new Deleted event pod %s", item.Name)
		case watch.Added:
			log.Default().Printf("new Added event pod %s", item.Name)

			feedchan <- task.Task{
				Name:   item.Name,
				Labels: item.Labels,
				Status: "OPEN",
			}
		}
	}

}

type TaskPopulatorService struct {
	Store       store.Store
	TaskWatcher TaskWatcher
}

func NewTaskPopulator(store store.Store, taskWatcher TaskWatcher) *TaskPopulatorService {
	return &TaskPopulatorService{
		Store:       store,
		TaskWatcher: taskWatcher,
	}
}

func (tp *TaskPopulatorService) StartWatching() {
	taskUpdatesChan := make(chan task.Task)
	labels := make(map[string]string)

	labels["app"] = "task1"

	go tp.TaskWatcher.WatchTasks(labels, taskUpdatesChan)
	go tp.TaskWatcher.WatchJobStatus()

	for v := range taskUpdatesChan {
		log.Default().Printf("coming update for %s", v.Name)
		//TODO store.saveOrUpdate
	}
}
