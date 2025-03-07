package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/GlebMoskalev/todo-api/internal/models/priority"
	"github.com/GlebMoskalev/todo-api/internal/models/status"
	"github.com/GlebMoskalev/todo-api/internal/models/todo"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"
)

var masterTestDb *TestDatabase

func TestMain(m *testing.M) {
	masterTestDb = SetupMasterDatabase()
	defer masterTestDb.container.Terminate(context.Background())
	os.Exit(m.Run())
}

func createTestTodo() *todo.Todo {
	return &todo.Todo{
		Title:       "test",
		Description: "for testing",
		DueDate: sql.NullTime{
			Time:  time.Now().UTC().Truncate(24 * time.Hour),
			Valid: true,
		},
		Tags:     []string{"test", "testing"},
		Priority: priority.High,
		Status:   status.InProgress,
		Overdue:  false,
	}
}

func TestCreateTodo(t *testing.T) {
	testCases := []struct {
		name          string
		todo          *todo.Todo
		setup         func(repo *TodoPostgresRepository)
		expectedId    int
		expectedError bool
	}{
		{
			name:       "successfully create",
			todo:       createTestTodo(),
			expectedId: 1,
			setup:      nil,
		},
		{
			name:       "increase id",
			todo:       createTestTodo(),
			expectedId: 2,
			setup: func(repo *TodoPostgresRepository) {
				firstId, err := repo.Create(createTestTodo())
				assert.NoError(t, err)
				assert.Equal(t, 1, firstId)
			},
		},
		{
			name: "invalid priority",
			todo: func() *todo.Todo {
				t := createTestTodo()
				t.Priority = "invalid priority"
				return t
			}(),
			expectedId:    0,
			expectedError: true,
			setup:         nil,
		},
		{
			name: "invalid status",
			todo: func() *todo.Todo {
				t := createTestTodo()
				t.Status = "invalid status"
				return t
			}(),
			expectedId:    0,
			expectedError: true,
			setup:         nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testDbName := fmt.Sprintf("test_db_%d", time.Now().UnixNano())
			testDb, err := SetupTestDatabase(masterTestDb.DbAddress, testDbName)
			assert.NoError(t, err)
			defer testDb.Close()
			defer TearDownTestDatabase(masterTestDb.DbAddress, testDbName)
			repo := TodoPostgresRepository{db: testDb, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
			if tc.setup != nil {
				tc.setup(&repo)
			}
			id, err := repo.Create(tc.todo)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedId, id)
		})
	}
}

func TestGetByIdTodo(t *testing.T) {
	testCases := []struct {
		name          string
		setup         func(repo *TodoPostgresRepository)
		expectedId    int
		expectedError bool
	}{
		{
			name:          "empty database",
			setup:         nil,
			expectedError: true,
		},
		{
			name: "returns task successfully",
			setup: func(repo *TodoPostgresRepository) {
				repo.Create(createTestTodo())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testDbName := fmt.Sprintf("test_db_%d", time.Now().UnixNano())
			testDb, err := SetupTestDatabase(masterTestDb.DbAddress, testDbName)
			assert.NoError(t, err)
			defer testDb.Close()
			defer TearDownTestDatabase(masterTestDb.DbAddress, testDbName)
			repo := TodoPostgresRepository{db: testDb, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
			if tc.setup != nil {
				tc.setup(&repo)
			}
			receivedTodo, err := repo.GetById(1)
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, receivedTodo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, receivedTodo)
			}

		})
	}
}

func TestUpdateTodo(t *testing.T) {
	testCases := []struct {
		name          string
		setup         func(repo *TodoPostgresRepository) *todo.Todo
		expectedError bool
	}{
		{
			name: "successfully update",
			setup: func(repo *TodoPostgresRepository) *todo.Todo {
				id, err := repo.Create(createTestTodo())
				assert.NoError(t, err)

				todoToUpdated := createTestTodo()
				todoToUpdated.ID = id
				todoToUpdated.Title = "updated titile"
				todoToUpdated.Description = "updated description"
				todoToUpdated.Priority = priority.Low
				todoToUpdated.Status = status.Completed
				todoToUpdated.Tags = []string{"updated", "tags"}
				return todoToUpdated
			},
		},
		{
			name: "update non-existing todo",
			setup: func(repo *TodoPostgresRepository) *todo.Todo {
				testTodo := createTestTodo()
				testTodo.ID = 999
				return testTodo
			},
			expectedError: true,
		},
		{
			name: "missing ID",
			setup: func(repo *TodoPostgresRepository) *todo.Todo {
				return createTestTodo()
			},
			expectedError: true,
		},
		{
			name: "invalid priority",
			setup: func(repo *TodoPostgresRepository) *todo.Todo {
				id, err := repo.Create(createTestTodo())
				assert.NoError(t, err)

				todoToUpdated := createTestTodo()
				todoToUpdated.ID = id
				todoToUpdated.Priority = "invalid priority"
				return todoToUpdated
			},
			expectedError: true,
		},
		{
			name: "invalid status",
			setup: func(repo *TodoPostgresRepository) *todo.Todo {
				id, err := repo.Create(createTestTodo())
				assert.NoError(t, err)

				todoToUpdated := createTestTodo()
				todoToUpdated.ID = id
				todoToUpdated.Status = "invalid status"
				return todoToUpdated
			},
			expectedError: true,
		}, {
			name: "update unchanged todo",
			setup: func(repo *TodoPostgresRepository) *todo.Todo {
				testTodo := createTestTodo()
				id, err := repo.Create(testTodo)
				assert.NoError(t, err)
				testTodo.ID = id
				return testTodo
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testDbName := fmt.Sprintf("test_db_%d", time.Now().UnixNano())
			testDb, err := SetupTestDatabase(masterTestDb.DbAddress, testDbName)
			assert.NoError(t, err)
			defer testDb.Close()
			defer TearDownTestDatabase(masterTestDb.DbAddress, testDbName)

			repo := TodoPostgresRepository{db: testDb, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}

			todoToUpdated := tc.setup(&repo)

			err = repo.Update(todoToUpdated)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				updatedTodo, err := repo.GetById(todoToUpdated.ID)
				assert.NoError(t, err)
				assert.Equal(t, todoToUpdated, updatedTodo)
			}
		})
	}
}

func TestDeleteTodo(t *testing.T) {
	testCases := []struct {
		name          string
		setup         func(repo *TodoPostgresRepository) []int
		expectedError bool
	}{
		{
			name: "empty ids",
			setup: func(repo *TodoPostgresRepository) []int {
				return []int{}
			},
			expectedError: true,
		},
		{
			name: "successfully delete one id",
			setup: func(repo *TodoPostgresRepository) []int {
				id, err := repo.Create(createTestTodo())
				assert.NoError(t, err)
				return []int{id}
			},
		},
		{
			name: "successfully delete many id",
			setup: func(repo *TodoPostgresRepository) []int {
				id1, err := repo.Create(createTestTodo())
				assert.NoError(t, err)
				id2, err := repo.Create(createTestTodo())
				assert.NoError(t, err)
				id3, err := repo.Create(createTestTodo())
				assert.NoError(t, err)
				id4, err := repo.Create(createTestTodo())
				assert.NoError(t, err)
				return []int{id1, id2, id3, id4}
			},
		},
		{
			name: "delete non-existing ids",
			setup: func(repo *TodoPostgresRepository) []int {
				return []int{999, 1239}
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testDbName := fmt.Sprintf("test_db_%d", time.Now().UnixNano())
			testDb, err := SetupTestDatabase(masterTestDb.DbAddress, testDbName)
			assert.NoError(t, err)
			defer testDb.Close()
			defer TearDownTestDatabase(masterTestDb.DbAddress, testDbName)

			repo := TodoPostgresRepository{db: testDb, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}

			idsToDelete := tc.setup(&repo)
			err = repo.Delete(idsToDelete)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, id := range idsToDelete {
					_, err = repo.GetById(id)
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestGetAllTodos(t *testing.T) {
	testCases := []struct {
		name          string
		prepareData   func(repo *TodoPostgresRepository) todo.Todos
		getAllTodos   func(repo *TodoPostgresRepository) (todo.Todos, error)
		expectedError bool
	}{
		{
			name: "successfully receiving todo",
			prepareData: func(repo *TodoPostgresRepository) todo.Todos {
				todo1 := createTestTodo()
				id, err := repo.Create(todo1)
				assert.NoError(t, err)
				todo1.ID = id

				todo2 := createTestTodo()
				assert.NoError(t, err)
				id, err = repo.Create(todo2)
				todo2.ID = id
				return todo.Todos{todo1, todo2}
			},
			getAllTodos: func(repo *TodoPostgresRepository) (todo.Todos, error) {
				todos, err := repo.GetAll([]string{}, "", "", nil, sql.NullTime{
					Valid: false,
				})
				return todos, err
			},
		},
		{
			name: "empty database",
			prepareData: func(repo *TodoPostgresRepository) todo.Todos {
				return todo.Todos{}
			},
			getAllTodos: func(repo *TodoPostgresRepository) (todo.Todos, error) {
				todos, err := repo.GetAll([]string{}, "", "", nil, sql.NullTime{
					Valid: false,
				})
				return todos, err
			},
		},
		{
			name: "successfully receiving todo for status",
			prepareData: func(repo *TodoPostgresRepository) todo.Todos {
				todo1 := createTestTodo()
				todo1.Status = status.InProgress
				id, err := repo.Create(todo1)
				assert.NoError(t, err)
				todo1.ID = id

				todo2 := createTestTodo()
				todo2.Status = status.InProgress
				id, err = repo.Create(todo2)
				assert.NoError(t, err)
				todo2.ID = id

				todo3 := createTestTodo()
				todo3.Status = status.Planned
				id, err = repo.Create(todo3)
				assert.NoError(t, err)
				todo3.ID = id
				return todo.Todos{todo1, todo2}
			},
			getAllTodos: func(repo *TodoPostgresRepository) (todo.Todos, error) {
				todos, err := repo.GetAll([]string{}, status.InProgress, "", nil, sql.NullTime{
					Valid: false,
				})
				return todos, err
			},
		},
		{
			name: "successfully receiving todo for priority",
			prepareData: func(repo *TodoPostgresRepository) todo.Todos {
				todo1 := createTestTodo()
				todo1.Priority = priority.High
				id, err := repo.Create(todo1)
				assert.NoError(t, err)
				todo1.ID = id

				todo2 := createTestTodo()
				todo2.Priority = priority.High
				id, err = repo.Create(todo2)
				assert.NoError(t, err)
				todo2.ID = id

				todo3 := createTestTodo()
				todo3.Priority = priority.Low
				id, err = repo.Create(todo3)
				assert.NoError(t, err)
				todo3.ID = id
				return todo.Todos{todo1, todo2}
			},
			getAllTodos: func(repo *TodoPostgresRepository) (todo.Todos, error) {
				todos, err := repo.GetAll([]string{}, "", priority.High, nil, sql.NullTime{
					Valid: false,
				})
				return todos, err
			},
		},
		{
			name: "invalid status",
			prepareData: func(repo *TodoPostgresRepository) todo.Todos {
				return make(todo.Todos, 0)
			},
			getAllTodos: func(repo *TodoPostgresRepository) (todo.Todos, error) {
				todos, err := repo.GetAll([]string{}, "invalid status", "", nil, sql.NullTime{
					Valid: false,
				})
				return todos, err
			},
			expectedError: true,
		},
		{
			name: "invalid priority",
			prepareData: func(repo *TodoPostgresRepository) todo.Todos {
				return make(todo.Todos, 0)
			},
			getAllTodos: func(repo *TodoPostgresRepository) (todo.Todos, error) {
				todos, err := repo.GetAll([]string{}, "", "invalid priority", nil, sql.NullTime{
					Valid: false,
				})
				return todos, err
			},
			expectedError: true,
		},
		{
			name: "successfully receiving todo for tags",
			prepareData: func(repo *TodoPostgresRepository) todo.Todos {
				todo1 := createTestTodo()
				todo1.Tags = []string{"test", "api"}
				id, err := repo.Create(todo1)
				assert.NoError(t, err)
				todo1.ID = id

				todo2 := createTestTodo()
				todo2.Tags = []string{"test", "todo"}
				id, err = repo.Create(todo2)
				assert.NoError(t, err)
				todo2.ID = id

				todo3 := createTestTodo()
				todo3.Tags = []string{"api", "test"}
				id, err = repo.Create(todo3)
				assert.NoError(t, err)
				todo3.ID = id

				todo4 := createTestTodo()
				todo4.Tags = []string{}
				_, err = repo.Create(todo4)
				assert.NoError(t, err)
				return todo.Todos{todo1, todo2, todo3}
			},
			getAllTodos: func(repo *TodoPostgresRepository) (todo.Todos, error) {
				todos, err := repo.GetAll([]string{"test", "api"}, "", priority.High, nil, sql.NullTime{
					Valid: false,
				})
				return todos, err
			},
		},
		{
			name: "successfully receiving todo for overdue",
			prepareData: func(repo *TodoPostgresRepository) todo.Todos {
				todo1 := createTestTodo()
				todo1.Overdue = true
				id, err := repo.Create(todo1)
				assert.NoError(t, err)
				todo1.ID = id

				todo2 := createTestTodo()
				todo2.Overdue = true
				id, err = repo.Create(todo2)
				assert.NoError(t, err)
				todo2.ID = id

				todo3 := createTestTodo()
				id, err = repo.Create(todo3)
				assert.NoError(t, err)
				return todo.Todos{todo1, todo2}
			},
			getAllTodos: func(repo *TodoPostgresRepository) (todo.Todos, error) {
				todos, err := repo.GetAll([]string{}, "", "", todo.BoolPtr(true), sql.NullTime{
					Valid: false,
				})
				return todos, err
			},
		},
		{
			name: "successfully receiving todo for time",
			prepareData: func(repo *TodoPostgresRepository) todo.Todos {
				todo1 := createTestTodo()
				todo1.DueDate = sql.NullTime{
					Time:  time.Date(2030, 12, 30, 0, 0, 0, 0, time.UTC),
					Valid: true,
				}
				id, err := repo.Create(todo1)
				assert.NoError(t, err)
				todo1.ID = id

				todo2 := createTestTodo()
				todo2.DueDate = sql.NullTime{
					Time:  time.Date(2030, 12, 30, 12, 0, 0, 0, time.UTC),
					Valid: true,
				}
				id, err = repo.Create(todo2)
				todo2.DueDate.Time = todo2.DueDate.Time.Truncate(24 * time.Hour)
				assert.NoError(t, err)
				todo2.ID = id

				todo3 := createTestTodo()
				todo3.DueDate = sql.NullTime{
					Time:  time.Date(2030, 12, 30, 14, 30, 300, 0, time.UTC),
					Valid: true,
				}
				id, err = repo.Create(todo3)
				assert.NoError(t, err)
				todo3.DueDate.Time = todo2.DueDate.Time.Truncate(24 * time.Hour)
				todo3.ID = id

				todo4 := createTestTodo()
				todo4.DueDate = sql.NullTime{
					Time:  time.Date(2030, 10, 30, 14, 30, 300, 0, time.UTC),
					Valid: true,
				}
				id, err = repo.Create(todo4)
				assert.NoError(t, err)
				return todo.Todos{todo1, todo2, todo3}
			},
			getAllTodos: func(repo *TodoPostgresRepository) (todo.Todos, error) {
				todos, err := repo.GetAll([]string{}, "", "", nil, sql.NullTime{
					Valid: true,
					Time:  time.Date(2030, 12, 30, 0, 0, 0, 0, time.UTC),
				})
				return todos, err
			},
		},
		{
			name: "successfully receiving todo by multiple parameters",
			prepareData: func(repo *TodoPostgresRepository) todo.Todos {
				todo1 := createTestTodo()
				todo1.Overdue = true
				todo1.Priority = priority.High
				todo1.Status = status.InProgress
				todo1.Tags = []string{"api", "todo1"}
				id, err := repo.Create(todo1)
				assert.NoError(t, err)
				todo1.ID = id

				todo2 := createTestTodo()
				todo2.Overdue = true
				todo2.Priority = priority.High
				todo2.Status = status.InProgress
				todo2.Tags = []string{"api", "todo2"}
				id, err = repo.Create(todo2)
				assert.NoError(t, err)
				todo2.ID = id

				todo3 := createTestTodo()
				id, err = repo.Create(todo3)
				assert.NoError(t, err)
				return todo.Todos{todo1, todo2}
			},
			getAllTodos: func(repo *TodoPostgresRepository) (todo.Todos, error) {
				todos, err := repo.GetAll([]string{"api"}, status.InProgress, priority.High, todo.BoolPtr(true), sql.NullTime{
					Valid: false,
				})
				return todos, err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testDbName := fmt.Sprintf("test_db_%d", time.Now().UnixNano())
			testDb, err := SetupTestDatabase(masterTestDb.DbAddress, testDbName)
			assert.NoError(t, err)
			defer testDb.Close()
			defer TearDownTestDatabase(masterTestDb.DbAddress, testDbName)

			repo := TodoPostgresRepository{db: testDb, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
			expectedTodos := tc.prepareData(&repo)
			fetchedTodos, err := tc.getAllTodos(&repo)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(expectedTodos), len(fetchedTodos))
				assert.ElementsMatch(t, expectedTodos, fetchedTodos)
			}
		})
	}
}
