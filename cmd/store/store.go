package store

import "hoa-control-app/cmd/task"

type Store interface {
	SaveOrUpdateTask(task.Task) error

	GetTaskByLabel(label string, value string) (task.Task, error)
	GetTasksByLabels(labels map[string]string) ([]task.Task, error)
}

type InMemStore struct {
}

func (s *InMemStore) SaveOrUpdateTask(task.Task) error {
	return nil
}

func (s *InMemStore) GetTaskByLabel(label string, value string) (task.Task, error) {
	return task.Task{}, nil
}
func (s *InMemStore) GetTasksByLabels(labels map[string]string) ([]task.Task, error) {
	return []task.Task{}, nil
}
