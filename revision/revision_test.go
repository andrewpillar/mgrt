package revision

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/andrewpillar/mgrt/config"
)

func TestWalk(t *testing.T) {
	revisions, err := walk(append_)

	if err != nil {
		t.Errorf("failed to walk revisions: %s\n", err)
		return
	}

	expected := int64(1136214245)

	if revisions[0].ID != expected {
		t.Errorf(
			"revision id does not match: expected = '%d', actual = '%d'\n",
			expected,
			revisions[0].ID,
		)
		return
	}

	revisions, err = walk(prepend_)

	expected = int64(1136214247)

	if revisions[0].ID != expected {
		t.Errorf(
			"revision id does not match: expected = '%d', actual = '%d'\n",
			expected,
			revisions[0].ID,
		)
		return
	}
}

func TestAdd(t *testing.T) {
	r, err := Add("")

	if err != nil {
		t.Errorf("failed to add revision: %s\n", err)
		return
	}

	path := filepath.Join(config.RevisionsDir(), strconv.FormatInt(r.ID, 10))

	info, err := os.Stat(path)

	if err != nil {
		t.Errorf("failed to stat path: %s\n", err)
		return
	}

	if !info.IsDir() {
		t.Errorf("revision is not a directory\n")
		return
	}

	for _, p := range []string{r.MessagePath, r.DownPath, r.UpPath} {
		info, err = os.Stat(p)

		if err != nil {
			t.Errorf("failed to stat path: %s\n", err)
			return
		}
	}

	if err := os.RemoveAll(path); err != nil {
		t.Errorf("failed to clear test files: %s\n", err)
	}
}

func TestFind(t *testing.T) {
	r, err := Find("1136214245")

	if err != nil {
		t.Errorf("failed to find revision: %s\n", err)
		return
	}

	message := "Some message"

	if r.Message != message {
		t.Errorf(
			"revision message does not match: expected = '%s', actual = '%s'\n",
			message,
			r.Message,
		)
		return
	}

	up := "CREATE TABLE example();\n"
	down := "DROP TABLE example;\n"

	if r.Up.String != up {
		t.Errorf("revision up does not match: expected = '%s', actual = '%s'\n", up, r.Up.String)
		return
	}

	if r.Down.String != down {
		t.Errorf("revision down does not match: expected = '%s', actual = '%s'\n", down, r.Down.String)
		return
	}
}

func TestMain(m *testing.M) {
	config.Root = "testdata"

	exitCode := m.Run()

	os.Exit(exitCode)
}
