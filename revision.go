// package mgrt provides a collection of functions for performing revisions
// against any given database connection.
package mgrt

import (
	"bufio"
	"bytes"
	"database/sql"
	"errors"
	"io"
	"os"
	"strings"
	"time"
)

// node is a node in the binary tree of a Collection. This stores the val used
// for sorting revisions in a Collection. The val will be the Unix time of the
// Revision ID, since Revision IDs are a time in the layout of 20060102150405.
type node struct {
	val   int64
	rev   *Revision
	left  *node
	right *node
}

// Errors is a collection of errors that occurred.
type Errors []error

// Revision is the type that represents what SQL code has been executed against
// a database as a revision. Typically, this would be changes made to the
// database schema itself.
type Revision struct {
	ID          string    // ID is the unique ID of the Revision.
	Category    string    // Category of the revision.
	Author      string    // Author is who authored the original Revision.
	Comment     string    // Comment provides a short description for the Revision.
	SQL         string    // SQL is the code that will be executed when the Revision is performed.
	PerformedAt time.Time // PerformedAt is when the Revision was executed.
}

// RevisionError represents an error that occurred with a revision.
type RevisionError struct {
	ID  string // ID is the ID of the revisions that errored.
	Err error  // Err is the underlying error itself.
}

// Collection stores revisions in a binary tree. This ensures that when they are
// retrieved, they will be retrieved in ascending order from when they were
// initially added.
type Collection struct {
	len  int
	root *node
}

var (
	revisionIdFormat = "20060102150405"

	// ErrInvalid is returned whenever an invalid Revision ID is encountered. A
	// Revision ID is considered invalid when the time layout 20060102150405
	// cannot be used for parse the ID.
	ErrInvalid = errors.New("revision id invalid")

	// ErrPerformed is returned whenever a Revision has already been performed.
	// This can be treated as a benign error.
	ErrPerformed = errors.New("revision already performed")

	ErrNotFound = errors.New("revision not found")
)

func insertNode(n **node, val int64, r *Revision) {
	if (*n) == nil {
		(*n) = &node{
			val: val,
			rev: r,
		}
		return
	}

	if val < (*n).val {
		insertNode(&(*n).left, val, r)
		return
	}
	insertNode(&(*n).right, val, r)
}

// NewRevision creates a new Revision with the given author, and comment.
func NewRevision(author, comment string) *Revision {
	return &Revision{
		ID:      time.Now().Format(revisionIdFormat),
		Author:  author,
		Comment: comment,
	}
}

// NewRevisionCategory creates a new Revision in the given category with the
// given author and comment.
func NewRevisionCategory(category, author, comment string) *Revision {
	rev := NewRevision(author, comment)
	rev.Category = category
	return rev
}

// RevisionPerformed checks to see if the given Revision has been performed
// against the given database.
func RevisionPerformed(db *DB, rev *Revision) error {
	var count int64

	if _, err := time.Parse(revisionIdFormat, rev.ID); err != nil {
		return ErrInvalid
	}

	q := db.Parameterize("SELECT COUNT(id) FROM mgrt_revisions WHERE (id = ?)")

	if err := db.QueryRow(q, rev.Slug()).Scan(&count); err != nil {
		return &RevisionError{
			ID:  rev.Slug(),
			Err: err,
		}
	}

	if count > 0 {
		return &RevisionError{
			ID:  rev.Slug(),
			Err: ErrPerformed,
		}
	}
	return nil
}

// GetRevision get's the Revision with the given ID.
func GetRevision(db *DB, id string) (*Revision, error) {
	var (
		rev Revision
		sec int64
	)

	q := "SELECT id, author, comment, sql, performed_at FROM mgrt_revisions WHERE (id = ?)"

	row := db.QueryRow(db.Parameterize(q), id)

	var categoryid string

	if err := row.Scan(&categoryid, &rev.Author, &rev.Comment, &rev.SQL, &sec); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &RevisionError{
				ID:  categoryid,
				Err: ErrNotFound,
			}
		}
		return nil, err
	}

	parts := strings.Split(categoryid, "/")

	end := len(parts) - 1

	rev.ID = parts[end]
	rev.Category = strings.Join(parts[:end], "/")

	rev.PerformedAt = time.Unix(sec, 0)
	return &rev, nil
}

// GetRevisions returns a list of all the revisions that have been performed
// against the given database. If n is <= 0 then all of the revisions will be
// retrieved, otherwise, only the given amount will be retrieved. The returned
// revisions will be ordered by their performance date descending.
func GetRevisions(db *DB, n int) ([]*Revision, error) {
	count := int64(n)

	if n <= 0 {
		q0 := "SELECT COUNT(id) FROM mgrt_revisions"

		if err := db.QueryRow(q0).Scan(&count); err != nil {
			return nil, err
		}
	}

	revs := make([]*Revision, 0, int(count))

	q := "SELECT id, author, comment, sql, performed_at FROM mgrt_revisions ORDER BY id DESC LIMIT ?"

	rows, err := db.Query(db.Parameterize(q), count)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			rev        Revision
			sec        int64
			categoryid string
		)

		err = rows.Scan(&categoryid, &rev.Author, &rev.Comment, &rev.SQL, &sec)

		if err != nil {
			return nil, err
		}

		parts := strings.Split(categoryid, "/")

		end := len(parts) - 1

		rev.ID = parts[end]
		rev.Category = strings.Join(parts[:end], "/")

		rev.PerformedAt = time.Unix(sec, 0)
		revs = append(revs, &rev)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return revs, nil
}

// PerformRevisions will perform the given revisions against the given database.
// The given revisions will be sorted into ascending order first before they
// are performed. If any of the given revisions have already been performed then
// the Errors type will be returned containing *RevisionError for each revision
// that was already performed.
func PerformRevisions(db *DB, revs0 ...*Revision) error {
	var c Collection

	for _, rev := range revs0 {
		c.Put(rev)
	}

	errs := Errors(make([]error, 0, len(revs0)))
	revs := c.Slice()

	for _, rev := range revs {
		if err := rev.Perform(db); err != nil {
			if errors.Is(err, ErrPerformed) {
				errs = append(errs, err)
				continue
			}
			return err
		}
	}
	return errs.err()
}

// OpenRevision opens the revision at the given path.
func OpenRevision(path string) (*Revision, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	return UnmarshalRevision(f)
}

// UnmarshalRevision will unmarshal a Revision from the given io.Reader. This
// will expect to see a comment block header that contains the metadata about
// the Revision itself. This will check to see if the given Revision ID is
// valid. A Revision id is considered valid when it can be parsed into a
// valid time via time.Parse using the layout of 20060102150405.
func UnmarshalRevision(r io.Reader) (*Revision, error) {
	br := bufio.NewReader(r)

	rev := &Revision{}

	var (
		buf     []rune = make([]rune, 0)
		r0      rune
		inBlock bool
	)

	for {
		r, _, err := br.ReadRune()

		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			rev.SQL = strings.TrimSpace(string(buf))
			break
		}

		if r == '*' {
			if r0 == '/' {
				inBlock = true
				continue
			}
		}

		if r == '/' {
			if r0 == '*' {
				rev.Comment = strings.TrimSpace(string(buf))
				buf = buf[0:0]
				inBlock = false
				continue
			}
		}

		if inBlock {
			if r == '\n' {
				if r0 == '\n' {
					if rev.ID == "" && rev.Author == "" {
						buf = buf[0:0]
					}
					goto cont
				}

				pos := -1

				for i, r := range buf {
					if r == ':' {
						pos = i
						break
					}
				}

				if pos < 0 {
					goto cont
				}

				if string(buf[pos-6:pos]) == "Author" {
					rev.Author = strings.TrimSpace(string(buf[pos+1:]))
					buf = buf[0:0]
					continue
				}

				if string(buf[pos-8:pos]) == "Revision" {
					rev.ID = strings.TrimSpace(string(buf[pos+1:]))
					buf = buf[0:0]
					continue
				}
			}
		}

		if r == '*' {
			peek, _, err := br.ReadRune()

			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				continue
			}

			if peek == '/' {
				br.UnreadRune()
				r0 = r
				continue
			}
		}

	cont:
		buf = append(buf, r)
		r0 = r
	}

	parts := strings.Split(rev.ID, "/")
	end := len(parts)-1

	rev.ID = parts[len(parts)-1]
	rev.Category = strings.Join(parts[:end], "/")

	if _, err := time.Parse(revisionIdFormat, rev.ID); err != nil {
		return nil, ErrInvalid
	}
	return rev, nil
}

func (n *node) walk(visit func(*Revision)) {
	if n.left != nil {
		n.left.walk(visit)
	}

	visit(n.rev)

	if n.right != nil {
		n.right.walk(visit)
	}
}

func (e Errors) err() error {
	if len(e) == 0 {
		return nil
	}
	return e
}

// Error returns the string representation of all the errors in the underlying
// slice. Each error will be on a separate line in the returned string.
func (e Errors) Error() string {
	var buf bytes.Buffer

	for _, err := range e {
		buf.WriteString(err.Error() + "\n")
	}
	return buf.String()
}

// Put puts the given Revision in the current Collection.
func (c *Collection) Put(r *Revision) error {
	if r.ID == "" {
		return ErrInvalid
	}

	t, err := time.Parse(revisionIdFormat, r.ID)

	if err != nil {
		return ErrInvalid
	}

	insertNode(&c.root, t.Unix(), r)
	c.len++
	return nil
}

// Len returns the number of items in the collection.
func (c *Collection) Len() int { return c.len }

// Slice returns a sorted slice of all the revisions in the collection.
func (c *Collection) Slice() []*Revision {
	if c.len == 0 {
		return nil
	}

	revs := make([]*Revision, 0, c.len)

	c.root.walk(func(r *Revision) {
		revs = append(revs, r)
	})
	return revs
}

func (e *RevisionError) Error() string {
	return "revision error " + e.ID + ": " + e.Err.Error()
}

// Unwrap returns the underlying error that caused the original RevisionError.
func (e *RevisionError) Unwrap() error { return e.Err }

// Slug returns the slug of the revision ID, this will be in the format of
// category/id if the revision belongs to a category.
func (r *Revision) Slug() string {
	if r.Category != "" {
		return r.Category + "/" + r.ID
	}
	return r.ID
}

// Perform will perform the current Revision against the given database. If
// the Revision is emtpy, then nothing happens. If the Revision has already
// been performed, then ErrPerformed is returned.
func (r *Revision) Perform(db *DB) error {
	if r.SQL == "" {
		return nil
	}

	if err := RevisionPerformed(db, r); err != nil {
		return err
	}

	if _, err := db.Exec(r.SQL); err != nil {
		return &RevisionError{
			ID: r.Slug(),
			Err: err,
		}
	}

	q := db.Parameterize("INSERT INTO mgrt_revisions (id, author, comment, sql, performed_at) VALUES (?, ?, ?, ?, ?)")

	if _, err := db.Exec(q, r.Slug(), r.Author, r.Comment, r.SQL, time.Now().Unix()); err != nil {
		return &RevisionError{
			ID: r.Slug(),
			Err: err,
		}
	}
	return nil
}

// Title will extract the title from the comment of the current Revision. First,
// this will truncate the title to being 72 characters. If the comment was longer
// than 72 characters, then the title will be suffixed with "...". If a LF
// character can be found in the title, then the title will be truncated again
// up to where that LF character occurs.
func (r *Revision) Title() string {
	title := r.Comment

	if l := len(title); l >= 72 {
		title = title[:72]

		if l > 72 {
			title += "..."
		}
	}

	if i := bytes.IndexByte([]byte(title), '\n'); i > 0 {
		title = title[:i]
	}
	return title
}

// String returns the string representation of the Revision. This will be the
// comment block header followed by the Revision SQL itself.
func (r *Revision) String() string {
	var buf bytes.Buffer

	buf.WriteString("/*\n")
	buf.WriteString("Revision: " + r.Slug() + "\n")
	buf.WriteString("Author:   " + r.Author + "\n")

	if r.Comment != "" {
		buf.WriteString("\n" + r.Comment + "\n")
	}
	buf.WriteString("*/\n\n")
	buf.WriteString(r.SQL)
	return buf.String()
}
