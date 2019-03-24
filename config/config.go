package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var (
	file      = "mgrt.yml"
	revisions = "revisions"

	stub = `

# The type of database, one of:
#   - postgres
#   - mysql
#   - sqlite3
type:

# The database address, if SQLite then the filepath instead.
address:

# Login credentials for the user that will run the revisions.
username:
password:

# Database to run the revisions against, if using SQLite then leave empty.
database:

# Details about the person creating the database revisions.
author:
  name:
  email:
`

	Root string

	DirMode  = os.FileMode(0755)
	FileMode = os.FileMode(0644)
)

type Config struct {
	*os.File

	Type     string
	Address  string
	Username string
	Password string
	Database string

	Author struct {
		Name  string
		Email string
	}
}

func Initialized() error {
	dir := filepath.Join(Root, revisions)

	for _, f := range []string{Root, filepath.Join(Root, file), dir} {
		info, err := os.Stat(f)

		if err != nil {
			return err
		}

		if f == Root || f == dir {
			if !info.IsDir() {
				return errors.New("not a directory " + f)
			}
		}
	}

	return nil
}

func Create() error {
	f, err := os.OpenFile(filepath.Join(Root, file), os.O_CREATE|os.O_WRONLY, FileMode)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write([]byte(stub))

	return err
}

func RevisionsDir() string {
	return filepath.Join(Root, revisions)
}

func Open() (*Config, error) {
	f, err := os.Open(filepath.Join(Root, file))

	if err != nil {
		return nil, err
	}

	dec := yaml.NewDecoder(f)

	cfg := &Config{
		File: f,
	}

	if err := dec.Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
