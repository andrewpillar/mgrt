package main

import (
	"flag"
	"fmt"
	"os"
)

var Build string

func run(args []string) error {
	cmds := CommandSet{
		Argv0: args[0],
		Long: `mgrt performs and keeps track of SQL revisions performed against a database.
This should be used during development to allow for quick iteration on database
schemas.

Usage:

    mgrt <command> [arguments]
`,
	}

	cmds.Add("add", AddCmd)
	cmds.Add("up", UpCmd)
	cmds.Add("down", DownCmd)
	cmds.Add("log", LogCmd)

	cmds.Add("help", HelpCmd(&cmds))

	var version bool

	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	fs.BoolVar(&version, "version", false, "display version information and exit")
	fs.Parse(args[1:])

	if version {
		fmt.Println(Build)
		return nil
	}
	return cmds.Parse(args[1:])
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
