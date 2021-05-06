package internal

import (
	"database/sql"
	"errors"
	"sync"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type openFunc func(string) (*sql.DB, error)

var (
	openMu    sync.RWMutex
	openfuncs = make(map[string]openFunc)

	postgresInit = `CREATE TABLE mgrt_revisions (
	id           VARCHAR NOT NULL UNIQUE,
	author       VARCHAR NOT NULL,
	comment      TEXT NOT NULL,
	query        TEXT NOT NULL,
	performed_at INT NOT NULL
);`
)

func init() {
	registerDB("postgresql", openPostgreSQL)
}

func registerDB(typ string, fn openFunc) {
	openMu.Lock()
	defer openMu.Unlock()

	if fn == nil {
		panic("nil database open function")
	}

	if _, ok := openfuncs[typ]; ok {
		panic("database open function already registered")
	}
	openfuncs[typ] = fn
}

func openDB(typ, dsn string) (*sql.DB, error) {
	openMu.RLock()
	defer openMu.RUnlock()

	open, ok := openfuncs[typ]

	if !ok {
		return nil, errors.New("unknown database type " + typ)
	}
	return open(dsn)
}

func openPostgreSQL(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(postgresInit); err != nil {
		return nil, err
	}
	return db, nil
}
