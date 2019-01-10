package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var (
	file = filepath.Join(RootDir, "config")

	stub = `

# The type of database, one of:
#   - postgres
#   - mysql
#   - sqlite3
type:

# The database address, if SQLite then the filepath instead.
address:

# Login credentials for the user that will run the migrations.
username:
password:

# Database name, and respective table for storing migration logs. If using
# SQLite, then disregard the 'name' property.
database:
  name:
  table:
`

	RootDir = "mgrt"

	RevisionsDir = filepath.Join(RootDir, "revisions")

	DirMode  = os.FileMode(0755)
	FileMode = os.FileMode(0644)
)

type Config struct {
	*os.File

	Type     string
	Address  string
	Username string
	Password string

	Database struct {
		Name  string
		Table string
	}
}

func Initialized() error {
	for _, f := range []string{RootDir, file, RevisionsDir} {
		if _, err := os.Stat(f); err != nil {
			return err
		}
	}

	return nil
}

func Create() error {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, FileMode)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write([]byte(stub))

	return err
}

func Open() (*Config, error) {
	f, err := os.Open(file)

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
