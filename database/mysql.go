package database

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/revision"

	"github.com/go-sql-driver/mysql"
)

type MySQL struct {
	*database
}

var (
	mysqlDsn = "%s:%s@%s/%s?parseTime=true"

	mysqlInit = `
CREATE TABLE mgrt_revisions (
	id         INT NOT NULL,
	message    TEXT NOT NULL,
	hash       BLOB NOT NULL,
	direction  INT NOT NULL,
	up         TEXT NULL,
	down       TEXT NULL,
	forced     TINYINT NOT NULL,
	created_at TIMESTAMP NOT NULL
);`

	mysqlLastQuery = `
SELECT id, hash, direction
FROM mgrt_revisions WHERE id = ?
ORDER BY created_at DESC LIMIT 1`

	mysqlLogQuery = `
INSERT INTO mgrt_revisions
(id, message, hash, direction, up, down, forced, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	mysqlHashQuery = `
SELECT hash
FROM mgrt_revisions
WHERE id = ? AND direction = ? AND forced = false
ORDER BY created_at DESC LIMIT 1`
)

func init() {
	databases["mysql"] = &MySQL{}
}

func (p *MySQL) FromConn(db *sql.DB) {
	p.database = &database{
		DB: db,
	}
}

func (m *MySQL) Open(cfg *config.Config) error {
	dsn := fmt.Sprintf(mysqlDsn, cfg.Username, cfg.Password, cfg.Address, cfg.Database)

	if cfg.SSL.Mode != "" {
		dsn += "&tls=" + cfg.SSL.Mode

		if cfg.SSL.Mode == "custom" {
			pool := x509.NewCertPool()

			pem, err := ioutil.ReadFile(cfg.SSL.Root)

			if err != nil {
				return err
			}

			if ok := pool.AppendCertsFromPEM(pem); !ok {
				return errors.New("failed to add certificate from PEM, check your root cert")
			}

			pair, err := tls.LoadX509KeyPair(cfg.SSL.Cert, cfg.SSL.Key)

			if err != nil {
				return err
			}

			mysql.RegisterTLSConfig("custom", &tls.Config{
				RootCAs:      pool,
				Certificates: []tls.Certificate{pair},
			})
		}
	}

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return err
	}

	m.database = &database{
		DB: db,
	}

	return nil
}

func (m *MySQL) Init() error {
	_, err := m.database.Exec(mysqlInit)

	if err != nil && strings.Contains(err.Error(), "already exists") {
		return ErrInitialized
	}

	return err
}

func (m *MySQL) Log(r *revision.Revision, forced bool) error {
	return m.database.log(r, forced, mysqlLogQuery)
}

func (m *MySQL) Perform(r *revision.Revision, forced bool) error {
	return m.database.perform(r, forced, mysqlLastQuery, mysqlHashQuery)
}
