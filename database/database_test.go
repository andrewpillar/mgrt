package database

import (
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

func TestPerform(t *testing.T) {
	config.Root = "testdata"

	cfg := &config.Config{
		Type:    "sqlite3",
		Address: "test.sqlite3",
	}

	db, err := Open(cfg)

	if err != nil {
		t.Errorf("failed to open sqlite3 database: %s\n", err)
	}

	defer db.Close()

	if err := db.Init(); err != nil {
		t.Errorf("failed to initialize sqlite3 database: %s\n", err)
	}

	for _, id := range revisionIds {
		r, err := revision.Find(id)

		if err != nil {
			t.Errorf("failed to find revision: %s\n", err)
			break
		}

		r.Direction = revision.Up

		if err := db.Perform(r, false); err != nil {
			t.Errorf("failed to perform revision %d: %s\n", r.ID, err)
			break
		}
	}

	os.Remove("test.sqlite3")
}
