package store

import (
	"errors"
	"hoa-control-app/cmd/task"
	"log"
)

type Store interface {
	SaveOrUpdateTask(task.Task) error

	GetTaskByName(name string) (task.Task, error)
	GetTaskByLabel(label string, value string) (task.Task, error)
	GetTasksByLabels(labels map[string]string) ([]task.Task, error)
}

type InMemStore struct {
	storage map[string]task.Task
}

func NewInMemStorage() *InMemStore {
	return &InMemStore{
		storage: make(map[string]task.Task),
	}
}

func (s *InMemStore) SaveOrUpdateTask(task task.Task) error {
	log.Default().Printf("save or update task with the name %s", task.Name)
	s.storage[task.Name] = task
	return nil
}

func (s *InMemStore) GetTaskByLabel(label string, value string) (task.Task, error) {
	for _, task := range s.storage {
		if taskLabel, found := task.Labels[label]; found && taskLabel == value {
			return task, nil
		}
	}
	return task.Task{}, errors.New("task not found")
}

func (s *InMemStore) GetTaskByName(name string) (task.Task, error) {
	log.Default().Printf("looking for %s in %s", name, s.storage)
	if task, found := s.storage[name]; found {
		return task, nil
	}
	return task.Task{}, errors.New("task not found")
}

func (s *InMemStore) GetTasksByLabels(labels map[string]string) ([]task.Task, error) {
	return []task.Task{}, nil
}
