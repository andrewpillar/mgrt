package database

import (
	"crypto/sha256"
	"database/sql"
	"os"
	"testing"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/revision"
)

var revisionId = "1136214245"

func performRevisions(db *DB, t *testing.T) {
	if err := db.Init(); err != nil {
		t.Errorf("failed to initialize database: %s\n", err)
	}

	r, err := revision.Find(revisionId)

	if err != nil {
		t.Errorf("failed to find revision: %s\n", err)
		return
	}

	if err := r.GenHash(); err != nil {
		t.Errorf("failed to generate revision hash: %s\n", err)
	}

	r.Direction = revision.Up

	if err := db.Perform(r, false); err != nil {
		t.Errorf("failed to perform revision: %s\n", err)
		return
	}

	if err := db.Log(r, false); err != nil {
		t.Errorf("failed to log revision: %s\n", err)
		return
	}

	var count int64

	row := db.QueryRow("SELECT COUNT(*) FROM mgrt_revisions")
	row.Scan(&count)

	if count != int64(1) {
		t.Errorf("performed revisions did not match expected: expected = '1', actual = '%d'\n", count)
		return
	}

	// Check to see if the revision performed, and the example table exists.
	_, err = db.Query("INSERT INTO example (id) VALUES (1)")

	if err != nil {
		t.Errorf("failed to insert test record: %s\n", err)
		return
	}

	r.Direction = revision.Down

	if err := db.Perform(r, false); err != nil {
		t.Errorf("failed to perform revision: %s\n", err)
		return
	}

	if err := db.Log(r, false); err != nil {
		t.Errorf("failed to log revision: %s\n", err)
		return
	}

	r.Direction = revision.Up
	r.Hash = [sha256.Size]byte{}

	if err := db.Perform(r, false); err != ErrCheckHashFailed {
		t.Errorf("performed revision did not fail hash check: %s\n", err)
		return
	}

	if err := db.Perform(r, true); err != nil {
		t.Errorf("failed to perform revision: %s\n", err)
	}

	if err := db.Log(r, false); err != nil {
		t.Errorf("failed to log revision: %s\n", err)
		return
	}

	_, err = db.ReadLog(revisionId)

	if err != nil {
		t.Errorf("failed to read revisions log: %s\n", err)
		return
	}
}

func TestPerformMySQL(t *testing.T) {
	mysqlsrc := os.Getenv("MYSQLSOURCE")

	if mysqlsrc == "" {
		t.Skip("skipping mysql tests: MYSQLSOURCE not set\n")
	}

	sqlDB, err := sql.Open("mysql", mysqlsrc)

	if err != nil {
		t.Errorf("failed to open mysql database: %s\n", err)
	}

	db := &DB{
		DB:   sqlDB,
		Type: MySQL,
	}

	defer db.Close()

	performRevisions(db, t)
}

func TestPerformPostgresql(t *testing.T) {
	pgsrc := os.Getenv("PGSOURCE")

	if pgsrc == "" {
		t.Skip("skipping postgresql tests: PGSOURCE not set\n")
	}

	sqlDB, err := sql.Open("postgres", pgsrc)

	if err != nil {
		t.Errorf("failed to open postgresql database: %s\n", err)
	}

	db := &DB{
		DB:   sqlDB,
		Type: Postgres,
	}

	defer db.Close()

	performRevisions(db, t)
}

func TestPerformSqlite3(t *testing.T) {
	cfg := &config.Config{
		Type:    "sqlite3",
		Address: "test.sqlite3",
	}

	db, err := Open(cfg)

	if err != nil {
		t.Errorf("failed to open sqlite3 database: %s\n", err)
	}

	defer os.Remove("test.sqlite3")
	defer db.Close()

	performRevisions(db, t)
}

func TestMain(m *testing.M) {
	config.Root = "testdata"

	exitCode := m.Run()

	os.Exit(exitCode)
}
