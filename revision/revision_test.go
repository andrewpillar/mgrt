package revision

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/andrewpillar/mgrt/config"
)

func TestFind(t *testing.T) {
	tests := []struct{
		id   string
		msg  string
		up   string
		down string
	}{
		{
			id:   "1136214245",
			msg:  "Revision one",
			up:   "CREATE TABLE example();\n",
			down: "DROP TABLE example;\n",
		},
		{
			id:  "1136214248",
			up:   "CREATE TABLE example();\n",
			down: "DROP TABLE example;\n",
		},
	}

	for _, tst := range tests {
		r, err := Find(tst.id)

		if err != nil {
			t.Fatal(err)
		}

		if tst.msg != r.Message {
			t.Errorf(
				"revision message does not match\n\texpeced = '%s'\n\t  actual = '%s'\n",
				tst.msg,
				r.Message,
			)
		}

		if tst.up != r.Up.String {
			t.Errorf(
				"revision up does not match\n\texpeced = '%s'\n\t  actual = '%s'\n",
				tst.up,
				r.Up.String,
			)
		}

		if tst.down != r.Down.String {
			t.Errorf(
				"revision up does not match\n\texpeced = '%s'\n\t  actual = '%s'\n",
				tst.down,
				r.Down.String,
			)
		}
	}
}

func TestAdd(t *testing.T) {
	tests := []struct{
		msg  string
		path string
	}{
		{
			msg:  "Test adding revision",
			path: "^testdata/revisions/[0-9]+_test_adding_revision$",
		},
		{
			msg:  "",
			path: "^testdata/revisions/[0-9]+$",
		},
	}

	for _, tst := range tests {
		r, err := Add(tst.msg)

		if err != nil {
			t.Fatal(err)
		}

		re := regexp.MustCompile(tst.path)

		if !re.Match([]byte(r.path)) {
			t.Errorf(
				"revision path does not match pattern\n\texpected = '%s'\n\t  actual = '%s'\n",
				tst.path,
				r.path,
			)
		}

		os.RemoveAll(r.path)
	}
}

func TestOldest(t *testing.T) {
	rr, err := Oldest()

	if err != nil {
		t.Fatal(err)
	}

	expected := int64(1136214245)

	if rr[0].ID != expected {
		t.Errorf(
			"revision id does not match\n\texpected = '%d'\n\t  actual = '%d'\n",
			expected,
			rr[0].ID,
		)
	}
}

func TestLatest(t *testing.T) {
	rr, err := Latest()

	if err != nil {
		t.Fatal(err)
	}

	expected := int64(1136214248)

	if rr[0].ID != expected {
		t.Errorf(
			"revision id does not match\n\texpected = '%d'\n\t  actual = '%d'\n",
			expected,
			rr[0].ID,
		)
	}
}

func TestHash(t *testing.T) {
	expected := "7b038fe74e91177f8daf2f05e06f95fe91b6904c8a4f651629999f783da4cdbe"

	r, err := Find("1136214245")

	if err != nil {
		t.Fatal(err)
	}

	if err := r.GenHash(); err != nil {
		t.Fatal(err)
	}

	hash := fmt.Sprintf("%x", r.Hash)

	if hash != expected {
		t.Errorf(
			"revision hash does not match\n\texpected = '%s'\n\t  actual = '%s'\n",
			expected,
			hash,
		)
	}
}

func TestMain(m *testing.M) {
	config.Root = "testdata"

	exitCode := m.Run()

	os.Exit(exitCode)
}
