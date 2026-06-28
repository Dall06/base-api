package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Connect establishes a connection to the database with retry logic
func Connect(databaseURI string, timeoutSec int) (*bun.DB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(databaseURI)))

	// Try initial connection
	if err := sqldb.Ping(); err == nil {
		db := bun.NewDB(sqldb, pgdialect.New())
		return db, nil
	}

	// Retry with timeout
	start := time.Now()
	timeout := time.Duration(timeoutSec) * time.Second
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := sqldb.Ping(); err == nil {
			db := bun.NewDB(sqldb, pgdialect.New())
			return db, nil
		}

		if time.Since(start) > timeout {
			return nil, fmt.Errorf("failed to connect to database after %d seconds", timeoutSec)
		}
	}

	return nil, fmt.Errorf("database connection failed")
}
