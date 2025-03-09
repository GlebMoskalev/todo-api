package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/GlebMoskalev/todo-api/internal/models/priority"
	"github.com/GlebMoskalev/todo-api/internal/models/status"
	"github.com/GlebMoskalev/todo-api/internal/models/todo"
	"github.com/lib/pq"
	"log/slog"
	"strings"
	"time"
)

type TodoPostgresRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewTodoPostgresRepository(db *sql.DB, logger *slog.Logger) *TodoPostgresRepository {
	return &TodoPostgresRepository{
		db:     db,
		logger: logger,
	}
}

func (r *TodoPostgresRepository) Create(todo *todo.Todo) (int, error) {
	r.logger.Debug("Attempting to create todo", slog.String("Title", todo.Title))
	if err := todo.Validate(); err != nil {
		r.logger.Warn("Validation failed", slog.String("error", err.Error()))
		return 0, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Error("Failed to begin transaction", slog.String("error", err.Error()))
		return 0, err
	}
	defer func() {
		if err != nil {
			r.logger.Warn("Rolling back transaction", slog.String("error", err.Error()))
			if err = tx.Rollback(); err != nil {
				r.logger.Error("Failed to rollback transaction", slog.String("error", err.Error()))
			}
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
		r.logger.Error("Failed to scan id", slog.String("error", err.Error()))
		return 0, fmt.Errorf("error scanning last insert id: %w", err)
	}
	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", slog.String("error", err.Error()))
		return 0, err
	}

	todo.ID = id
	r.logger.Debug("Todo created successfully", slog.String("Title", todo.Title), slog.Int("ID", id))
	return todo.ID, nil
}

func (r *TodoPostgresRepository) GetById(id int) (*todo.Todo, error) {
	r.logger.Debug("Fetching todo by id", slog.Int("ID", id))
	t := &todo.Todo{}

	var dueDate sql.NullTime
	err := r.db.QueryRow(
		"SELECT id, title, description, due_date, tags, priority, status, overdue FROM todos WHERE id = $1",
		id,
	).Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&dueDate,
		pq.Array(&t.Tags),
		&t.Priority,
		&t.Status,
		&t.Overdue,
	)
	if err != nil {
		r.logger.Warn("Record not found", slog.Int("id", id), slog.String("error", err.Error()))
		return nil, errors.New("record not found")
	}

	if dueDate.Valid {
		t.DueDate = todo.NullTime{
			Time:  dueDate.Time.UTC(),
			Valid: true,
		}
	} else {
		t.DueDate = todo.NullTime{
			Valid: false,
		}
	}

	r.logger.Debug("Todo fetched", slog.Int("id", t.ID))
	return t, nil
}

func (r *TodoPostgresRepository) GetAll(
	tags []string,
	statusFilter status.Status,
	priorityFilter priority.Priority,
	overdue *bool,
	dueDate todo.NullTime) (todo.Todos, error) {
	r.logger.Debug("Fetching all todos", slog.Any("tags", tags),
		slog.String("status", string(statusFilter)), slog.String("priority", string(priorityFilter)),
		slog.Any("overdue", overdue))
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
	if err != nil {
		r.logger.Error("Query failed", slog.String("query", query), slog.String("error", err.Error()))
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Warn("Failed to close rows", slog.String("error", err.Error()))
		}
	}()

	var todos todo.Todos
	for rows.Next() {
		t := todo.Todo{}
		var dueDate sql.NullTime
		err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Description,
			&dueDate,
			pq.Array(&t.Tags),
			&t.Priority,
			&t.Status,
			&t.Overdue,
		)
		if err != nil {
			r.logger.Error("Failed to scan row", slog.String("error", err.Error()))
			return nil, err
		}
		if dueDate.Valid {
			t.DueDate = todo.NullTime{
				Time:  dueDate.Time.UTC(),
				Valid: true,
			}
		} else {
			t.DueDate = todo.NullTime{
				Valid: false,
			}
		}
		todos = append(todos, &t)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Rows processing error", slog.String("error", err.Error()))
		return nil, err
	}

	r.logger.Debug("Todos fetched", slog.Int("count", len(todos)))
	return todos, err
}

func (r *TodoPostgresRepository) Update(todo *todo.Todo) error {
	r.logger.Debug("Updating todo", slog.Int("ID", todo.ID))
	if todo.ID == 0 {
		r.logger.Warn("Missing ID for update")
		return errors.New("absent id")
	}
	if !priority.IsValidPriority(todo.Priority) {
		r.logger.Warn("Invalid priority", slog.String("priority", string(todo.Priority)))
		return errors.New("invalid priority")
	}
	if !status.IsValidStatus(todo.Status) {
		r.logger.Warn("Invalid status", slog.String("status", string(todo.Status)))
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
		r.logger.Error("Failed to begin transaction", slog.String("error", err.Error()))
		return err
	}
	defer func() {
		if err != nil {
			r.logger.Warn("Rolling back transaction", slog.String("error", err.Error()))
			if err = tx.Rollback(); err != nil {
				r.logger.Error("Failed to rollback transaction", slog.String("error", err.Error()))
			}
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
		r.logger.Error("Failed to execute update", slog.String("error", err.Error()))
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", slog.String("error", err.Error()))
		return err
	}

	if rowsAffected == 0 {
		r.logger.Warn("Update failed: no rows affected", slog.Int("id", todo.ID))
		return errors.New("updated failed")
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", slog.String("error", err.Error()))
		return err
	}
	r.logger.Debug("Todo updated", slog.Int("ID", todo.ID))
	return nil
}

func (r *TodoPostgresRepository) Delete(ids []int) error {
	r.logger.Debug("Attempting to delete todo", slog.Any("ids", ids))
	if len(ids) == 0 {
		r.logger.Warn("No IDs provided for deletion")
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
		r.logger.Error("Failed to begin transaction", slog.String("error", err.Error()))
		return err
	}
	defer func() {
		if err != nil {
			r.logger.Warn("Rolling back transaction", slog.String("error", err.Error()))
			if err = tx.Rollback(); err != nil {
				r.logger.Error("Failed to rollback transaction", slog.String("error", err.Error()))
			}
		}
	}()
	res, err := tx.Exec(query, params...)
	if err != nil {
		r.logger.Error("Failed to execute delete", slog.String("error", err.Error()))
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", slog.String("error", err.Error()))
		return err
	}

	if rowsAffected == 0 {
		r.logger.Warn("Delete failed: no rows affected", slog.Any("ids", ids))
		return errors.New("deleted failed")
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", slog.String("error", err.Error()))
		return err
	}
	r.logger.Debug("Todos deleted successfully", slog.Any("ids", ids))
	return nil
}
