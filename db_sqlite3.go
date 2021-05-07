// +build sqlite3

package mgrt

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
	Register("sqlite3", &DB{
		Type:         "sqlite3",
		Init:         initSqlite3,
		Parameterize: func(s string) string { return s },
	})
}

func initSqlite3(db *sql.DB) error {
	if _, err := db.Exec(sqlite3Init); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}
