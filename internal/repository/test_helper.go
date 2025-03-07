package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"path/filepath"
	"runtime"
	"time"
)

const (
	masterDbName = "master_db"
	dbUser       = "test_user"
	dbPassword   = "test_password"
)

type TestDatabase struct {
	DbAddress string
	container testcontainers.Container
}

func SetupMasterDatabase() *TestDatabase {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	container, dbInstance, dbAddr, err := createContainer(ctx, masterDbName)
	if err != nil {
		log.Fatal("failed to prepareData test", err)
	}

	err = migrateDb(dbAddr)
	if err != nil {
		log.Fatal("failed to perform db migration", err)
	}
	cancel()
	dbInstance.Close()
	return &TestDatabase{
		DbAddress: dbAddr,
		container: container,
	}
}

func SetupTestDatabase(masterAddr string, testDbName string) (*sql.DB, error) {
	masterConnStr := fmt.Sprintf("postgres://%s:%s@%s/postgres?sslmode=disable", dbUser, dbPassword, masterAddr)
	masterDb, err := sql.Open("postgres", masterConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %v", err)
	}
	defer masterDb.Close()
	query := fmt.Sprintf("CREATE DATABASE %s WITH TEMPLATE %s", testDbName, masterDbName)
	_, err = masterDb.Exec(query)
	if err != nil {
		return nil, fmt.Errorf("failed to create test database: %v", err)
	}

	testConnStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, masterAddr, testDbName)
	testDb, err := sql.Open("postgres", testConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %v", err)
	}

	err = testDb.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	return testDb, nil
}

func TearDownTestDatabase(masterAddr string, testDbName string) error {
	masterConnStr := fmt.Sprintf("postgres://%s:%s@%s/postgres?sslmode=disable", dbUser, dbPassword, masterAddr)
	masterDb, err := sql.Open("postgres", masterConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %v", err)
	}
	defer masterDb.Close()

	_, err = masterDb.Exec(fmt.Sprintf("DROP DATABASE %s", testDbName))
	return err
}

func createContainer(ctx context.Context, dbName string) (testcontainers.Container, *sql.DB, string, error) {
	var env = map[string]string{
		"POSTGRES_PASSWORD": dbPassword,
		"POSTGRES_USER":     dbUser,
		"POSTGRES_DB":       dbName,
	}
	var port = "5432/tcp"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:17-alpine",
			ExposedPorts: []string{port},
			Env:          env,
			WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(ctx, req)

	if err != nil {
		return container, nil, "", fmt.Errorf("failed to start container: %v", err)
	}

	p, err := container.MappedPort(ctx, "5432")

	log.Println("postgres container ready and running at port: ", p.Port())

	time.Sleep(time.Second)
	dbAddr := fmt.Sprintf("localhost:%s", p.Port())
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbAddr, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return container, db, dbAddr, fmt.Errorf("failed to establish database connection: %v", err)
	}

	err = db.Ping()

	if err != nil {
		return container, db, dbAddr, fmt.Errorf("failed to ping database: %v", err)
	}

	return container, db, dbAddr, nil
}

func migrateDb(dbAddr string) error {
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get path")
	}

	pathToRepository := filepath.Dir(path)
	pathToMigrationFiles := filepath.Join(pathToRepository, "../../migrations")
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbAddr, masterDbName)

	m, err := migrate.New(fmt.Sprintf("file:%s", pathToMigrationFiles), databaseURL)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("migration done")

	return nil
}
