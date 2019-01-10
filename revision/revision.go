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

const (
	Up Direction	= iota
	Down
)

var (
	stubf = `-- mgrt: revision: %s: %s
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
}

type Direction uint32

type Revision struct {
	ID      string
	Message string
	Up      string
	Down    string
	Hash    [sha256.Size]byte
	Path    string
}

func Add(msg string) (*Revision, error) {
	id := strconv.FormatInt(time.Now().Unix(), 10)

	path := filepath.Join(config.RevisionsDir, id + ".sql")

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

	_, err = fmt.Fprintf(f, stubf, r.ID, r.Message)

	return r, err
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
				}
			}

			directive = strings.TrimPrefix(parts[1], " ")

			if directive == "revision" {
				r.ID = strings.TrimPrefix(parts[2], " ")

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

	r.Up = up.String()
	r.Down = down.String()
	r.Hash = sha256.Sum256(hash.Bytes())

	return r, nil
}

func walk(f appendFunc) ([]*Revision, error) {
	revisions := make([]*Revision, 0)

	realWalk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == config.RevisionsDir {
			return nil
		}

		r, err := resolveFromPath(path)

		if err != nil {
			return err
		}

		revisions = f(revisions, r)
		return nil
	}

	err := filepath.Walk(config.RevisionsDir, realWalk)

	return revisions, err
}

func (e *errMalformedRevision) Error() string {
	return fmt.Sprintf("malformed revision: %s:%d", e.file, e.line)
}

func (d Direction) String() string {
	switch d {
		case Up:
			return "up"
		case Down:
			return "down"
		default:
			return ""
	}
}

func (d *Direction) Scan(src interface{}) error {
	s, ok := src.(string)

	if !ok {
		return errors.New("failed to scan direction type: could not type assert to string")
	}

	if s == "up" {
		(*d) = Up
		return nil
	}

	if s == "down" {
		(*d) = Down
		return nil
	}

	return errors.New("failed to scan direction type: unknown direction: " + s)
}
