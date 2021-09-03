package internal

import (
	"flag"
	"fmt"
	"os"

	"github.com/andrewpillar/mgrt/v3"
)

var CatCmd = &Command{
	Usage: "cat <revisions,...>",
	Short: "display the given revisions",
	Long:  `Cat will print out the given revisions`,
	Run:   catCmd,
}

func catCmd(cmd *Command, args []string) {
	argv0 := args[0]

	var sql bool

	fs := flag.NewFlagSet(cmd.Argv0+" "+argv0, flag.ExitOnError)
	fs.BoolVar(&sql, "sql", false, "only display the sql of the revision")
	fs.Parse(args[1:])

	args = fs.Args()

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "usage: %s %s <revisions,...>\n", cmd.Argv0, argv0)
		os.Exit(1)
	}

	info, err := os.Stat(revisionsDir)

	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s %s: no migrations created\n", cmd.Argv0, argv0)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%s %s: failed to cat revision(s): %s\n", cmd.Argv0, argv0, err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "%s %s: %s is not a directory\n", cmd.Argv0, argv0, revisionsDir)
		os.Exit(1)
	}

	for _, id := range args {
		r, err := mgrt.OpenRevision(revisionPath(id))

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: failed to cat revision: %s\n", cmd.Argv0, argv0, err)
			os.Exit(1)
		}

		if sql {
			fmt.Println(r.SQL)
			return
		}
		fmt.Println(r.String())
	}
}
