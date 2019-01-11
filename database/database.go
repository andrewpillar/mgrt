package database

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/revision"

	_ "github.com/mattn/go-sqlite3"
)

const (
	SQLite3 Type = iota
	Postgres
	MySQL
)

var (
	ErrInitialized    = errors.New("database already initialized")
	ErrAlreadyRan     = errors.New("already ran revision")
	ErrChecksumFailed = errors.New("revision checksum failed")
)

type Type uint32

type DB struct {
	*sql.DB

	name  string
	table string

	Type Type
}

func checksum(a [sha256.Size]byte, b []byte) bool {
	if len(b) != sha256.Size {
		return false
	}

	for i := 0; i < sha256.Size; i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
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
		default:
			err = errors.New("unknown database " + cfg.Type)
			break
	}

	return &DB{
		name:  cfg.Database.Name,
		table: cfg.Database.Table,
		DB:    db,
		Type:  typ,
	}, err
}

func (d *DB) Init() error {
	if d.Type == SQLite3 {
		stmt, err := d.Prepare(`
			SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = $1
		`)

		if err != nil {
			return err
		}

		defer stmt.Close()

		var count int

		row := stmt.QueryRow(d.table)

		if err := row.Scan(&count); err != nil {
			return err
		}

		if count > 0 {
			return ErrInitialized
		}

		query := fmt.Sprintf(initSqlite3f, d.table)

		_, err = d.Exec(query)

		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (d *DB) Log(r *revision.Revision, dir revision.Direction) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (revision, hash, direction, created_at) VALUES ($1, $2, $3, $4)
	`, d.table)

	if d.Type == SQLite3 {
		return d.logSqlite3(query, r, dir)
	}

	return nil
}

func (d *DB) Run(r *revision.Revision, dir revision.Direction) error {
	query := fmt.Sprintf(`
		SELECT * FROM %s WHERE revision = $1 ORDER BY created_at DESC LIMIT 1
	`, d.table)

	stmt, err := d.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	row := stmt.QueryRow(r.ID)

	if d.Type == SQLite3 {
		return d.runSqlite3(row, r, dir)
	}

	return err
}
