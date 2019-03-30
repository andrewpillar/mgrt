package database

import "strings"

func (db *DB) initPostgres() error {
	_, err := db.Exec(`
		CREATE TABLE mgrt_revisions (
			id         INT NOT NULL,
			author     VARCHAR NOT NULL,
			message    TEXT NOT NULL,
			hash       BYTEA NOT NULL,
			direction  INT NOT NULL,
			forced     BOOLEAN NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
	`)

	if err != nil && strings.Contains(err.Error(), "already exists") {
		return ErrInitialized
	}

	return nil
}
