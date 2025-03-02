package main

import (
	"fmt"
	"github.com/GlebMoskalev/todo-api/internal/database"
	"github.com/GlebMoskalev/todo-api/internal/models/priority"
	"github.com/GlebMoskalev/todo-api/internal/models/status"
	"github.com/GlebMoskalev/todo-api/internal/models/todo"
	"github.com/GlebMoskalev/todo-api/internal/repository"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"time"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	db, err := database.InitPostgres()
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer db.Close()

	todoRepo := repository.NewTodoPostgresRepository(db)
	id, err := todoRepo.Create(&todo.Todo{
		Title:       "hah",
		Description: "lol",
		DueDate:     time.Time{},
		Tags:        []string{"api"},
		Priority:    priority.Low,
		Status:      status.InProgress,
		Overdue:     false,
	})
	//t, err := todoRepo.GetAll([]string{}, status.Canceled, priority.High, false, time.Time{})
	//err = todoRepo.Delete([]int{0, 1, 2, 3})
	fmt.Println(id, err)
}
