package mgrt

import (
	"strings"
	"testing"
)

func Test_UnmarshalRevision(t *testing.T) {
	r := strings.NewReader(`/*
Revision: 20060102150405
Author:   Author <me@example.com>

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

	if rev.Comment != "Comment line 1\nComment line 2" {
		t.Errorf("unexpected revision comment, expected=%q, got=%q\n", "Comment line 1\nComment line 2", rev.Comment)
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

	tests := []struct{
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
