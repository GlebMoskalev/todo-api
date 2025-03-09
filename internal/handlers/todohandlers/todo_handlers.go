package todohandlers

import (
	"encoding/json"
	"github.com/GlebMoskalev/todo-api/internal/models/pagination"
	"github.com/GlebMoskalev/todo-api/internal/models/priority"
	"github.com/GlebMoskalev/todo-api/internal/models/status"
	"github.com/GlebMoskalev/todo-api/internal/models/todo"
	"github.com/GlebMoskalev/todo-api/internal/repository"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func CreateTodo(repo *repository.TodoPostgresRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newTodo todo.Todo
		err := json.NewDecoder(r.Body).Decode(&newTodo)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		id, err := repo.Create(&newTodo)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		jsonId, err := json.Marshal(map[string]int{"id": id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(jsonId)
	}
}

func DeleteTodos(repo *repository.TodoPostgresRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type deleteRequest struct {
			TodoIds []int `json:"ids"`
		}

		var todoIds deleteRequest
		err := json.NewDecoder(r.Body).Decode(&todoIds)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = repo.Delete(todoIds.TodoIds)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte("ok"))
	}
}

func GetByIdTodo(repo *repository.TodoPostgresRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		todoId := chi.URLParam(r, "id")
		id, err := strconv.Atoi(todoId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		todoResponse, err := repo.GetById(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		jsonTodo, err := json.Marshal(todoResponse)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(jsonTodo)
	}
}

func UpdateTodo(repo *repository.TodoPostgresRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var todoForUpdate *todo.Todo
		err := json.NewDecoder(r.Body).Decode(&todoForUpdate)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = repo.Update(todoForUpdate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte("ok"))
	}
}

func GetAllTodos(repo *repository.TodoPostgresRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tags []string
		var statusFilter status.Status
		var priorityFilter priority.Priority
		var overdue *bool
		var dueDate todo.NullTime
		var paginationParams pagination.Pagination

		query := r.URL.Query()

		if rawTags, ok := query["tags"]; ok && len(rawTags) > 0 {
			for _, tag := range rawTags {
				splitTags := strings.Split(tag, ",")
				for _, t := range splitTags {
					trimmed := strings.TrimSpace(t)
					if trimmed != "" {
						tags = append(tags, trimmed)
					}
				}
			}
		}

		overdueStr := query.Get("overdue")
		if overdueStr != "" {
			overdueBool, err := strconv.ParseBool(overdueStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			} else {
				overdue = todo.BoolPtr(overdueBool)
			}
		}

		priorityString := query.Get("priority")
		if priorityString != "" && !priority.IsValidPriority(priority.Priority(priorityString)) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid priority"))
			return
		} else {
			priorityFilter = priority.Priority(priorityString)
		}

		statusString := query.Get("status")
		if statusString != "" && !status.IsValidStatus(status.Status(statusString)) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid status"))
			return
		} else {
			statusFilter = status.Status(statusString)
		}

		dueDateString := query.Get("dueDate")
		if dueDateString != "" {
			date, err := time.Parse(time.DateOnly, dueDateString)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("invalid time"))
				return
			} else {
				dueDate = todo.NullTime{Time: date, Valid: true}
			}
		} else {
			dueDate = todo.NullTime{Valid: false}
		}

		if rawLimit := query.Get("limit"); rawLimit != "" {
			limitInt, err := strconv.Atoi(rawLimit)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			paginationParams.Limit = limitInt
		} else {
			paginationParams.Limit = pagination.DefaultLimit
		}

		if rawOffset := query.Get("offset"); rawOffset != "" {
			offsetInt, err := strconv.Atoi(rawOffset)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			paginationParams.Offset = offsetInt
		} else {
			paginationParams.Offset = pagination.DefaultOffset
		}

		todos, err := repo.GetAll(tags, statusFilter, priorityFilter, overdue, dueDate, paginationParams)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		jsonTodos, err := json.Marshal(todos)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(jsonTodos)
	}
}
