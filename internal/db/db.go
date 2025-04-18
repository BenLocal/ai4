package db

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Datebase interface {
	GetDb() *sqlx.DB
	GetNativeDb() *sql.DB
}

func DatebaseFactory(dbType, url string) (Datebase, error) {
	switch dbType {
	case "sqlite":
		return newSqlite(url)
	default:
		return nil, errors.New("unsupported database type: " + dbType)
	}
}
