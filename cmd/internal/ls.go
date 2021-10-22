package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	revs := make([]*mgrt.Revision, 0)

	err = filepath.Walk(revisionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		rev, err := mgrt.OpenRevision(path)

		if err != nil {
			return err
		}

		rev.ID = strings.TrimPrefix(path, revisionsDir+string(os.PathSeparator))

		if l := len(rev.Author); l > pad {
			pad = l
		}

		revs = append(revs, rev)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to list revision: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	for _, r := range revs {
		if r.Comment != "" {
			fmt.Printf("%s: %-*s - %s\n", r.ID, pad, r.Author, r.Title())
			continue
		}
		fmt.Printf("%s: %s\n", r.ID, r.Author)
	}
}
