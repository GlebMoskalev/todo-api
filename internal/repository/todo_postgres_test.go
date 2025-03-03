package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/GlebMoskalev/todo-api/internal/models/priority"
	"github.com/GlebMoskalev/todo-api/internal/models/status"
	"github.com/GlebMoskalev/todo-api/internal/models/todo"
	"github.com/stretchr/testify/assert"
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
		DueDate:     time.Now().Add(24 * time.Hour),
		Tags:        []string{"test", "testing"},
		Priority:    priority.High,
		Status:      status.InProgress,
		Overdue:     false,
	}
}

func TestCreateTodo(t *testing.T) {
	testCases := []struct {
		name          string
		todo          *todo.Todo
		setup         func(db *sql.DB)
		expectedId    int
		expectedError bool
	}{
		{
			name:          "successfully create",
			todo:          createTestTodo(),
			expectedId:    1,
			expectedError: false,
			setup:         nil,
		},
		{
			name:       "increase id",
			todo:       createTestTodo(),
			expectedId: 2,
			setup: func(db *sql.DB) {
				repo := TodoPostgresRepository{db: db}
				firstId, err := repo.Create(createTestTodo())
				assert.NoError(t, err)
				assert.Equal(t, 1, firstId)
			},
			expectedError: false,
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
			repo := TodoPostgresRepository{db: testDb}
			if tc.setup != nil {
				tc.setup(testDb)
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
		setup         func(db *sql.DB)
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
			setup: func(db *sql.DB) {
				repo := TodoPostgresRepository{db: db}
				repo.Create(createTestTodo())
			},
			expectedError: false,
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
			repo := TodoPostgresRepository{db: testDb}
			if tc.setup != nil {
				tc.setup(testDb)
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
		setup         func(db *sql.DB) *todo.Todo
		expectedError bool
	}{
		{
			name: "successfully update",
			setup: func(db *sql.DB) *todo.Todo {
				repo := TodoPostgresRepository{db: db}
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
			expectedError: false,
		},
		{
			name: "update non-existing todo",
			setup: func(db *sql.DB) *todo.Todo {
				testTodo := createTestTodo()
				testTodo.ID = 999
				return testTodo
			},
			expectedError: true,
		},
		{
			name: "missing ID",
			setup: func(db *sql.DB) *todo.Todo {
				return createTestTodo()
			},
			expectedError: true,
		},
		{
			name: "invalid priority",
			setup: func(db *sql.DB) *todo.Todo {
				repo := TodoPostgresRepository{db: db}
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
			setup: func(db *sql.DB) *todo.Todo {
				repo := TodoPostgresRepository{db: db}
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
			setup: func(db *sql.DB) *todo.Todo {
				repo := TodoPostgresRepository{db: db}
				testTodo := createTestTodo()
				id, err := repo.Create(testTodo)
				assert.NoError(t, err)
				testTodo.ID = id
				return testTodo
			},
			expectedError: false,
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

			repo := TodoPostgresRepository{db: testDb}

			var todoToUpdated *todo.Todo
			todoToUpdated = tc.setup(testDb)

			err = repo.Update(todoToUpdated)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				updatedTodo, err := repo.GetById(todoToUpdated.ID)
				assert.NoError(t, err)
				assert.Equal(t, todoToUpdated.Title, updatedTodo.Title)
				assert.Equal(t, todoToUpdated.Description, updatedTodo.Description)
				assert.Equal(t, todoToUpdated.Priority, updatedTodo.Priority)
				assert.Equal(t, todoToUpdated.Overdue, updatedTodo.Overdue)
				assert.Equal(t, todoToUpdated.Status, updatedTodo.Status)
				assert.ElementsMatch(t, todoToUpdated.Tags, updatedTodo.Tags)
			}
		})
	}
}

func TestDeleteTodo(t *testing.T) {
	testCases := []struct {
		name          string
		setup         func(db *sql.DB) []int
		expectedError bool
	}{
		{
			name: "empty ids",
			setup: func(db *sql.DB) []int {
				return []int{}
			},
			expectedError: true,
		},
		{
			name: "successfully delete one id",
			setup: func(db *sql.DB) []int {
				repo := TodoPostgresRepository{db: db}
				id, err := repo.Create(createTestTodo())
				assert.NoError(t, err)
				return []int{id}
			},
			expectedError: false,
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

			repo := TodoPostgresRepository{db: testDb}

			var idsToDelete []int
			idsToDelete = tc.setup(testDb)
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
