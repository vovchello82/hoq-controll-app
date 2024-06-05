package task

import "time"

type Task struct {
	Name          string     `json:"name"`
	Status        TaskStatus `json:"status"`
	Labels        map[string]string
	TimeLastCheck time.Time `json:"timeLastCheck"`
	References    string    `json:"references"`
}

type TaskStatus int64

const (
	UNDEFINED TaskStatus = iota
	OPEN
	DONE
)

func (ts TaskStatus) String() string {
	switch ts {
	case OPEN:
		return "OPEN"
	case DONE:
		return "DONE"
	}

	return "UNDEFINED"
}

type TaskImpl struct {
	Name    string `json:"name"`
	Labels  map[string]string
	ImageID string `json:"imageID"`
}
