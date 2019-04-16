package database

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/revision"

	"github.com/go-sql-driver/mysql"

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
	ErrCheckHashFailed  = errors.New("revision hash check failed")

	postgresSource = "host=%s port=%s user=%s dbname=%s password=%s sslmode=%s"
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

			if cfg.SSL.Mode != "disable" {
				if cfg.SSL.Cert != "" {
					postgresSource += " sslcert=" + cfg.SSL.Cert
				}

				if cfg.SSL.Key != "" {
					postgresSource += " sslkey=" + cfg.SSL.Key
				}

				if cfg.SSL.Root != "" {
					postgresSource += " sslrootcert=" + cfg.SSL.Root
				}
			}

			source = fmt.Sprintf(
				postgresSource,
				host,
				port,
				cfg.Username,
				cfg.Database,
				cfg.Password,
				cfg.SSL.Mode,
			)
			break
		case "mysql":
			typ = MySQL

			source = fmt.Sprintf(
				mysqlSource,
				cfg.Username,
				cfg.Password,
				cfg.Address,
				cfg.Database,
			)

			if cfg.SSL.Mode == "custom" {
				source += "?tls=" + cfg.SSL.Mode

				pool := x509.NewCertPool()

				pem, err := ioutil.ReadFile(cfg.SSL.Root)

				if err != nil {
					return nil, err
				}

				if ok := pool.AppendCertsFromPEM(pem); !ok {
					return nil, err
				}

				pair, err := tls.LoadX509KeyPair(cfg.SSL.Cert, cfg.SSL.Key)

				if err != nil {
					return nil, err
				}

				mysql.RegisterTLSConfig("custom", &tls.Config{
					RootCAs:      pool,
					Certificates: []tls.Certificate{pair},
				})
			}

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
			fallthrough
		case Postgres:
			stmt, err = db.Prepare(`
				INSERT INTO mgrt_revisions (id, author, message, hash, direction, forced, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`)
			break
		case MySQL:
			stmt, err = db.Prepare(`
				INSERT INTO mgrt_revisions (id, author, message, hash, direction, forced, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?)
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

	hash := make([]byte, len(r.Hash), len(r.Hash))

	for i := range hash {
		hash[i] = r.Hash[i]
	}

	_, err = stmt.Exec(r.ID, r.Author, r.Message, hash, r.Direction, forced, time.Now())

	return err
}

func (db *DB) ReadLogReverse(ids ...string) ([]*revision.Revision, error) {
	query := "SELECT id, author, message, hash, direction, forced, created_at FROM mgrt_revisions"

	if len(ids) > 0 {
		query += " WHERE id IN(" + strings.Join(ids, ", ") + ")"
	}

	query += " ORDER BY created_at ASC"

	return db.realReadLog(query)
}

func (db *DB) ReadLog(ids ...string) ([]*revision.Revision, error) {
	query := "SELECT id, author, message, hash, direction, forced, created_at FROM mgrt_revisions"

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
		r := &revision.Revision{}

		hash := []byte{}

		err := rows.Scan(&r.ID, &r.Author, &r.Message, &hash, &r.Direction, &r.Forced, &r.CreatedAt)

		if err != nil {
			return []*revision.Revision{}, err
		}

		for i := range r.Hash {
			r.Hash[i] = hash[i]
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
			fallthrough
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

	var lastRevision revision.Revision
	var hash []byte

	row := stmt.QueryRow(r.ID)

	err = row.Scan(&lastRevision.ID, &hash, &lastRevision.Direction)

	if err == sql.ErrNoRows {
		_, err = db.Exec(r.Query())

		return err
	}

	for i := range lastRevision.Hash {
		lastRevision.Hash[i] = hash[i]
	}

	if r.Direction == lastRevision.Direction {
		return ErrAlreadyPerformed
	}

	switch db.Type {
		case SQLite3:
			fallthrough
		case Postgres:
			stmt, err = db.Prepare(`
				SELECT hash
				FROM mgrt_revisions
				WHERE id = $1 AND direction = $2 AND forced = false
				ORDER BY created_at DESC LIMIT 1
			`)
		case MySQL:
			stmt, err = db.Prepare(`
				SELECT hash
				FROM mgrt_revisions
				WHERE id = ? AND direction = ? AND forced = false
				ORDER BY created_at DESC LIMIT 1
			`)
		default:
			err = errors.New("unknown database type")
			break
	}

	if err != nil {
		return err
	}

	row = stmt.QueryRow(r.ID, r.Direction)

	err = row.Scan(&hash)

	if err == sql.ErrNoRows {
		_, err = db.Exec(r.Query())

		return err
	}

	if err != nil {
		return err
	}

	for i := range hash {
		if hash[i] != r.Hash[i] && !force {
			return ErrCheckHashFailed
		}
	}

	_, err = db.Exec(r.Query())

	return err
}
