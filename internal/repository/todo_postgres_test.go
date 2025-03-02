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

//func TestGetByIdTodo(t *testing.T) {
//	repo := TodoPostgresRepository{db: testDbInstance}
//	id, err := repo.Create(createTestTodo())
//	assert.NoError(t, err)
//	todo, err := repo.GetById(id)
//	assert.NoError(t, err)
//	assert.NotNil(t, todo)
//}
