package mgrt

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	_ "modernc.org/sqlite"
)

//go:embed *.sql
var revisions embed.FS

func TestLoad(t *testing.T) {
	tests := []struct {
		name string
		want Revision
	}{
		{
			"rev0.sql",
			Revision{
				Ref:  "36e2d57d260a34c879622877fa57e80024210b13fe41f6a3766bd2f5b0fb7c98",
				Name: "rev0.sql",
				Up: sql.Null[string]{
					V:     "CREATE TABLE IF NOT EXISTS t (\n\tcol VARCHAR NOT NULL\n);\n",
					Valid: true,
				},
				Down: sql.Null[string]{
					V:     "DROP TABLE IF EXISTS t;\n",
					Valid: true,
				},
			},
		},
		{
			"rev1.sql",
			Revision{
				Ref:  "ea2884c422e67d1ab51dbfede75e40d2a7c1d53a527c4b24e9f9175408dae4d3",
				Name: "rev1.sql",
				Up: sql.Null[string]{
					V:     "CREATE TABLE IF NOT EXISTS t (\n\tcol VARCHAR NOT NULL\n);\n",
					Valid: true,
				},
			},
		},
		{
			"rev2.sql",
			Revision{
				Ref:  "3089e2c405853d3cbfb1108b2b7c34858296966fa31c6f36b2ef827bd1155679",
				Name: "rev2.sql",
				Down: sql.Null[string]{
					V:     "DROP TABLE IF EXISTS t;\n",
					Valid: true,
				},
			},
		},
		{
			"rev3.sql",
			Revision{
				Ref:  "4eac6c6786fefd31bbd56e7c88cda24502a46fe777c4e339bb6f9cb0ab1fdd38",
				Name: "rev3.sql",
				Comment: sql.Null[string]{
					V:     "Empty revision.\n",
					Valid: true,
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test.%d", i), func(t *testing.T) {
			rev, err := Load(revisions, test.name)

			if err != nil {
				t.Fatalf("Load(revisions, %q): %v\n", test.name, err)
			}

			if diff := cmp.Diff(test.want, *rev); diff != "" {
				t.Fatalf("revision mismatch(-want, +got):\n%s", diff)
			}
		})
	}
}

func DoPerform(t *testing.T, db *sql.DB, d Direction, revs ...*Revision) {
	done := make(chan *Revision)

	go func() {
		defer close(done)

		if err := Perform(t.Context(), db, done, d, revs...); err != nil {
			t.Fatalf("Perform(ctx, db, done, %v, revs...): %v\n", d, err)
		}
	}()

	for rev := range done {
		t.Log(rev.Ref[:9], rev.Name, rev.Direction)
	}
}

func TestPerform(t *testing.T) {
	db, err := sql.Open("sqlite", t.Name())

	if err != nil {
		t.Fatalf("sql.Open(%q, %q): %v\n", "sqlite", t.Name(), err)
	}

	defer os.Remove(t.Name())
	defer db.Close()

	revs, err := LoadDir(revisions, ".")

	if err != nil {
		t.Fatalf("LoadDir(revisions, %q): %v\n", ".", err)
	}

	ctx := t.Context()

	DoPerform(t, db, Up, revs...)

	performed, err := Log(ctx, db)

	if err != nil {
		t.Fatalf("Log(ctx, db): %v\n", err)
	}

	if l := len(performed); l != 2 {
		t.Fatalf("len(performed) = %v, want = %v\n", l, 2)
	}

	if name := performed[0].Name; name != "rev1.sql" {
		t.Fatalf("performed[0].Name = %v, want = %v\n", name, "rev1.sql")
	}

	DoPerform(t, db, Down, revs...)

	performed, err = Log(ctx, db)

	if err != nil {
		t.Fatalf("Log(ctx, db): %v\n", err)
	}

	if l := len(performed); l != 4 {
		t.Fatalf("len(performed) = %v, want = %v\n", l, 4)
	}

	if name := performed[0].Name; name != "rev0.sql" {
		t.Fatalf("performed[0].Name = %v, want = %v\n", name, "rev0.sql")
	}
}
