package database

import (
	"database/sql"
	"fmt"
	"net"
	"strings"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/revision"

	_ "github.com/lib/pq"
)

type Postgres struct {
	*database
}

var (
	postgresDsn = "host=%s port=%s user=%s dbname=%s password=%s sslmode=%s"

	postgresInit = `
CREATE TABLE mgrt_revisions (
	id         INT NOT NULL,
	message    TEXT NOT NULL,
	hash       BYTEA NOT NULL,
	direction  INT NOT NULL,
	up         TEXT NULL,
	down       TEXT NULL,
	forced     BOOLEAN NOT NULL,
	created_at TIMESTAMP NOT NULL
);`

	postgresLastQuery = `
SELECT id, hash, direction
FROM mgrt_revisions WHERE id = $1
ORDER BY created_at DESC LIMIT 1`

	postgresLogQuery = `
INSERT INTO mgrt_revisions
(id, message, hash, direction, up, down, forced, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	postgresHashQuery = `
SELECT hash
FROM mgrt_revisions
WHERE id = $1 AND direction = $2 AND forced = false
ORDER BY created_at DESC LIMIT 1`
)

func init() {
	databases["postgres"] = &Postgres{}
}

func (p *Postgres) Open(cfg *config.Config) error {
	host, port, err := net.SplitHostPort(cfg.Address)

	if err != nil {
		return err
	}

	if cfg.SSL.Mode == "" {
		cfg.SSL.Mode = "disable"
	}

	if cfg.SSL.Mode != "disable" {
		if cfg.SSL.Cert != "" {
			postgresDsn += " sslcert=" + cfg.SSL.Cert
		}

		if cfg.SSL.Key != "" {
			postgresDsn += " sslkey=" + cfg.SSL.Key
		}

		if cfg.SSL.Root != "" {
			postgresDsn += " sslrootcert=" + cfg.SSL.Root
		}
	}

	dsn := fmt.Sprintf(postgresDsn, host, port, cfg.Username, cfg.Database, cfg.Password, cfg.SSL.Mode)

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return err
	}

	p.database = &database{
		DB: db,
	}

	return nil
}

func (p *Postgres) Init() error {
	_, err := p.database.Exec(postgresInit)

	if err != nil && strings.Contains(err.Error(), "already exists") {
		return ErrInitialized
	}

	return err
}

func (p *Postgres) Log(r *revision.Revision, forced bool) error {
	return p.database.log(r, forced, postgresLogQuery)
}

func (p *Postgres) Perform(r *revision.Revision, forced bool) error {
	return p.database.perform(r, forced, postgresLastQuery, postgresHashQuery)
}
