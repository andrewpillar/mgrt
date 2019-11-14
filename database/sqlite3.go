// +build sqlite3

package database

import (
	"database/sql"
	"strings"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/revision"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite3 struct {
	*database
}

var sqlite3Init = `
CREATE TABLE mgrt_revisions (
	id         INTEGER NOT NULL,
	message    TEXT NOT NULL,
	hash       BLOB NOT NULL,
	direction  INTEGER NOT NULL,
	up         TEXT NULL,
	down       TEXT NULL,
	forced     INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL
);`

func init() {
	databases["sqlite3"] = &SQLite3{}
}

func (s *SQLite3) Open(cfg *config.Config) error {
	db, err := sql.Open("sqlite3", cfg.Address)

	if err != nil {
		return err
	}

	s.database = &database{
		DB: db,
	}

	return nil
}

func (s *SQLite3) Init() error {
	_, err := s.database.Exec(sqlite3Init)

	if err != nil && strings.Contains(err.Error(), "already exists") {
		return ErrInitialized
	}

	return err
}

func (s *SQLite3) Log(r *revision.Revision, forced bool) error {
	return s.database.log(r, forced, postgresLogQuery)
}

func (s *SQLite3) Perform(r *revision.Revision, forced bool) error {
	return s.database.perform(r, forced, postgresLastQuery, postgresHashQuery)
}
