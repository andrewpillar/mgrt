package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type dbItem struct {
	Name string
	Type string
	DSN  string
}

var (
	DBLsCmd = &Command{
		Usage: "ls",
		Short: "list the databases configured",
		Run:   dbLsCmd,
	}

	DBSetCmd = &Command{
		Usage: "set <name> <type> <dsn>",
		Short: "set the database connection",
		Run:   dbSetCmd,
	}

	DBRmCmd = &Command{
		Usage: "rm <name,...>",
		Short: "remove a database connection",
		Run:   dbRmCmd,
	}
)

func mgrtdir() (string, error) {
	cfgdir, err := os.UserConfigDir()

	if err != nil {
		return "", err
	}

	dir := filepath.Join(cfgdir, "mgrt")

	info, err := os.Stat(dir)

	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.FileMode(0700)); err != nil {
				return "", err
			}
			return dir, nil
		}
		return "", err
	}

	if !info.IsDir() {
		return "", errors.New("not a directory " + dir)
	}
	return dir, nil
}

func getdbitem(name string) (dbItem, error) {
	dir, err := mgrtdir()

	if err != nil {
		return dbItem{}, err
	}

	f, err := os.Open(filepath.Join(dir, name))

	if err != nil {
		return dbItem{}, err
	}

	defer f.Close()

	it := dbItem{
		Name: name,
	}

	if err := json.NewDecoder(f).Decode(&it); err != nil {
		return it, err
	}
	return it, nil
}

func DBCmd(argv0 string) *Command {
	cmd := &Command{
		Usage: "db <command> [arguments]",
		Short: "manage configured databases",
		Run:   dbCmd,
		Commands: &CommandSet{
			Argv0: argv0 + " db",
		},
	}

	cmd.Commands.Add("ls", DBLsCmd)
	cmd.Commands.Add("rm", DBRmCmd)
	cmd.Commands.Add("set", DBSetCmd)
	return cmd
}

func dbCmd(cmd *Command, args []string) {
	if len(args[1:]) < 1 {
		fmt.Println("usage:", cmd.Argv0, cmd.Usage)
	}

	if err := cmd.Commands.Parse(args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}
}

func dbLsCmd(cmd *Command, args []string) {
	argv0 := args[0]

	dir, err := mgrtdir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
		os.Exit(1)
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		b, err := os.ReadFile(path)

		if err != nil {
			return err
		}

		it := dbItem{
			Name: filepath.Base(path),
		}

		if err := json.Unmarshal(b, &it); err != nil {
			return err
		}

		fmt.Println(it.Name, it.Type, it.DSN)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
		os.Exit(1)
	}
}

func dbSetCmd(cmd *Command, args []string) {
	argv0 := args[0]

	if len(args[1:]) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s %s <name> <type> <dsn>\n", cmd.Argv0, argv0)
		os.Exit(1)
	}

	dir, err := mgrtdir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
		os.Exit(1)
	}

	it := dbItem{
		Name: args[1],
		Type: args[2],
		DSN:  args[3],
	}

	f, err := os.OpenFile(filepath.Join(dir, it.Name), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0400))

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
		os.Exit(1)
	}

	defer f.Close()

	if err := json.NewEncoder(f).Encode(&it); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
		os.Exit(1)
	}
}

func dbRmCmd(cmd *Command, args []string) {
	argv0 := args[0]

	if len(args[1:]) < 1 {
		fmt.Fprintf(os.Stderr, "usage: %s %s <name,...>\n", cmd.Argv0, argv0)
		os.Exit(1)
	}

	dir, err := mgrtdir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
		os.Exit(1)
	}

	for _, name := range args[1:] {
		os.Remove(filepath.Join(dir, name))
	}
}
