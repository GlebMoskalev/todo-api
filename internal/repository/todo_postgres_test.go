package repository

import (
	"database/sql"
	"github.com/GlebMoskalev/todo-api/internal/models/priority"
	"github.com/GlebMoskalev/todo-api/internal/models/status"
	"github.com/GlebMoskalev/todo-api/internal/models/todo"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var testDbInstance *sql.DB

func TestMain(m *testing.M) {
	testSetupDb := SetupTestDataBase()
	testDbInstance = testSetupDb.DbInstance
	defer testSetupDb.TearDown()
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

func cleanupDatabase(db *sql.DB) error {
	_, err := db.Exec("TRUNCATE todos RESTART IDENTITY CASCADE")
	return err
}

func TestCreateTodo(t *testing.T) {
	repo := TodoPostgresRepository{db: testDbInstance}

	testCases := []struct {
		name          string
		todo          *todo.Todo
		setup         func()
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
			setup: func() {
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
			if tc.setup != nil {
				tc.setup()
			}
			id, err := repo.Create(tc.todo)
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedId, id)

			err = cleanupDatabase(testDbInstance)
			assert.NoError(t, err)
		})
	}
}

func TestGetByIdTodo(t *testing.T) {
	repo := TodoPostgresRepository{db: testDbInstance}
	id, err := repo.Create(createTestTodo())
	assert.NoError(t, err)
	todo, err := repo.GetById(id)
	assert.NoError(t, err)
	assert.NotNil(t, todo)
}
