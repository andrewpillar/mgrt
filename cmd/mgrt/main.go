package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/andrewpillar/mgrt/v3/cmd/internal"
)

var Build string

func run(args []string) error {
	cmds := &internal.CommandSet{
		Argv0: args[0],
		Long: `mgrt is a simple migration tool.

Usage:

    mgrt [-version] <command> [arguments]
`,
	}

	cmds.Add("add", internal.AddCmd)
	cmds.Add("cat", internal.CatCmd)
	cmds.Add("db", internal.DBCmd(cmds.Argv0))
	cmds.Add("log", internal.LogCmd)
	cmds.Add("ls", internal.LsCmd)
	cmds.Add("run", internal.RunCmd)
	cmds.Add("show", internal.ShowCmd)
	cmds.Add("sync", internal.SyncCmd)
	cmds.Add("help", internal.HelpCmd(cmds))

	var version bool

	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	fs.BoolVar(&version, "version", false, "display version information and exit")
	fs.Parse(args[1:])

	if version {
		fmt.Println(Build)
		return nil
	}
	return cmds.Parse(fs.Args())
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}
}
