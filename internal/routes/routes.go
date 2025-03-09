package routes

import (
	"github.com/GlebMoskalev/todo-api/internal/repository"
	"github.com/GlebMoskalev/todo-api/internal/routes/todoroutes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRouter(repo *repository.TodoPostgresRepository) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Mount("/todo", todoroutes.Routes(repo))

	return r
}
