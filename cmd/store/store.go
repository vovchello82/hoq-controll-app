package store

import (
	"errors"
	"hoa-control-app/cmd/task"
	"log"
	"sync"
)

type Store interface {
	SaveOrUpdateTask(task.Task) error

	GetAllTasks() ([]task.Task, error)
	GetTaskByName(name string) (task.Task, error)
	GetTaskByLabel(label string, value string) (task.Task, error)
	GetTasksByLabels(labels map[string]string) ([]task.Task, error)
}

type InMemStore struct {
	lock    sync.RWMutex
	storage map[string]task.Task
}

func NewInMemStorage() *InMemStore {
	return &InMemStore{
		lock:    sync.RWMutex{},
		storage: make(map[string]task.Task),
	}
}

func (s *InMemStore) SaveOrUpdateTask(task task.Task) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Default().Printf("save or update task with the name %s", task.Name)
	s.storage[task.Name] = task
	return nil
}

func (s *InMemStore) GetTaskByLabel(label string, value string) (task.Task, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, task := range s.storage {
		if taskLabel, found := task.Labels[label]; found && taskLabel == value {
			return task, nil
		}
	}
	return task.Task{}, errors.New("task not found")
}

func (s *InMemStore) GetTaskByName(name string) (task.Task, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	log.Default().Printf("looking for %s in %s", name, s.storage)
	if task, found := s.storage[name]; found {
		return task, nil
	}
	return task.Task{}, errors.New("task not found")
}

func (s *InMemStore) GetTasksByLabels(labels map[string]string) ([]task.Task, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return []task.Task{}, nil
}

func (s *InMemStore) GetAllTasks() ([]task.Task, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	tasks := []task.Task{}

	for _, t := range s.storage {
		tasks = append(tasks, t)
	}

	return tasks, nil
}
