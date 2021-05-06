package mgrt

import (
	"database/sql"
	"errors"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// InitFunc is a function that is called to initialize a database with the
// necessary table for performing revisions.
type InitFunc func(*sql.DB) error

var (
	initMu    sync.RWMutex
	initfuncs = make(map[string]InitFunc)

	mysqlInit = `CREATE TABLE mgrt_revisions (
	id           VARCHAR NOT NULL UNIQUE,
	author       VARCHAR NOT NULL,
	comment      TEXT NOT NULL,
	query        TEXT NOT NULL,
	performed_at INT NOT NULL
);`

	postgresInit = `CREATE TABLE mgrt_revisions (
	id           VARCHAR NOT NULL UNIQUE,
	author       VARCHAR NOT NULL,
	comment      TEXT NOT NULL,
	query        TEXT NOT NULL,
	performed_at INT NOT NULL
);`
)

func init() {
	Register("mysql", doMysqlInit)
	Register("postgresql", doPostgresqlInit)
}

func doMysqlInit(db *sql.DB) error {
	if _, err := db.Exec(mysqlInit); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}

func doPostgresqlInit(db *sql.DB) error {
	if _, err := db.Exec(postgresInit); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}

// Register will register the given InitFunc for the given database type. If the
// given type if a duplicate, then this function panics. If the given function
// is nil, then this function panics.
func Register(typ string, fn InitFunc) {
	initMu.Lock()
	defer initMu.Unlock()

	if fn == nil {
		panic("mgrt: nil database init function")
	}

	if _, ok := initfuncs[typ]; ok {
		panic("mgrt: init function already registered for " + typ)
	}
	initfuncs[typ] = fn
}

// Init will initialize the given database with the necessary table for
// performing revisions.
func Init(typ string, db *sql.DB) error {
	initMu.RLock()
	defer initMu.RUnlock()

	init, ok := initfuncs[typ]

	if !ok {
		return errors.New("unknown database type" + typ)
	}
	return init(db)
}

// Open is a utility function that will call sql.Open with the given typ and
// dsn. The database connection returned from this will then be passed to Init
// for initializing the database.
func Open(typ, dsn string) (*sql.DB, error) {
	db, err := sql.Open(typ, dsn)

	if err != nil {
		return nil, err
	}

	if err := Init(typ, db); err != nil {
		return nil, err
	}
	return db, nil
}
