package todo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GlebMoskalev/todo-api/internal/models/priority"
	"github.com/GlebMoskalev/todo-api/internal/models/status"
	"time"
)

type NullTime sql.NullTime

func (nt *NullTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nt.Valid = false
		return nil
	}
	errInvalidDueDate := errors.New("invalid due_date format, expected YYYY-MM-DD")
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return errInvalidDueDate
	}
	if str == "" {
		nt.Valid = false
		return nil
	}
	parsedTime, err := time.Parse(time.DateOnly, str)
	if err != nil {
		return errInvalidDueDate
	}
	nt.Valid = true
	nt.Time = parsedTime
	return nil
}
func (nt *NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time.Format(time.DateOnly))
}

type Todo struct {
	ID          int               `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	DueDate     NullTime          `json:"due_date"`
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
