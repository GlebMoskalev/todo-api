package repository

import (
	"github.com/GlebMoskalev/todo-api/internal/models/pagination"
	"github.com/GlebMoskalev/todo-api/internal/models/priority"
	"github.com/GlebMoskalev/todo-api/internal/models/status"
	"github.com/GlebMoskalev/todo-api/internal/models/todo"
)

type TodoRepository interface {
	Create(todo *todo.Todo) (int, error)
	GetById(id int) (*todo.Todo, error)
	GetAll(tags []string, statusFilter status.Status, priorityFilter priority.Priority,
		overdue *bool, dueDate todo.NullTime, pagination pagination.Pagination) (*todo.Todos, error)
	Update(todo *todo.Todo) error
	Delete(ids []int) error
}
