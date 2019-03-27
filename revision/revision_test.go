package revision

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/andrewpillar/mgrt/config"
)

func TestAdd(t *testing.T) {
	expectedId := time.Now().Unix()
	expectedAuthor := "test <test@example.com>"
	expectedPath := fmt.Sprintf("%s/%d.sql", config.RevisionsDir(), expectedId)
	expectedStub := fmt.Sprintf(stub, expectedId, "", expectedAuthor)

	r, err := Add("", "test", "test@example.com")

	if err != nil {
		t.Errorf("failed to add revision: %s\n", err)
	}

	if r.ID != expectedId {
		t.Errorf("actual revision id does not match expected: expected = '%d', actual = '%d'\n", expectedId, r.ID)
	}

	if r.Author != expectedAuthor {
		t.Errorf("actual revision author does not match expected: expected = '%s', actual = '%s'\n", expectedAuthor, r.Author)
	}

	if r.Path != expectedPath {
		t.Errorf("actual revision path does not match expected: expected = '%s', actual = '%s'\n", expectedPath, r.Path)
	}

	b, err := ioutil.ReadFile(r.Path)

	if err != nil {
		t.Errorf("failed to read file: %s\n", err)
	}

	bexpected := []byte(expectedStub)

	if len(b) != len(bexpected) {
		t.Errorf("actual revision length does not match expected: expected = '%d', actual = '%d'\n", len(bexpected), len(b))
	}

	for i := range b {
		if b[i] != bexpected[i] {
			t.Errorf("actual revision character does not match expected: expected = '%s', actual = '%s'\n", string(bexpected[i]), string(b[i]))
		}
	}

	os.Remove(r.Path)
}

func TestFind(t *testing.T) {
	expectedId := int64(1136214245)
	expectedAuthor := "test <test@example.com>"

	expectedUp := "CREATE TABLE example();\n"
	expectedDown := "DROP TABLE example;\n"

	r, err := Find("1136214245")

	if err != nil {
		t.Errorf("failed to find revision: %s\n", err)
	}

	if r.ID != expectedId {
		t.Errorf("actual revision id does not match expected: expected = '%d', actual = '%d'\n", expectedId, r.ID)
	}

	if r.Author != expectedAuthor {
		t.Errorf("actual revision author does not match expected: expected = '%s', actual = '%s'\n", expectedAuthor, r.Author)
	}

	if r.up != expectedUp {
		t.Errorf("actual revision up does not match expected: expected = '%s', actual = '%s'\n", expectedUp, r.up)
	}

	if r.down != expectedDown {
		t.Errorf("actual revision down does not match expected: expected = '%s', actual = '%s'\n", expectedDown, r.down)
	}
}

func TestMain(m *testing.M) {
	config.Root = "testdata"

	exitCode := m.Run()

	os.Exit(exitCode)
}
