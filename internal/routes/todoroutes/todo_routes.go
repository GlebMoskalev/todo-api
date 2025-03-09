package todoroutes

import (
	"github.com/GlebMoskalev/todo-api/internal/handlers/todohandlers"
	"github.com/GlebMoskalev/todo-api/internal/repository"
	"github.com/go-chi/chi/v5"
)

func Routes(repo *repository.TodoPostgresRepository) chi.Router {
	r := chi.NewRouter()

	r.Post("/", todohandlers.CreateTodo(repo))
	r.Delete("/", todohandlers.DeleteTodos(repo))
	r.Get("/{id}", todohandlers.GetByIdTodo(repo))
	r.Get("/", todohandlers.GetAllTodos(repo))
	r.Put("/", todohandlers.UpdateTodo(repo))
	return r
}
