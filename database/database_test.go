package database

import (
	"crypto/sha256"
	"os"
	"testing"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/revision"
)

var revisionIds = []string{
	"1136214245",
	"1136214246",
	"1136214247",
	"1136214248",
}

func performRevisions(db *DB, t *testing.T) {
	if err := db.Init(); err != nil {
		t.Errorf("failed to initialize database: %s\n", err)
	}

	performed := make([]*revision.Revision, len(revisionIds) - 1, len(revisionIds) - 1)

	for i, id := range revisionIds {
		r, err := revision.Find(id)

		if err != nil {
			t.Errorf("failed to find revision: %s\n", err)
			break
		}

		if i > 0 {
			performed[i - 1] = r
		}

		r.Direction = revision.Up

		if err := db.Perform(r, false); err != nil {
			t.Errorf("failed to perform revision %d: %s\n", r.ID, err)
			break
		}

		if err := db.Log(r, false); err != nil {
			t.Errorf("failed to log revision %d: %s\n", r.ID, err)
			break
		}
	}

	for _, r := range performed {
		r.Direction = revision.Down
		r.Hash = [sha256.Size]byte{}

		if err := db.Perform(r, false); err != ErrChecksumFailed {
			t.Errorf("expected revision to fail checksum, it did not %d: %s\n", r.ID, err)
			break
		}

		if err := db.Perform(r, true); err != nil {
			t.Errorf("expected revision to not fail, it did %d: %s\n", r.ID, err)
			break
		}
	}
}

func TestPerformMySQL(t *testing.T) {
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPswd := os.Getenv("MYSQL_PSWD")
	mysqlDb := os.Getenv("MYSQL_DB")

	if mysqlUser == "" || mysqlPswd == "" || mysqlDb == "" {
		t.Log("missing one of: MYSQL_USER, MYSQL_PSWD, MYSQL_DB")
		t.Log("not running MySQL tests")
		return
	}

	cfg := &config.Config{
		Type:     "mysql",
		Username: mysqlUser,
		Password: mysqlPswd,
		Database: mysqlDb,
	}

	db, err := Open(cfg)

	if err != nil {
		t.Errorf("failed to open mysql database: %s\n", err)
	}

	defer db.Close()

	performRevisions(db, t)
}

func TestPerformPostgresql(t *testing.T) {
	pgAddr := os.Getenv("PG_ADDR")
	pgUser := os.Getenv("PG_USER")
	pgPswd := os.Getenv("PG_PSWD")
	pgDb := os.Getenv("PG_DB")

	if pgAddr == "" || pgUser == "" || pgPswd == "" || pgDb == "" {
		t.Log("missing one of: PG_ADDR, PG_USER, PG_PSWD, PG_DB")
		t.Log("not running Postgresql tests")
		return
	}

	cfg := &config.Config{
		Type:     "postgres",
		Address:  pgAddr,
		Username: pgUser,
		Password: pgPswd,
		Database: pgDb,
	}

	db, err := Open(cfg)

	if err != nil {
		t.Errorf("failed to open postgresql database: %s\n", err)
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
