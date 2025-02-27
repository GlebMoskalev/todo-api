package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/GlebMoskalev/todo-api/internal/models/priority"
	"github.com/GlebMoskalev/todo-api/internal/models/status"
	"github.com/GlebMoskalev/todo-api/internal/models/todo"
	"github.com/lib/pq"
	"strings"
	"time"
)

type TodoPostgresRepository struct {
	db *sql.DB
}

func NewTodoPostgresRepository(db *sql.DB) *TodoPostgresRepository {
	return &TodoPostgresRepository{
		db: db,
	}
}

func (r *TodoPostgresRepository) Create(todo *todo.Todo) (int, error) {
	if err := todo.Validate(); err != nil {
		return 0, err
	}

	_, err := r.db.Exec(
		"INSERT INTO todos (id, title, description, due_date, tags, priority, status, overdue) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		todo.ID,
		todo.Title,
		todo.Description,
		todo.DueDate,
		pq.Array(todo.Tags),
		todo.Priority,
		todo.Status,
		todo.Overdue,
	)
	if err != nil {
		return 0, fmt.Errorf("error inserting todo: %w", err)
	}
	return todo.ID, nil
}

func (r *TodoPostgresRepository) GetById(id int) (*todo.Todo, error) {
	t := &todo.Todo{}
	err := r.db.QueryRow(
		"SELECT id, title, description, due_date, tags, priority, status, overdue FROM todos WHERE id = $1",
		id,
	).Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&t.DueDate,
		pq.Array(&t.Tags),
		&t.Priority,
		&t.Status,
		&t.Overdue,
	)
	if err != nil {
		return nil, errors.New("record not found")
	}
	return t, nil
}

func (r *TodoPostgresRepository) GetAll(
	tags []string,
	statusFilter status.Status,
	priorityFilter priority.Priority,
	overdue bool,
	dueDate time.Time) (*todo.Todos, error) {
	query := "SELECT id, title, description, due_date, tags, priority, status, overdue FROM todos"

	var conditions []string
	var params []interface{}
	paramsCount := 1

	if len(tags) > 0 {
		var tagConditions []string
		for _, tag := range tags {
			tagConditions = append(tagConditions, fmt.Sprintf("$%d = ANY(tags)", paramsCount))
			params = append(params, tag)
			paramsCount++
		}
		if len(tagConditions) > 0 {
			conditions = append(conditions, "("+strings.Join(tagConditions, " OR ")+")")
		}
	}

	if statusFilter != "" {
		if !status.IsValidStatus(statusFilter) {
			return nil, fmt.Errorf("invalid value field \"Status\": %s", statusFilter)
		}
		conditions = append(conditions, fmt.Sprintf("statusFilter = $%d", paramsCount))
		params = append(params, string(statusFilter))
		paramsCount++
	}

	if priorityFilter != "" {
		if !priority.IsValidPriority(priorityFilter) {
			return nil, fmt.Errorf("invalid value field \"Priority\": %s", priorityFilter)
		}
		conditions = append(conditions, fmt.Sprintf("priorityFilter = $%d", paramsCount))
		params = append(params, string(priorityFilter))
		paramsCount++
	}
	if overdue {
		conditions = append(conditions, fmt.Sprintf("overdue = $%d", paramsCount))
		params = append(params, "true")
		paramsCount++
	}

	if !dueDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("due_date = $%d", paramsCount))
		params = append(params, dueDate)
		paramsCount++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, err
	}

	var todos todo.Todos
	for rows.Next() {
		t := todo.Todo{}
		err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Description,
			&t.DueDate,
			pq.Array(&t.Tags),
			&t.Priority,
			&t.Status,
			&t.Overdue,
		)
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &todos, err
}

func (r *TodoPostgresRepository) Update(todo *todo.Todo) error {
	res, err := r.db.Exec(
		"UPDATE todos set title = $1, description = $2, due_date = $3, tags = $4, priority = $5,"+
			" status = $6, overdue = $7 WHERE id = $8",
		todo.Title,
		todo.Description,
		todo.DueDate,
		pq.Array(todo.Tags),
		todo.Priority,
		todo.Status,
		todo.Overdue,
		todo.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("updated failed")
	}

	return nil
}

func (r *TodoPostgresRepository) Delete(ids []int) error {
	if len(ids) == 0 {
		return errors.New("no ids provided for deletion")
	}

	placeholders := make([]string, len(ids))
	params := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		params[i] = id
	}

	query := fmt.Sprintf("DELETE FROM todos WHERE id IN (%s)", strings.Join(placeholders, ", "))

	res, err := r.db.Exec(query, params...)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("deleted failed")
	}

	return nil
}
