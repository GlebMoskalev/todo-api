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

	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var utcDueDate any
	if todo.DueDate.Valid {
		utcDueDate = todo.DueDate.Time.UTC()
	} else {
		utcDueDate = nil
	}
	row := tx.QueryRow(
		"INSERT INTO todos (title, description, due_date, tags, priority, status, overdue) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		todo.Title,
		todo.Description,
		utcDueDate,
		pq.Array(todo.Tags),
		todo.Priority,
		todo.Status,
		todo.Overdue,
	)
	var id int
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("error scanning last insert id: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	todo.ID = id
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

	if t.DueDate.Valid {
		t.DueDate.Time = t.DueDate.Time.UTC()
	}

	return t, nil
}

func (r *TodoPostgresRepository) GetAll(
	tags []string,
	statusFilter status.Status,
	priorityFilter priority.Priority,
	overdue *bool,
	dueDate sql.NullTime) (todo.Todos, error) {
	query := "SELECT id, title, description, due_date, tags, priority, status, overdue FROM todos"

	var conditions []string
	var params []interface{}
	paramsCount := 1

	if len(tags) > 0 {
		var tagConditions []string
		for _, tag := range tags {
			if tag != "" {
				tagConditions = append(tagConditions, fmt.Sprintf("$%d = ANY(tags)", paramsCount))
				params = append(params, tag)
				paramsCount++
			}
		}
		if len(tagConditions) > 0 {
			conditions = append(conditions, "("+strings.Join(tagConditions, " OR ")+")")
		}
	}

	if statusFilter != "" {
		if !status.IsValidStatus(statusFilter) {
			return nil, fmt.Errorf("invalid value field \"Status\": %s", statusFilter)
		}
		conditions = append(conditions, fmt.Sprintf("status = $%d", paramsCount))
		params = append(params, string(statusFilter))
		paramsCount++
	}

	if priorityFilter != "" {
		if !priority.IsValidPriority(priorityFilter) {
			return nil, fmt.Errorf("invalid value field \"Priority\": %s", priorityFilter)
		}
		conditions = append(conditions, fmt.Sprintf("priority = $%d", paramsCount))
		params = append(params, string(priorityFilter))
		paramsCount++
	}

	if overdue != nil {
		conditions = append(conditions, fmt.Sprintf("overdue = $%d", paramsCount))
		params = append(params, *overdue)
		paramsCount++
	}

	if dueDate.Valid {
		conditions = append(conditions, fmt.Sprintf("due_date = $%d", paramsCount))
		params = append(params, dueDate.Time.UTC().Format(time.DateOnly))
		paramsCount++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := r.db.Query(query, params...)
	defer rows.Close()

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
		if t.DueDate.Valid {
			t.DueDate.Time = t.DueDate.Time.UTC()
		}
		todos = append(todos, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, err
}

func (r *TodoPostgresRepository) Update(todo *todo.Todo) error {
	if todo.ID == 0 {
		return errors.New("absent id")
	}
	if !priority.IsValidPriority(todo.Priority) {
		return errors.New("invalid priority")
	}
	if !status.IsValidStatus(todo.Status) {
		return errors.New("invalid status")
	}

	var utcDueDate any
	if todo.DueDate.Valid {
		utcDueDate = todo.DueDate.Time.UTC().Format(time.DateOnly)
	} else {
		utcDueDate = nil
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	res, err := tx.Exec(
		"UPDATE todos set title = $1, description = $2, due_date = $3, tags = $4, priority = $5,"+
			" status = $6, overdue = $7 WHERE id = $8",
		todo.Title,
		todo.Description,
		utcDueDate,
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

	if err := tx.Commit(); err != nil {
		return err
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

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	res, err := tx.Exec(query, params...)
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

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
