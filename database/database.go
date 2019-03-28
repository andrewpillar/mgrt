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

	_ "github.com/go-sql-driver/mysql"
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
	mysqlSource    = "%s:%s@%s/%s"


)

type Type uint32

type DB struct {
	*sql.DB

	Type Type
}

func Open(cfg *config.Config) (*DB, error) {
	if cfg.Type == "sqlite3" {
		db, err := sql.Open(cfg.Type, cfg.Address)

		if err != nil {
			return nil, err
		}

		return &DB{DB: db, Type: SQLite3}, nil
	}

	var typ Type
	var source string

	switch cfg.Type {
		case "postgres":
			host, port, err := net.SplitHostPort(cfg.Address)

			if err != nil {
				return nil, err
			}

			typ = Postgres
			source = fmt.Sprintf(postgresSource, host, port, cfg.Username, cfg.Database, cfg.Password)
			break
		case "mysql":
			typ = MySQL
			source = fmt.Sprintf(mysqlSource, cfg.Username, cfg.Password, cfg.Address, cfg.Database)
			break
		default:
			return nil, errors.New("unknown database type " + cfg.Type)
	}

	db, err := sql.Open(cfg.Type, source)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{DB: db, Type: typ}, nil
}

func (db *DB) Init() error {
	switch db.Type {
		case SQLite3:
			return db.initSqlite3()
		case Postgres:
			return db.initPostgres()
		case MySQL:
			return db.initMysql()
		default:
			return errors.New("unknown database type")
	}
}

func (db *DB) Log(r *revision.Revision, forced bool) error {
	var stmt *sql.Stmt
	var err error

	switch db.Type {
		case SQLite3:
		case Postgres:
			stmt, err = db.Prepare(`
				INSERT INTO mgrt_revisions (id, author, hash, direction, forced, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`)
			break
		case MySQL:
			stmt, err = db.Prepare(`
				INSERT INTO mgrt_revisions (id, author, hash, direction, forced, created_at)
				VALUES (?, ?, ?, ?, ?, ?)
			`)
			break
		default:
			err = errors.New("unknown database type")
			break
	}

	if err != nil {
		return err
	}

	defer stmt.Close()

	blob := make([]byte, len(r.Hash), len(r.Hash))

	for i, b := range r.Hash {
		blob[i] = b
	}

	_, err = stmt.Exec(r.ID, r.Author, blob, r.Direction, forced, time.Now())

	return err
}

func (db *DB) ReadLogReverse(ids ...string) ([]*revision.Revision, error) {
	query := "SELECT id, author, hash, direction, forced, created_at FROM mgrt_revisions"

	if len(ids) > 0 {
		query += " WHERE id IN(" + strings.Join(ids, ", ") + ")"
	}

	query += " ORDER BY created_at ASC"

	return db.realReadLog(query)
}

func (db *DB) ReadLog(ids ...string) ([]*revision.Revision, error) {
	query := "SELECT id, author, hash, direction, forced, created_at FROM mgrt_revisions"

	if len(ids) > 0 {
		query += " WHERE id IN (" + strings.Join(ids, ", ") + ")"
	}

	query += " ORDER BY created_at DESC"

	return db.realReadLog(query)
}

func (db *DB) realReadLog(query string) ([]*revision.Revision, error) {
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

		err := rows.Scan(&r.ID, &r.Author, &blob, &r.Direction, &r.Forced, &r.CreatedAt)

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

func (db *DB) Perform(r *revision.Revision, force bool) error {
	var stmt *sql.Stmt
	var err error

	switch db.Type {
		case SQLite3:
		case Postgres:
			stmt, err = db.Prepare(`
				SELECT id, hash, direction
				FROM mgrt_revisions WHERE id = $1
				ORDER BY created_at DESC LIMIT 1
			`)
			break
		case MySQL:
			stmt, err = db.Prepare(`
				SELECT id, hash, direction
				FROM mgrt_revisions WHERE id = ?
				ORDER BY created_at DESC LIMIT 1
			`)
			break
		default:
			err = errors.New("unknown database type")
			break
	}

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

		if r.Hash != performed.Hash && !force {
			return ErrChecksumFailed
		}
	}

	_, err = db.Exec(r.Query())

	return err
}
