package internal

import (
	"fmt"
	"os"

	"github.com/andrewpillar/mgrt/v3"
)

var LsCmd = &Command{
	Usage: "ls",
	Short: "list revisions",
	Long:  `List will display all of the revisions you have.`,
	Run:   lsCmd,
}

func lsCmd(cmd *Command, args []string) {
	info, err := os.Stat(revisionsDir)

	if err != nil {
		if os.IsNotExist(err) {
			return
		}

		fmt.Fprintf(os.Stderr, "%s %s: failed to list revisions: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "%s %s: %s is not a directory\n", cmd.Argv0, args[0], revisionsDir)
		os.Exit(1)
	}

	pad := 0

	revs, err := mgrt.LoadRevisions(revisionsDir)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to list revision: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	for _, rev := range revs {
		if l := len(rev.Author); l > pad {
			pad = l
		}
	}

	for _, r := range revs {
		if r.Comment != "" {
			fmt.Printf("%s: %-*s - %s\n", r.Slug(), pad, r.Author, r.Title())
			continue
		}
		fmt.Printf("%s: %s\n", r.Slug(), r.Author)
	}
}
