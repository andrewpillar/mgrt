package mgrt

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func Test_UnmarshalRevision(t *testing.T) {
	r := strings.NewReader(`/*
Revision: 20060102150405
Author:   Author <me@example.com>

Title

Comment line 1
Comment line 2
*/
DROP TABLE users;`)

	rev, err := UnmarshalRevision(r)

	if err != nil {
		t.Fatal(err)
	}

	if rev.ID != "20060102150405" {
		t.Errorf("unexpected revision id, expected=%q, got=%q\n", "20060102150405", rev.ID)
	}

	if rev.Author != "Author <me@example.com>" {
		t.Errorf("unexpected revision author, expected=%q, got=%q\n", "Author <me@example.com>", rev.Author)
	}

	if rev.Comment != "Title\n\nComment line 1\nComment line 2" {
		t.Errorf("unexpected revision comment, expected=%q, got=%q\n", "Title\n\nComment line 1\nComment line 2", rev.Comment)
	}

	if title := rev.Title(); title != "Title" {
		t.Errorf("unexpected revision comment title, expected=%q, got=%q\n", "Title", title)
	}

	if rev.SQL != "DROP TABLE users;" {
		t.Errorf("unexpected revision sql, expected=%q, got=%q\n", "DROP TABLE users;", rev.SQL)
	}
}

func Test_RevisionTitle(t *testing.T) {
	singleLineComment := "A title that is longer than 72 characters in length this should be trimmed with an ellipsis."
	multiLineComment := `A comment that will have multiple lines and a long title line

This is the body of the comment.`
	shortComment := "A simple comment that is shorter thant 72 characters in length"

	tests := []struct {
		comment  string
		expected string
	}{
		{
			comment:  singleLineComment,
			expected: "A title that is longer than 72 characters in length this should be trimm...",
		},
		{
			comment:  multiLineComment,
			expected: "A comment that will have multiple lines and a long title line",
		},
		{
			comment:  shortComment,
			expected: shortComment,
		},
	}

	for i, test := range tests {
		rev := Revision{Comment: test.comment}

		if title := rev.Title(); title != test.expected {
			t.Errorf("tests[%d] - expected=%q, got=%q\n", i, test.expected, title)
		}
	}
}

func Test_RevisionPerformMultiple(t *testing.T) {
	tmp, err := ioutil.TempFile("", "mgrt-db-*")

	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmp.Name())

	db, err := Open("sqlite3", tmp.Name())

	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	tests := []struct {
		id      string
		author  string
		comment string
		sql     string
	}{
		{
			"20060102150407",
			"Andrew",
			"Add password to users table",
			"ALTER TABLE users ADD COLUMN password VARCHAR NOT NULL;",
		},
		{
			"20060102150405",
			"Andrew",
			"Add users table",
			"CREATE TABLE users ( id INT NOT NULL UNIQUE );",
		},
		{
			"20060102150406",
			"Andrew",
			"Add username to users table",
			"ALTER TABLE users ADD COLUMN username VARCHAR NOT NULL;",
		},
	}

	revs := make([]*Revision, 0, len(tests))

	for _, test := range tests {
		rev := NewRevision(test.author, test.comment)
		rev.ID = test.id
		rev.SQL = test.sql

		revs = append(revs, rev)
	}

	if err := PerformRevisions(db, revs...); err != nil {
		t.Fatal(err)
	}

	_, err = GetRevision(db, "foo")

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("unexpected error, expected=%T, got=%T\n", ErrNotFound, err)
	}

	if _, err = GetRevision(db, "20060102150406"); err != nil {
		t.Fatal(err)
	}
}

func Test_RevisionPerform(t *testing.T) {
	tmp, err := ioutil.TempFile("", "mgrt-db-*")

	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmp.Name())

	db, err := Open("sqlite3", tmp.Name())

	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	tests := []struct {
		id      string
		repeat  bool
		author  string
		comment string
		sql     string
	}{
		{
			"20060102150405",
			true,
			"Andrew",
			"Add users table",
			"CREATE TABLE users ( id INT NOT NULL UNIQUE );",
		},
		{
			"20060102150406",
			false,
			"Andrew",
			"Add username to users table",
			"ALTER TABLE users ADD COLUMN username VARCHAR NOT NULL;",
		},
		{
			"20060102150407",
			false,
			"Andrew",
			"Add password to users table",
			"ALTER TABLE users ADD COLUMN password VARCHAR NOT NULL;",
		},
	}

	for i, test := range tests {
		rev := NewRevision(test.author, test.comment)
		rev.ID = test.id
		rev.SQL = test.sql

		if err := rev.Perform(db); err != nil {
			t.Fatalf("tests[%d] - unexpected error %T %q\n", i, err, err)
		}

		if test.repeat {
			if err := rev.Perform(db); err != nil {
				if !errors.Is(err, ErrPerformed) {
					t.Fatalf("tests[%d] - unexpected error, expected=%T, got=%t\n", i, ErrPerformed, errors.Unwrap(err))
				}
			}
		}
	}

	revs, err := GetRevisions(db, -1)

	if err != nil {
		t.Fatal(err)
	}

	if len(revs) != len(tests) {
		t.Fatalf("unexpected revision count, expected=%d, got=%d\n", len(tests), len(revs))
	}
}

func Test_LoadRevisions(t *testing.T) {
	revs, err := LoadRevisions("revisions")

	if err != nil {
		t.Fatal(err)
	}

	if l := len(revs); l != 2 {
		t.Fatalf("unexpected revision count, expected=%d, got=%d\n", l, 2)
	}
}
