package database

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/revision"
)

var (
	databases = make(map[string]DB)

	ErrInitialized      = errors.New("database already initialized")
	ErrAlreadyPerformed = errors.New("already performed revision")
	ErrCheckHashFailed  = errors.New("revision hash check failed")
)

type database struct {
	*sql.DB
}

type DB interface {
	Open(cfg *config.Config) error

	Init() error

	Perform(r *revision.Revision, forced bool) error

	Log(r *revision.Revision, forced bool) error

	ReadLog(ids ...string) ([]*revision.Revision, error)

	ReadLogReverse(ids ...string) ([]*revision.Revision, error)

	Close() error
}

func Open(cfg *config.Config) (DB, error) {
	db, ok := databases[cfg.Type]

	if !ok {
		return nil, errors.New("unknown database type: " + cfg.Type)
	}

	err := db.Open(cfg)

	return db, err
}

func (db *database) log(r *revision.Revision, forced bool, query string) error {
	stmt, err := db.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	hash := make([]byte, len(r.Hash), len(r.Hash))

	for i := range hash {
		hash[i] = r.Hash[i]
	}

	_, err = stmt.Exec(r.ID, r.Message, hash, r.Direction, r.Up, r.Down, forced, time.Now())

	return err
}

func (db *database) ReadLogReverse(ids ...string) ([]*revision.Revision, error) {
	query := "SELECT * FROM mgrt_revisions"

	if len(ids) > 0 {
		query += " WHERE id IN(" + strings.Join(ids, ", ") + ")"
	}

	query += " ORDER BY created_at ASC"

	return db.realReadLog(query)
}

func (db *database) ReadLog(ids ...string) ([]*revision.Revision, error) {
	query := "SELECT * FROM mgrt_revisions"

	if len(ids) > 0 {
		query += " WHERE id IN (" + strings.Join(ids, ", ") + ")"
	}

	query += " ORDER BY created_at DESC"

	return db.realReadLog(query)
}

func (db *database) realReadLog(query string) ([]*revision.Revision, error) {
	stmt, err := db.Prepare(query)

	if err != nil {
		return []*revision.Revision{}, err
	}

	defer stmt.Close()

	rows, err := stmt.Query()

	if err != nil && err != sql.ErrNoRows {
		return []*revision.Revision{}, err
	}

	revisions := make([]*revision.Revision, 0)

	if err == sql.ErrNoRows {
		return revisions, nil
	}

	for rows.Next() {
		r := &revision.Revision{}

		hash := []byte{}

		err := rows.Scan(
			&r.ID,
			&r.Message,
			&hash,
			&r.Direction,
			&r.Up,
			&r.Down,
			&r.Forced,
			&r.CreatedAt,
		)

		if err != nil {
			return []*revision.Revision{}, err
		}

		for i := range r.Hash {
			r.Hash[i] = hash[i]
		}

		revisions = append(revisions, r)
	}

	return revisions, nil
}

func (db *database) Close() error {
	return db.DB.Close()
}

func (db *database) perform(r *revision.Revision, forced bool, lastQuery, hashQuery string) error {
	lastStmt, err := db.Prepare(lastQuery)

	if err != nil {
		return err
	}

	defer lastStmt.Close()

	var (
		prev revision.Revision
		hash []byte
	)

	err = lastStmt.QueryRow(r.ID).Scan(&prev.ID, &hash, &prev.Direction)

	if err != nil {
		if err == sql.ErrNoRows {
			_, err = db.Exec(r.Query())

			return err
		}

		return err
	}

	if r.Direction == prev.Direction {
		return ErrAlreadyPerformed
	}

	hashStmt, err := db.Prepare(hashQuery)

	if err != nil {
		return err
	}

	defer hashStmt.Close()

	err = hashStmt.QueryRow(r.ID, r.Direction).Scan(&hash)

	if err != nil {
		if err == sql.ErrNoRows {
			_, err = db.Exec(r.Query())

			return err
		}

		return err
	}

	for i := range hash {
		if hash[i] != r.Hash[i] && !forced {
			return ErrCheckHashFailed
		}
	}

	_, err = db.Exec(r.Query())

	return err
}
