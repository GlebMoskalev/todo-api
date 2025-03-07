package todo

import (
	"database/sql"
	"fmt"
	"github.com/GlebMoskalev/todo-api/internal/models/priority"
	"github.com/GlebMoskalev/todo-api/internal/models/status"
)

type Todo struct {
	ID          int               `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	DueDate     sql.NullTime      `json:"due_date"`
	Tags        []string          `json:"tags"`
	Priority    priority.Priority `json:"priority"`
	Status      status.Status     `json:"status"`
	Overdue     bool              `json:"overdue"`
}

type Todos []*Todo

func (t *Todo) Validate() error {
	if !status.IsValidStatus(t.Status) {
		return fmt.Errorf("invalid value field \"Status\": %s", t.Status)
	}
	if !priority.IsValidPriority(t.Priority) {
		return fmt.Errorf("invalid value field \"Priority\": %s", t.Priority)
	}
	return nil
}

func BoolPtr(b bool) *bool {
	return &b
}
