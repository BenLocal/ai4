package db

import (
	"database/sql"
	"errors"
)

type Datebase interface {
	GetDb() *sql.DB
}

func DatebaseFactory(dbType, url string) (Datebase, error) {
	switch dbType {
	case "sqlite":
		return newSqlite(url)
	default:
		return nil, errors.New("unsupported database type: " + dbType)
	}
}
