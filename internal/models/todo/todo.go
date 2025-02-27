package todo

import (
	"fmt"
	"github.com/GlebMoskalev/todo-api/internal/models/priority"
	"github.com/GlebMoskalev/todo-api/internal/models/status"
	"github.com/google/uuid"
	"time"
)

type Todo struct {
	ID          uuid.UUID         `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	DueDate     time.Time         `json:"due_date"`
	Tags        []string          `json:"tags"`
	Priority    priority.Priority `json:"priority"`
	Status      status.Status     `json:"status"`
	Overdue     bool              `json:"overdue"`
}

type Todos []Todo

func (t *Todo) Validate() error {
	if !IsValidStatus(t.Status) {
		return fmt.Errorf("invalid value field \"Status\": %s", t.Status)
	}
	if !IsValidPriority(t.Priority) {
		return fmt.Errorf("invalid value field \"Priority\": %s", t.Priority)
	}
	return nil
}

func IsValidStatus(s status.Status) bool {
	switch s {
	case status.Planned, status.InProgress, status.Completed, status.Canceled:
		return true
	default:
		return false
	}
}

func IsValidPriority(p priority.Priority) bool {
	switch p {
	case priority.Low, priority.Medium, priority.High, priority.Urgent:
		return true
	default:
		return false
	}
}
