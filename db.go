package mgrt

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// DB is a thin abstraction over the *sql.DB struct from the stdlib.
type DB struct {
	*sql.DB

	// Type is the type of database being connected to. This will be passed to
	// sql.Open when the connection is being opened.
	Type string

	// Init is the function to call to initialize the database for performing
	// revisions.
	Init func(*sql.DB) error

	// Parameterize is the function that is called to parameterize the query
	// that will be executed against the database. This will make sure the
	// correct SQL dialect is being used for the type of database.
	Parameterize func(string) string
}

var (
	dbMu sync.RWMutex
	dbs  = make(map[string]*DB)

	mysqlInit = `CREATE TABLE mgrt_revisions (
	id           VARCHAR NOT NULL UNIQUE,
	author       VARCHAR NOT NULL,
	comment      TEXT NOT NULL,
	sql          TEXT NOT NULL,
	performed_at INT NOT NULL
);`

	postgresInit = `CREATE TABLE mgrt_revisions (
	id           VARCHAR NOT NULL UNIQUE,
	author       VARCHAR NOT NULL,
	comment      TEXT NOT NULL,
	sql          TEXT NOT NULL,
	performed_at INT NOT NULL
);`
)

func init() {
	Register("mysql", &DB{
		Type:         "mysql",
		Init:         initMysql,
		Parameterize: parameterizeMysql,
	})

	Register("postgresql", &DB{
		Type:         "pgx",
		Init:         initPostgresql,
		Parameterize: parameterizePostgresql,
	})
}

func initMysql(db *sql.DB) error {
	if _, err := db.Exec(mysqlInit); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}

func initPostgresql(db *sql.DB) error {
	if _, err := db.Exec(postgresInit); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}

func parameterizeMysql(s string) string { return s }

func parameterizePostgresql(s string) string {
	q := make([]byte, 0, len(s))
	n := int64(0)

	for i := strings.Index(s, "?"); i != -1; i = strings.Index(s, "?") {
		n++

		q = append(q, s[:i]...)
		q = append(q, '$')
		q = strconv.AppendInt(q, n, 10)

		s = s[i+1:]
	}
	return string(append(q, []byte(s)...))
}

// Register will register the given *DB for the given database type. If the
// given type is a duplicate, then this panics. If the given *DB is nil, then
// this panics.
func Register(typ string, db *DB) {
	dbMu.Lock()
	defer dbMu.Unlock()

	if db == nil {
		panic("mgrt: nil database registered")
	}

	if _, ok := dbs[typ]; ok {
		panic("mgrt: database already registered for " + typ)
	}
	dbs[typ] = db
}

// Open is a utility function that will call sql.Open with the given typ and
// dsn. The database connection returned from this will then be passed to Init
// for initializing the database.
func Open(typ, dsn string) (*DB, error) {
	dbMu.RLock()
	defer dbMu.RUnlock()

	db, ok := dbs[typ]

	if !ok {
		return nil, errors.New("unknown database type " + typ)
	}

	sqldb, err := sql.Open(db.Type, dsn)

	if err != nil {
		return nil, err
	}

	if err := db.Init(sqldb); err != nil {
		return nil, err
	}

	db.DB = sqldb
	return db, nil
}
