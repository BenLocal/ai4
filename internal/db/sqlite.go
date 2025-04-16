package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteDatebase struct {
	db *sql.DB
}

func newSqlite(url string) (*SqliteDatebase, error) {
	db, err := sql.Open("sqlite3", url)
	if err != nil {
		return nil, err
	}
	return &SqliteDatebase{db: db}, nil
}

func (s *SqliteDatebase) GetDb() *sql.DB {
	return s.db
}
