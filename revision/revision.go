package revision

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/andrewpillar/mgrt/config"
)

var (
	stub = `-- mgrt: revision: %d: %s
-- mgrt: up

-- mgrt: down

`

	append_ = func(revisions []*Revision, r *Revision) []*Revision {
		return append(revisions, r)
	}

	prepend_ = func(revisions []*Revision, r *Revision) []*Revision {
		return append([]*Revision{r}, revisions...)
	}
)

type appendFunc func(revisions []*Revision, r *Revision) []*Revision

type errMalformedRevision struct {
	file string
	line int
	err  error
}

type Revision struct {
	up   string
	down string

	ID        int64
	Message   string
	Hash      [sha256.Size]byte
	Direction Direction
	CreatedAt *time.Time
	Path      string
}

func Add(msg string) (*Revision, error) {
	id := time.Now().Unix()

	path := filepath.Join(config.RevisionsDir(), strconv.FormatInt(id, 10) + ".sql")

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, config.FileMode)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	r := &Revision{
		ID:      id,
		Message: msg,
		Path:    path,
	}

	_, err = fmt.Fprintf(f, stub, r.ID, r.Message)

	return r, err
}

func Find(id string) (*Revision, error) {
	return resolveFromPath(filepath.Join(config.RevisionsDir(), id + ".sql"))
}

func Oldest() ([]*Revision, error) {
	return walk(append_)
}

func Latest() ([]*Revision, error) {
	return walk(prepend_)
}

func resolveFromPath(path string) (*Revision, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	s := bufio.NewScanner(f)
	r := &Revision{}

	var directive string
	var direction Direction
	var up, down, hash bytes.Buffer

	for i := 1; s.Scan(); i++ {
		line := s.Text()

		if strings.HasPrefix(line, "-- mgrt:") {
			parts := strings.Split(line, ":")
			l := len(parts)

			if l < 2 {
				return nil, &errMalformedRevision{
					file: path,
					line: i,
					err:  errors.New("expected directive"),
				}
			}

			directive = strings.TrimPrefix(parts[1], " ")

			if directive == "revision" {
				id, err := strconv.ParseInt(strings.TrimPrefix(parts[2], " "), 10, 64)

				if err != nil {
					return nil, &errMalformedRevision{
						file: path,
						line: i,
						err:  err,
					}
				}

				r.ID = id

				if l == 4 {
					r.Message = strings.TrimPrefix(parts[3], " ")
				}

				directive = ""
				continue
			}

			if directive == "up" {
				direction = Up
				continue
			}

			if directive == "down" {
				direction = Down
				continue
			}
		}

		if direction == Up {
			hash.WriteString(line)
			up.WriteString(line + "\n")
		}

		if direction == Down {
			hash.WriteString(line)
			down.WriteString(line + "\n")
		}
	}

	r.up = up.String()
	r.down = down.String()
	r.Hash = sha256.Sum256(hash.Bytes())

	return r, nil
}

func walk(f appendFunc) ([]*Revision, error) {
	revisions := make([]*Revision, 0)

	realWalk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == config.RevisionsDir() {
			return nil
		}

		r, err := resolveFromPath(path)

		if err != nil {
			return err
		}

		revisions = f(revisions, r)
		return nil
	}

	err := filepath.Walk(config.RevisionsDir(), realWalk)

	return revisions, err
}

func (e *errMalformedRevision) Error() string {
	return fmt.Sprintf("malformed revision: %s:%d: %s", e.file, e.line, e.err)
}

// Revisions returned from the database log will not have the Message, up, or down properties
// populated. The Load method will populate these properties by looking up the revision on the
// filesystem.
func (r *Revision) Load() error {
	realrev, err := Find(strconv.FormatInt(r.ID, 10))

	if err != nil {
		return err
	}

	r.Message = realrev.Message
	r.up = realrev.up
	r.down = realrev.down

	return nil
}

func (r *Revision) Query() string {
	switch r.Direction {
		case Up:
			return r.up
		case Down:
			return r.down
		default:
			return ""
	}
}
