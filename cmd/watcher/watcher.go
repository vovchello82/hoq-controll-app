package watcher

import (
	"context"
	"hoa-control-app/cmd/store"
	"hoa-control-app/cmd/task"
	"log"
	"os"
	"strings"
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

var K8S_NAMESPACE = "default"

type TaskWatcher interface {
	WatchTasks(labelMap map[string]string, feedchan chan<- task.TaskImpl)
	WatchJobStatus(labelsMap map[string]string, feedchan chan<- task.Task)
}
type TaskWatcherService struct {
}

func (t *TaskWatcherService) WatchJobStatus(labelsMap map[string]string, feedchan chan<- task.Task) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	labelSelector := labels.FormatLabels(labelsMap)

	log.Default().Printf("start watching pods with labels: %s", labelSelector)
	timeOut := int64(60)
	for {
		jobs, err := clientset.BatchV1().CronJobs(K8S_NAMESPACE).List(context.TODO(), metav1.ListOptions{
			TimeoutSeconds: &timeOut,
			LabelSelector:  labelSelector,
		})
		if err != nil {
			panic(err.Error())
		}
		if jobs == nil {
			continue
		}

		for _, j := range jobs.Items {
			//skip if job is currently running
			if len(j.Status.Active) > 0 {
				continue
			}

			timeScheduled := j.Status.LastScheduleTime
			timeLastSuccess := j.Status.LastSuccessfulTime
			log.Printf("%s with timeScheduled %s and timeLastSuccess %s", j.Name, timeScheduled, timeLastSuccess)

			if timeScheduled != nil &&
				timeLastSuccess != nil &&
				timeScheduled.Time.Before(timeLastSuccess.Time) {
				log.Printf("%s job is done", j.Name)
				if name, ok := j.Labels["task"]; ok {
					feedchan <- task.Task{
						Name:          name,
						Labels:        j.Labels,
						Status:        task.DONE,
						TimeLastCheck: time.Now(),
					}
				}
			} else {
				log.Printf("%s job stil open", j.Name)
				if name, ok := j.Labels["task"]; ok {
					feedchan <- task.Task{
						Name:          name,
						Labels:        j.Labels,
						Status:        task.OPEN,
						TimeLastCheck: time.Now(),
					}
				}
			}
		}

		// this seems to be sufficient -> delete pod watching and move the persistence logic inside of this method

		time.Sleep(35 * time.Second)
	}
}
func (t *TaskWatcherService) WatchTasks(labelMap map[string]string, feedchan chan<- task.TaskImpl) {
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

	log.Default().Printf("start watching pods in namespace %s with labels: %s", K8S_NAMESPACE, labelSelector)

	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		timeOut := int64(60)

		return clientset.CoreV1().Pods(K8S_NAMESPACE).Watch(context.Background(), metav1.ListOptions{
			TimeoutSeconds: &timeOut,
			LabelSelector:  labelSelector,
		})
	}

	watcher, err := toolsWatch.NewRetryWatcher("1", &cache.ListWatch{WatchFunc: watchFunc})
	if err != nil {
		panic(err.Error())
	}

	for event := range watcher.ResultChan() {
		item := event.Object.(*corev1.Pod)

		switch event.Type {
		case watch.Modified:
			log.Default().Printf("new Modified event pod %s", item.Name)
		case watch.Bookmark:
		case watch.Error:
			log.Default().Printf("new Error event pod %s", item.Name)
		case watch.Deleted:
			log.Default().Printf("new Deleted event pod %s", item.Name)
		case watch.Added:
			log.Default().Printf("new Added event pod %s time %s", item.Name, item.CreationTimestamp)
			var containerImageIds []string
			for _, s := range item.Status.ContainerStatuses {
				log.Default().Printf("image in usage %s", s.ImageID)
				containerImageIds = append(containerImageIds, s.ImageID)
			}
			if name, ok := item.Labels["task"]; ok && len(containerImageIds) > 0 {
				feedchan <- task.TaskImpl{
					Name:    name,
					Labels:  item.Labels,
					ImageID: strings.Join(containerImageIds, ";"),
				}
			}
		}
	}
}

type TaskPopulatorService struct {
	Store       store.Store
	TaskWatcher TaskWatcher
}

func NewTaskPopulator(store store.Store, taskWatcher TaskWatcher) *TaskPopulatorService {
	if namespace, found := os.LookupEnv("NAMESPACE"); found {
		K8S_NAMESPACE = namespace
	}

	return &TaskPopulatorService{
		Store:       store,
		TaskWatcher: taskWatcher,
	}
}

func (tp *TaskPopulatorService) StartWatchingImpls() {
	taskUpdateImageIdChan := make(chan task.TaskImpl)
	labels := make(map[string]string)

	labels["type"] = "impl"

	go tp.TaskWatcher.WatchTasks(labels, taskUpdateImageIdChan)

	for taskImpl := range taskUpdateImageIdChan {
		tp.Store.SaveOrUpdateTask(task.Task{
			Name:       taskImpl.Name,
			Labels:     taskImpl.Labels,
			References: taskImpl.ImageID,
		})
	}
}

func (tp *TaskPopulatorService) StartWatchingChecks() {
	taskUpdatesChan := make(chan task.Task)
	labels := make(map[string]string)

	labels["type"] = "check"

	go tp.TaskWatcher.WatchJobStatus(labels, taskUpdatesChan)

	for v := range taskUpdatesChan {
		tp.Store.SaveOrUpdateTask(task.Task{
			Name:          v.Name,
			Labels:        v.Labels,
			Status:        v.Status,
			TimeLastCheck: v.TimeLastCheck,
		})
	}
}
