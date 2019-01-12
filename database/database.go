package database

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
	ErrInitialized      = errors.New("database already initialized")
	ErrAlreadyPerformed = errors.New("already performed revision")
	ErrChecksumFailed   = errors.New("revision checksum failed")
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
			err = errors.New("unknown database type " + cfg.Type)
			break
	}

	return &DB{
		name:  cfg.Database.Name,
		table: cfg.Database.Table,
		DB:    db,
		Type:  typ,
	}, err
}

func (db *DB) Init() error {
	switch db.Type {
		case SQLite3:
			return db.initSqlite3()
		default:
			return errors.New("unknown database type")
	}
}

func (db *DB) Log(r *revision.Revision, d revision.Direction) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, hash, direction, created_at)
		VALUES ($1, $2, $3, $4)
	`, db.table)

	stmt, err := db.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	blob := make([]byte, len(r.Hash), len(r.Hash))

	for i, b := range r.Hash {
		blob[i] = b
	}

	_, err = stmt.Exec(r.ID, blob, d, time.Now())

	return err
}

func (db *DB) Perform(r *revision.Revision, d revision.Direction) error {
	query := fmt.Sprintf(`
		SELECT id, hash, direction
		FROM %s WHERE id = $1
		ORDER BY created_at DESC LIMIT 1
	`, db.table)

	stmt, err := db.Prepare(query)

	if err != nil {
		return err
	}

	defer stmt.Close()

	var performed revision.Revision
	var direction revision.Direction
	var blob []byte

	row := stmt.QueryRow(r.ID)

	err = row.Scan(&performed.ID, &blob, &direction)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if len(blob) > 0 {
		for i := range performed.Hash {
			performed.Hash[i] = blob[i]
		}
	}

	if err == nil {
		if d == direction {
			return ErrAlreadyPerformed
		}

		if r.Hash != performed.Hash {
			return ErrChecksumFailed
		}
	}

	if d == revision.Up {
		_, err = db.Exec(r.Up)

		return err
	}

	if d == revision.Down {
		_, err = db.Exec(r.Down)

		return err
	}

	return nil
}
