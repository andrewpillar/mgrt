package revision

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andrewpillar/mgrt/config"
)

var (
	upFile   = "up.sql"
	downFile = "down.sql"

	reslug = regexp.MustCompile("[^a-zA-Z0-9]")
	redup  = regexp.MustCompile("-{2,}")

	append_ = func(revisions []*Revision, r *Revision) []*Revision {
		return append(revisions, r)
	}

	prepend_ = func(revisions []*Revision, r *Revision) []*Revision {
		return append([]*Revision{r}, revisions...)
	}
)

type appendFunc func(revisions []*Revision, r *Revision) []*Revision

type Revision struct {
	path string

	ID        int64
	Message   string
	Hash      [sha256.Size]byte
	Direction Direction
	Up        sql.NullString
	Down      sql.NullString
	Forced    bool
	CreatedAt *time.Time
	UpPath    string
	DownPath  string
}

func resolveFromPath(path string) (*Revision, error) {
	info, err := os.Stat(path)

	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, errors.New("invalid revision: not a directory: " + info.Name())
	}

	parts := strings.Split(filepath.Base(path), "_")

	id, err := strconv.ParseInt(parts[0], 10, 64)

	if err != nil {
		return nil, errors.New("invalid revision: " + err.Error())
	}

	r := &Revision{
		path: path,
		ID:   id,
	}

	if len(parts) > 1 {
		r.Message = strings.Join(append([]string{strings.Title(parts[1])}, parts[2:]...), " ")
	}

	b, err := ioutil.ReadFile(filepath.Join(r.path, upFile))

	if err != nil {
		return nil, err
	}

	r.Up = sql.NullString{
		String: string(b),
		Valid:  true,
	}

	b, err = ioutil.ReadFile(filepath.Join(r.path, downFile))

	if err != nil {
		return nil, err
	}

	r.Down = sql.NullString{
		String: string(b),
		Valid:  true,
	}

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

func Add(msg string) (*Revision, error) {
	id := time.Now().Unix()

	slug := redup.ReplaceAllString(reslug.ReplaceAllString(strings.TrimSpace(msg), "_"), "_")

	path := filepath.Join(config.RevisionsDir(), strconv.FormatInt(id, 10))

	if slug != "" {
		path += "_" + strings.ToLower(slug)
	}

	if err := os.MkdirAll(path, config.DirMode); err != nil {
		return nil, err
	}

	upPath := filepath.Join(path, upFile)
	downPath := filepath.Join(path, downFile)

	var (
		f   *os.File
		err error
	)

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

	return &Revision{
		path:     path,
		ID:       id,
		Message:  msg,
		DownPath: downPath,
		UpPath:   upPath,
	}, nil
}

func Find(id string) (*Revision, error) {
	dir := config.RevisionsDir()

	infos, err := ioutil.ReadDir(dir)

	if err != nil {
		return nil, err
	}

	base := ""

	for _, info := range infos {
		name := info.Name()

		parts := strings.Split(name, "_")

		if parts[0] == id {
			base = name
			break
		}
	}

	if base == "" {
		return nil, errors.New("no revision found with ID: " + id)
	}

	return resolveFromPath(filepath.Join(config.RevisionsDir(), base))
}

func Oldest() ([]*Revision, error) {
	return walk(append_)
}

func Latest() ([]*Revision, error) {
	return walk(prepend_)
}

func (r *Revision) GenHash() error {
	buf := &bytes.Buffer{}

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

func (r *Revision) Query() string {
	if r.Direction == Up && r.Up.Valid {
		return r.Up.String
	}

	if r.Direction == Down && r.Down.Valid {
		return r.Down.String
	}

	return "---\n"
}
