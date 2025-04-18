package db

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteDatebase struct {
	db *sqlx.DB
}

func newSqlite(url string) (*SqliteDatebase, error) {
	db, err := sqlx.Open("sqlite3", url)
	if err != nil {
		return nil, err
	}
	// Execute pragmas
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA cache_size=-2000;",
		// Optional additional pragmas for performance
		"PRAGMA synchronous=NORMAL;", // Less durable but faster
		"PRAGMA temp_store=MEMORY;",  // Store temp tables in memory
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to execute %s: %w", pragma, err)
		}
	}
	return &SqliteDatebase{db: db}, nil
}

func (s *SqliteDatebase) GetDb() *sqlx.DB {
	return s.db
}

func (s *SqliteDatebase) GetNativeDb() *sql.DB {
	return s.db.DB
}
