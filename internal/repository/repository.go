package repository

import (
	"github.com/GlebMoskalev/todo-api/internal/models/priority"
	"github.com/GlebMoskalev/todo-api/internal/models/status"
	"github.com/GlebMoskalev/todo-api/internal/models/todo"
	"time"
)

type TodoRepository interface {
	Create(todo *todo.Todo) (int, error)
	GetById(id int) (*todo.Todo, error)
	GetAll(tags []string, status status.Status, priority priority.Priority, overdue bool, dueDate time.Time) (*todo.Todos, error)
	Update(todo *todo.Todo) error
	Delete(ids []int) error
}
