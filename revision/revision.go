package revision

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
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
	Up   *bytes.Buffer
	Down *bytes.Buffer

	ID        int64
	Author    string
	Message   string
	Hash      [sha256.Size]byte
	Direction Direction
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

	var fup, fdown, fmessage *os.File

	fup, err = os.Open(filepath.Join(path, upFile))

	if err != nil {
		return nil, err
	}

	defer fup.Close()

	fdown, err = os.Open(filepath.Join(path, downFile))

	if err != nil {
		return nil, err
	}

	defer fdown.Close()

	r := &Revision{
		Up:     &bytes.Buffer{},
		Down:   &bytes.Buffer{},
		ID:     id,
	}

	_, err = io.Copy(r.Up, fup)

	if err != nil {
		return nil, err
	}

	_, err = io.Copy(r.Down, fdown)

	if err != nil {
		return nil, err
	}

	messageBuf := &bytes.Buffer{}

	fmessage, err = os.Open(filepath.Join(path, messageFile))

	if err != nil {
		return nil, err
	}

	defer fmessage.Close()

	s := bufio.NewScanner(fmessage)

	s.Scan()

	line := s.Text()

	if !strings.HasPrefix(line, "Author:") {
		return nil, errors.New("invalid revision: missing revision author")
	}

	for s.Scan() {
		messageBuf.Write(s.Bytes())
		messageBuf.Write([]byte{'\n'})
	}

	if err != nil {
		return nil, err
	}

	parts := strings.Split(line, ":")

	r.Author = strings.TrimSpace(parts[1])
	r.Message = strings.TrimSuffix(messageBuf.String(), "\n")

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
		l = r.Up.Len()
		b = r.Up.Bytes()
	}

	if r.Direction == Down {
		l = r.Down.Len()
		b = r.Down.Bytes()
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

// Revisions returned from the database log will not have the Message, up, or down properties
// populated. The Load method will populate these properties by looking up the revision on the
// filesystem.
func (r *Revision) Load() error {
	realrev, err := Find(strconv.FormatInt(r.ID, 10))

	if err != nil {
		return err
	}

	r.Up = realrev.Up
	r.Down = realrev.Down

	return nil
}

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
	if r.Direction == Up && r.Up.Len() != 0 {
		return r.Up.String()
	}

	if r.Direction == Down && r.Down.Len() != 0 {
		return r.Down.String()
	}

	return "---\n"
}
