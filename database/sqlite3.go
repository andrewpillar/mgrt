package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/andrewpillar/mgrt/revision"
)

var initSqlite3f = `CREATE TABLE %s (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	revision   INTEGER NOT NULL,
	hash       TEXT NOT NULL,
	direction  TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL
);`

func (d *DB) logSqlite3(query string, r *revision.Revision, dir revision.Direction) error {
	stmt, err := d.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(r.ID, fmt.Sprintf("%x", r.Hash), dir.String(), time.Now())

	return err
}

func (d *DB) runSqlite3(row *sql.Row, r *revision.Revision, dir revision.Direction) error {
	scanned := struct{
		id        int
		revision  int
		hash      string
		direction string
		createdAt time.Time
	}{}

	err := row.Scan(
		&scanned.id,
		&scanned.revision,
		&scanned.hash,
		&scanned.direction,
		&scanned.createdAt,
	)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == nil {
		if scanned.direction == dir.String() {
			return errors.New("already ran revision " + r.ID)
		}

		if !checksum(r.Hash, []byte(scanned.hash)) {
			return errors.New("revision checksum failed")
		}
	}

	var query string

	if dir == revision.Up {
		query = r.Up
	}

	if dir == revision.Down {
		query = r.Down
	}

	_, err = d.Exec(query)

	if err != nil {
		return err
	}

	return nil
}
