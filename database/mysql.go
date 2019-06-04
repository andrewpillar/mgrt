package database

import "strings"

func (db *DB) initMysql() error {
	_, err := db.Exec(`
		CREATE TABLE mgrt_revisions (
			id         INT NOT NULL,
			message    TEXT NOT NULL,
			hash       BLOB NOT NULL,
			direction  INT NOT NULL,
			up         TEXT NULL,
			down       TEXT NULL,
			forced     TINYINT NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
	`)

	if err != nil && strings.Contains(err.Error(), "already exists") {
		return ErrInitialized
	}

	return err
}
