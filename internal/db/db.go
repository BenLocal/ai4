package db

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

type Datebase interface {
	GetDb() *sqlx.DB
}

func DatebaseFactory(dbType, url string) (Datebase, error) {
	switch dbType {
	case "sqlite":
		return newSqlite(url)
	default:
		return nil, errors.New("unsupported database type: " + dbType)
	}
}
