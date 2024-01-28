package task

type Task struct {
	Name   string            `json:"name"`
	Status TaskStatus        `json:"status"`
	Labels map[string]string `json:"labels"`
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
