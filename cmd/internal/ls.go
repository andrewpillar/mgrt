package internal

import (
	"flag"
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
	var category string

	fs := flag.NewFlagSet(cmd.Argv0, flag.ExitOnError)
	fs.StringVar(&category, "c", "", "the category to list the revisions of")
	fs.Parse(args[1:])

	info, err := os.Stat(revisionsDir)

	if err != nil {
		if os.IsNotExist(err) {
			return
		}

		fmt.Fprintf(os.Stderr, "%s: failed to list revisions: %s\n", cmd.Argv0, err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "%s: %s is not a directory\n", cmd.Argv0, revisionsDir)
		os.Exit(1)
	}

	pad := 0

	revs, err := mgrt.LoadRevisions(revisionsDir)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: failed to list revision: %s\n", cmd.Argv0, err)
		os.Exit(1)
	}

	for _, rev := range revs {
		if l := len(rev.Author); l > pad {
			pad = l
		}
	}

	show := true

	for _, r := range revs {
		if category != "" {
			show = r.Category == category
		}

		if show {
			if r.Comment != "" {
				fmt.Printf("%s: %-*s - %s\n", r.Slug(), pad, r.Author, r.Title())
				continue
			}
			fmt.Printf("%s: %s\n", r.Slug(), r.Author)
		}
	}
}
