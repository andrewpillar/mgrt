package mgrt

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

type Direction uint

//go:generate stringer -type Direction -linecomment
const (
	Up   Direction = iota + 1 // up
	Down                      // down
)

// Revision represents some SQL code that has been performed against a database.
// A revision can be performed in either an Up or Down direction. Each revision
// will have an ID which will be the unix nano timestamp for when it was
// performed, and a ref which will be a sha26 sum of the file contents from
// which the revision came from.
type Revision struct {
	ID          int64
	Ref         string
	Name        string
	Comment     sql.Null[string]
	Up          sql.Null[string]
	Down        sql.Null[string]
	Direction   Direction
	PerformedAt time.Time
}

// SQL returns the SQL code for the revision. The code returned will depend on
// the direction the revision was performed in. If the revision direction is
// invalid then this returns an empty string.
func (r *Revision) SQL() string {
	switch r.Direction {
	case Up:
		return r.Up.V
	case Down:
		return r.Down.V
	}
	return ""
}

type buffer struct {
	buf []byte
	pos int
	eof int
	w   int
	lit []rune
}

func newBuffer(r io.Reader) (*buffer, error) {
	b, err := io.ReadAll(r)

	if err != nil {
		return nil, err
	}

	return &buffer{
		buf: b,
		pos: 0,
		eof: len(b),
	}, nil
}

func (b *buffer) get() rune {
	if b.pos >= b.eof {
		return -1
	}

	r := rune(b.buf[b.pos])
	w := 1

	if r >= utf8.RuneSelf {
		r, w = utf8.DecodeRune(b.buf[b.pos:])
	}

	b.pos += w
	b.w = w

	return r
}

func (b *buffer) unget() {
	b.pos -= b.w
	b.w = 0
}

func (b *buffer) getLine() (string, bool) {
	b.lit = b.lit[0:0]

	r := b.get()

	for r != '\n' && r != -1 {
		b.lit = append(b.lit, r)
		r = b.get()
	}

	if r == -1 {
		if len(b.lit) > 0 {
			goto ret
		}
		return "", false
	}

ret:
	return string(b.lit), true
}

// Parse a revision from the given [io.Reader] and set the name of it.
func Parse(name string, r io.Reader) (*Revision, error) {
	sha256 := sha256.New()

	buf, err := newBuffer(io.TeeReader(r, sha256))

	if err != nil {
		return nil, err
	}

	rev := Revision{
		Ref:  hex.EncodeToString(sha256.Sum(nil)),
		Name: name,
	}

	var tmp bytes.Buffer

	for {
		line, ok := buf.getLine()

		if !ok {
			break
		}

		if strings.HasPrefix(line, "/*") {
			for {
				line, ok = buf.getLine()
				line = strings.TrimSpace(line)

				if line == "*/" || !ok {
					break
				}

				if len(line) > 0 {
					if line[0] == '*' {
						line = strings.TrimSpace(line[1:])
					}
				}

				tmp.WriteString(strings.TrimSpace(line))
				tmp.WriteString("\n")
			}

			rev.Comment.V = tmp.String()
			rev.Comment.Valid = true

			tmp.Reset()
		}

		if strings.HasPrefix(line, "-- +") {
			direction := line[4:]

			if direction == "up" || direction == "down" {
				for {
					line, ok = buf.getLine()

					if !ok {
						if direction == "up" {
							rev.Up.V = tmp.String()
							rev.Up.Valid = true

							tmp.Reset()
						}

						if direction == "down" {
							rev.Down.V = tmp.String()
							rev.Down.Valid = true

							tmp.Reset()
						}
						break
					}

					if strings.HasPrefix(line, "-- +") {
						line = line[4:]

						if direction == "up" && line == "down" {
							rev.Up.V = tmp.String()
							rev.Up.Valid = true

							tmp.Reset()
							direction = line

							continue
						}
					}

					tmp.WriteString(line)
					tmp.WriteString("\n")
				}
			}
		}
	}
	return &rev, nil
}

// Load returns the revision in the named file from the given [io.fs.FS].
func Load(fsys fs.FS, name string) (*Revision, error) {
	f, err := fsys.Open(name)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	return Parse(name, f)
}

// LoadDir returns all of the revisions in the named directory from the given
// [io.fs.FS].
func LoadDir(fsys fs.FS, name string) ([]*Revision, error) {
	ents, err := fs.ReadDir(fsys, filepath.Clean(name))

	if err != nil {
		return nil, err
	}

	revs := make([]*Revision, 0, len(ents))

	for _, ent := range ents {
		if ent.IsDir() {
			continue
		}

		if !strings.HasSuffix(ent.Name(), ".sql") {
			continue
		}

		rev, err := Load(fsys, filepath.Join(name, ent.Name()))

		if err != nil {
			return nil, err
		}
		revs = append(revs, rev)
	}
	return revs, nil
}

var ErrDirectionInvalid = errors.New("revision direction invalid")

const (
	revisionTable  = "_mgrt_revisions"
	revisionSchema = `CREATE TABLE IF NOT EXISTS _mgrt_revisions (
	id           INTEGER PRIMARY KEY,
	ref          VARCHAR,
	name         VARCHAR NOT NULL,
	comment      TEXT NULL,
	up           TEXT NULL,
	down         TEXT NULL,
	direction    INTEGER NOT NULL,
	performed_at TIMESTAMP NOT NULL
);`
)

func lastRevision(ctx context.Context, db *sql.DB, ref string) (*Revision, bool, error) {
	q := "SELECT * FROM _mgrt_revisions WHERE (ref = $1) ORDER BY performed_at DESC"

	row := db.QueryRowContext(ctx, q, ref)

	var rev Revision

	if err := row.Scan(&rev.ID, &rev.Ref, &rev.Name, &rev.Comment, &rev.Up, &rev.Down, &rev.Direction, &rev.PerformedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	if err := row.Err(); err != nil {
		return nil, false, err
	}
	return &rev, rev.Ref != "", nil
}

const insertRevision = `
INSERT INTO _mgrt_revisions (id, ref, name, comment, up, down, direction, performed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

// Perform the given revisions against the given database in the given
// [Direction]. This will create the necessary table in the database to store
// the revisions that have been performed. Each revision that is performed will
// be sent to the given channel, if not nil, it is valid to pass a nil channel.
// If no revisions are given then this is a no-op. If a given revision has not
// SQL code associated with the direction, then the revision is not performed
// and not logged in the database.
func Perform(ctx context.Context, db *sql.DB, done chan<- *Revision, d Direction, revs ...*Revision) error {
	if len(revs) == 0 {
		return nil
	}

	if d < Up && d > Down {
		return ErrDirectionInvalid
	}

	if _, err := db.ExecContext(ctx, revisionSchema); err != nil {
		return err
	}

	order := make([]*Revision, 0, len(revs))

	for _, rev := range revs {
		order = append(order, rev)
	}

	if d == Down {
		slices.Reverse(order)
	}

	for _, rev := range order {
		rev.Direction = d

		last, ok, err := lastRevision(ctx, db, rev.Ref)

		if err != nil {
			return err
		}

		if ok {
			if last.Direction == rev.Direction {
				continue
			}
		}

		sql := rev.SQL()

		if sql == "" {
			continue
		}

		if _, err := db.ExecContext(ctx, sql); err != nil {
			return err
		}

		rev.ID = time.Now().UnixNano()
		rev.PerformedAt = time.Now()

		args := []any{
			rev.ID, rev.Ref, rev.Name, rev.Comment, rev.Up, rev.Down, rev.Direction, rev.PerformedAt,
		}

		if _, err := db.ExecContext(ctx, insertRevision, args...); err != nil {
			return err
		}

		if done != nil {
			done <- rev
		}
	}
	return nil
}

const selectRevisions = "SELECT * FROM _mgrt_revisions ORDER BY performed_at DESC"

// Log returns all of the revisions that have been performed against the given
// database in the order of most recent.
func Log(ctx context.Context, db *sql.DB) ([]*Revision, error) {
	rows, err := db.QueryContext(ctx, selectRevisions)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	revs := make([]*Revision, 0)

	for rows.Next() {
		rev := &Revision{}

		dest := []any{
			&rev.ID, &rev.Ref, &rev.Name, &rev.Comment, &rev.Up, &rev.Down, &rev.Direction, &rev.PerformedAt,
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		revs = append(revs, rev)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return revs, nil
}
