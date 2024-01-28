package task

type Task struct {
	Name   string
	Status TaskStatus
	Labels map[string]string
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
