package main

import (
	"database/sql"
	"fmt"
	"github.com/GlebMoskalev/todo-api/internal/database"
	"github.com/GlebMoskalev/todo-api/internal/models/priority"
	"github.com/GlebMoskalev/todo-api/internal/models/status"
	"github.com/GlebMoskalev/todo-api/internal/models/todo"
	"github.com/GlebMoskalev/todo-api/internal/repository"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
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
	id, _ := todoRepo.Create(&todo.Todo{
		Title:       "hah",
		Description: "lol",
		DueDate: sql.NullTime{
			Valid: false,
		},
		Tags:     []string{"api"},
		Priority: priority.High,
		Status:   status.Planned,
		Overdue:  false,
	})
	//t, err := todoRepo.GetAll([]string{}, "", "", nil, sql.NullTime{
	//	Valid: false,
	//})
	//for _, td := range t {
	//	fmt.Println(*td)
	//}
	f, _ := todoRepo.GetById(id)
	fmt.Println((*f).DueDate)
}
