// Package integration provides helper functions for integration tests
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	postgrescontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDatabase provides a PostgreSQL test container for integration tests
type TestDatabase struct {
	Container *postgrescontainers.PostgresContainer
	DB        *sql.DB
}

// NewTestDatabase creates a new PostgreSQL test container
func NewTestDatabase(ctx context.Context) (*TestDatabase, error) {
	// Start PostgreSQL container
	container, err := postgrescontainers.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	// Get database connection
	dbPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	dbHost, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable",
		dbHost, dbPort.Port())

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection
	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &TestDatabase{
		Container: container,
		DB:        db,
	}, nil
}

// RunMigrations executes database migrations
func (td *TestDatabase) RunMigrations(ctx context.Context) error {
	migrationsDir := "../../migrations"

	// Read migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		path := filepath.Join(migrationsDir, file.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		_, err = td.DB.ExecContext(ctx, string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
		}
	}

	return nil
}

// Close closes the database connection and terminates the container
func (td *TestDatabase) Close(ctx context.Context) error {
	var errs []error

	if td.DB != nil {
		if err := td.DB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database: %w", err))
		}
	}

	if td.Container != nil {
		if err := td.Container.Terminate(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to terminate container: %w", err))
		}
	}

	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("Error during cleanup: %v", err)
		}
		return errs[0]
	}

	return nil
}
