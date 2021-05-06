// +build sqlite3

package internal

import (
	"database/sql"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var sqlite3Init = `CREATE TABLE mgrt_revisions (
	id           VARCHAR NOT NULL,
	author       VARCHAR NOT NULL,
	comment      TEXT NOT NULL,
	sql          TEXT NOT NULL,
	performed_at INT NOT NULL
);`

func init() {
	registerDB("sqlite3", openSqlite3)
}

func openSqlite3(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)

	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(sqlite3Init); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return nil, err
		}
	}
	return db, nil
}
