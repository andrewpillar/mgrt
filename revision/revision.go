package revision

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/andrewpillar/mgrt/config"
)

var (
	upFile      = "up.sql"
	downFile    = "down.sql"
	messageFile = "_message"

	append_ = func(revisions []*Revision, r *Revision) []*Revision {
		return append(revisions, r)
	}

	prepend_ = func(revisions []*Revision, r *Revision) []*Revision {
		return append([]*Revision{r}, revisions...)
	}
)

type appendFunc func(revisions []*Revision, r *Revision) []*Revision

type Revision struct {
	ID        int64
	Author    string
	Message   string
	Hash      [sha256.Size]byte
	Direction Direction
	Up        sql.NullString
	Down      sql.NullString
	Forced    bool
	CreatedAt *time.Time

	MessagePath string
	UpPath      string
	DownPath    string
}

func Add(msg, name, email string) (*Revision, error) {
	id := time.Now().Unix()

	path := filepath.Join(config.RevisionsDir(), strconv.FormatInt(id, 10))

	if err := os.MkdirAll(path, config.DirMode); err != nil {
		return nil, err
	}

	upPath := filepath.Join(path, upFile)
	downPath := filepath.Join(path, downFile)
	messagePath := filepath.Join(path, messageFile)

	var f *os.File
	var err error

	f, err = os.Create(upPath)

	if err != nil {
		return nil, err
	}

	f.Close()

	f, err = os.Create(downPath)

	if err != nil {
		return nil, err
	}

	f.Close()

	f, err = os.OpenFile(messagePath, os.O_CREATE|os.O_WRONLY, config.FileMode)

	if err != nil {
		return nil, err
	}

	author := fmt.Sprintf("%s <%s>", name, email)

	_, err = fmt.Fprintf(f, "Author: %s\n", author)

	if msg != "" {
		_, err = f.Write([]byte(msg))

		if err != nil {
			return nil, err
		}
	}

	f.Close()

	return &Revision{
		ID:          id,
		Author:      author,
		Message:     msg,
		MessagePath: messagePath,
		DownPath:    downPath,
		UpPath:      upPath,
	}, nil
}

func Find(id string) (*Revision, error) {
	return resolveFromPath(filepath.Join(config.RevisionsDir(), id))
}

func Oldest() ([]*Revision, error) {
	return walk(append_)
}

func Latest() ([]*Revision, error) {
	return walk(prepend_)
}

func resolveFromPath(path string) (*Revision, error) {
	info, err := os.Stat(path)

	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, errors.New("invalid revision: not a directory: " + info.Name())
	}

	id, err := strconv.ParseInt(filepath.Base(path), 10, 64)

	if err != nil {
		return nil, errors.New("invalid revision: " + err.Error())
	}

	r := &Revision{
		ID: id,
	}

	b, err := ioutil.ReadFile(filepath.Join(path, upFile))

	if err != nil {
		return nil, err
	}

	r.Up = sql.NullString{
		String: string(b),
		Valid:  true,
	}

	b, err = ioutil.ReadFile(filepath.Join(path, downFile))

	if err != nil {
		return nil, err
	}

	r.Down = sql.NullString{
		String: string(b),
		Valid:  true,
	}

	buf := &bytes.Buffer{}

	f, err := os.Open(filepath.Join(path, messageFile))

	if err != nil {
		return nil, err
	}

	defer f.Close()

	s := bufio.NewScanner(f)
	s.Scan()

	line := s.Text()

	if !strings.HasPrefix(line, "Author:") {
		return nil, errors.New("invalid revision: missing revision author")
	}

	for s.Scan() {
		buf.Write(s.Bytes())
		buf.Write([]byte{'\n'})
	}

	if err != nil {
		return nil, err
	}

	parts := strings.Split(line, ":")

	r.Author = strings.TrimSpace(parts[1])
	r.Message = strings.TrimSuffix(buf.String(), "\n")

	return r, nil
}

func walk(f appendFunc) ([]*Revision, error) {
	dir := config.RevisionsDir()

	files, err := ioutil.ReadDir(dir)

	if err != nil {
		return []*Revision{}, err
	}

	revisions := make([]*Revision, 0, len(files))

	for _, file := range files {
		r, err := resolveFromPath(filepath.Join(dir, file.Name()))

		if err != nil {
			return []*Revision{}, err
		}

		revisions = f(revisions, r)
	}

	return revisions, nil
}

func (r *Revision) GenHash() error {
	buf := bytes.NewBufferString(r.Author)

	b := []byte{}
	l := 0

	if r.Direction == Up {
		l = len(r.Up.String)
		b = []byte(r.Up.String)
	}

	if r.Direction == Down {
		l = len(r.Down.String)
		b = []byte(r.Down.String)
	}

	tmp := make([]byte, l, l)

	copy(tmp, b)

	if _, err := buf.Write(tmp); err != nil {
		return err
	}

	hash := sha256.Sum256(buf.Bytes())

	for i := range hash {
		r.Hash[i] = hash[i]
	}

	return nil
}

// Split the subject of the message from the body, and return them as two
// separate values.
func (r Revision) SplitMessage() (string, string) {
	s := bufio.NewScanner(strings.NewReader(r.Message))
	s.Scan()

	subject := s.Text()
	body := &bytes.Buffer{}

	for s.Scan() {
		body.Write(s.Bytes())
		body.Write([]byte{'\n'})
	}

	return subject, body.String()
}

func (r *Revision) Query() string {
	if r.Direction == Up && r.Up.Valid {
		return r.Up.String
	}

	if r.Direction == Down && r.Down.Valid {
		return r.Down.String
	}

	return "---\n"
}
