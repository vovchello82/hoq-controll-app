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
	WatchJobStatus(feedchan chan<- task.Task)
}
type TaskWatcherService struct {
}

func (t *TaskWatcherService) WatchJobStatus(feedchan chan<- task.Task) {
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
		if jobs == nil {
			continue
		}

		for _, j := range jobs.Items {
			timeScheduled := j.Status.LastScheduleTime
			timeLastSuccess := j.Status.LastSuccessfulTime
			log.Printf("%s with timeScheduled %s and timeLastSuccess %s", j.Name, timeScheduled, timeLastSuccess)
			if timeScheduled != nil &&
				timeLastSuccess != nil &&
				timeScheduled.Time.Before(timeLastSuccess.Time) {
				log.Printf("%s job is done", j.Name)
				feedchan <- task.Task{
					Name:   j.Name,
					Labels: j.Labels,
					Status: task.DONE,
				}
			} else {
				log.Printf("%s job stil open", j.Name)
				feedchan <- task.Task{
					Name:   j.Name,
					Labels: j.Labels,
					Status: task.OPEN,
				}
			}
		}

		// this seems to be sufficient -> delete pod warching and move the persistence logic inside of this method

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
				Status: task.UNDEFINED,
			}
		case watch.Bookmark:
		case watch.Error:
			log.Default().Printf("new Error event pod %s", item.Name)
		case watch.Deleted:
			log.Default().Printf("new Deleted event pod %s", item.Name)
		case watch.Added:
			log.Default().Printf("new Added event pod %s time %s", item.Name, item.CreationTimestamp)
			feedchan <- task.Task{
				Name:   item.Name,
				Labels: item.Labels,
				Status: task.OPEN,
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

	labels["type"] = "task"

	//go tp.TaskWatcher.WatchTasks(labels, taskUpdatesChan)
	go tp.TaskWatcher.WatchJobStatus(taskUpdatesChan)

	for v := range taskUpdatesChan {
		if t, err := tp.Store.GetTaskByName(v.Name); err != nil {
			log.Default().Printf("update old status %s new %s", t.Status, v.Status)
		}
		tp.Store.SaveOrUpdateTask(v)
	}
}
