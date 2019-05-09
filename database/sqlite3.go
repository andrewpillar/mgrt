package database

import "strings"

func (db *DB) initSqlite3() error {
	_, err := db.Exec(`
		CREATE TABLE mgrt_revisions (
			id         INTEGER NOT NULL,
			author     TEXT NOT NULL,
			message    TEXT NOT NULL,
			hash       BLOB NOT NULL,
			direction  INTEGER NOT NULL,
			up         TEXT NULL,
			down       TEXT NULL,
			forced     INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
	`)

	if err != nil && strings.Contains(err.Error(), "already exists") {
		return ErrInitialized
	}

	return err
}
