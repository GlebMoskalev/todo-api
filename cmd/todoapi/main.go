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
	"log/slog"
	"os"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	logger := setupLogger()
	logger.Info("Starting todo-api...")

	db, err := database.InitPostgres()
	if err != nil {
		logger.Error("Error initializing PostgreSQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Error closing database connection", slog.String("error", err.Error()))
		} else {
			logger.Info("Database connection closed.")
		}
	}()

	logger.Info("Database connection established successfully.")

	todoRepo := repository.NewTodoPostgresRepository(db, logger)
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

func setupLogger() *slog.Logger {
	var level slog.Level
	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	handlerOpts := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, handlerOpts))
	logger.Debug("Logger initialized", slog.String("level", os.Getenv("LOG_LEVEL")))
	return logger
}
