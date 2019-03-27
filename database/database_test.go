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

	defer os.Remove("test.sqlite3")
	defer db.Close()

	if err := db.Init(); err != nil {
		t.Errorf("failed to initialize sqlite3 database: %s\n", err)
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
