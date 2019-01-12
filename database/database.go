package database

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/revision"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const (
	SQLite3 Type = iota
	Postgres
	MySQL
)

var (
	ErrInitialized      = errors.New("database already initialized")
	ErrAlreadyPerformed = errors.New("already performed revision")
	ErrChecksumFailed   = errors.New("revision checksum failed")

	postgresSource = "host=%s port=%s user=%s dbname=%s password=%s sslmode=disable"
)

type Type uint32

type DB struct {
	*sql.DB

	Type Type
}

func Open(cfg *config.Config) (*DB, error) {
	var db *sql.DB
	var typ Type
	var err error

	switch cfg.Type {
		case "sqlite3":
			db, err = sql.Open(cfg.Type, cfg.Address)
			typ = SQLite3
			break
		case "postgres":
			host, port, e:= net.SplitHostPort(cfg.Address)

			if e != nil {
				err = e
				break
			}

			source := fmt.Sprintf(postgresSource, host, port, cfg.Username, cfg.Database, cfg.Password)

			db, err = sql.Open(cfg.Type, source)
			typ = Postgres
			break
		default:
			err = errors.New("unknown database type " + cfg.Type)
			break
	}

	return &DB{
		DB:    db,
		Type:  typ,
	}, err
}

func (db *DB) Init() error {
	switch db.Type {
		case SQLite3:
			return db.initSqlite3()
		case Postgres:
			return db.initPostgres()
		default:
			return errors.New("unknown database type")
	}
}

func (db *DB) Log(r *revision.Revision) error {
	stmt, err := db.Prepare(`
		INSERT INTO mgrt_revisions (id, hash, direction, created_at)
		VALUES ($1, $2, $3, $4)
	`)

	if err != nil {
		return err
	}

	defer stmt.Close()

	blob := make([]byte, len(r.Hash), len(r.Hash))

	for i, b := range r.Hash {
		blob[i] = b
	}

	_, err = stmt.Exec(r.ID, blob, r.Direction, time.Now())

	return err
}

func (db *DB) ReadLog(ids ...string) ([]*revision.Revision, error) {
	query := "SELECT id, hash, direction, created_at FROM mgrt_revisions"

	if len(ids) > 0 {
		query += " WHERE id IN (" + strings.Join(ids, ", ") + ")"
	}

	query += " ORDER BY created_at DESC"

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
		var blob []byte

		r := &revision.Revision{}

		err := rows.Scan(&r.ID, &blob, &r.Direction, &r.CreatedAt)

		for i := range r.Hash {
			r.Hash[i] = blob[i]
		}

		if err != nil {
			return []*revision.Revision{}, err
		}

		if err := r.Load(); err != nil {
			continue
		}

		revisions = append(revisions, r)
	}

	return revisions, nil
}

func (db *DB) Perform(r *revision.Revision) error {
	stmt, err := db.Prepare(`
		SELECT id, hash, direction
		FROM mgrt_revisions WHERE id = $1
		ORDER BY created_at DESC LIMIT 1
	`)

	if err != nil {
		return err
	}

	defer stmt.Close()

	var performed revision.Revision
	var blob []byte

	row := stmt.QueryRow(r.ID)

	err = row.Scan(&performed.ID, &blob, &performed.Direction)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if len(blob) > 0 {
		for i := range performed.Hash {
			performed.Hash[i] = blob[i]
		}
	}

	if err == nil {
		if r.Direction == performed.Direction {
			return ErrAlreadyPerformed
		}

		if r.Hash != performed.Hash {
			return ErrChecksumFailed
		}
	}

	_, err = db.Exec(r.Query())

	return err
}
