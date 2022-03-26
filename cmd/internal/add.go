package internal

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/andrewpillar/mgrt/v3"
)

var (
	revisionsDir = "revisions"

	AddCmd = &Command{
		Usage: "add [comment]",
		Short: "add a new revision",
		Long: `Add will open up the editor specified via EDITOR for creating the new revision.
The -c flag can be given to specify a category for the new revision.`,
		Run: addCmd,
	}
)

func revisionPath(id string) string {
	return filepath.Join(revisionsDir, id+".sql")
}

func openInEditor(path string) error {
	editor := os.Getenv("EDITOR")

	if editor == "" {
		return errors.New("EDITOR not set")
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func addCmd(cmd *Command, args []string) {
	var category string

	argv0 := args[0]

	fs := flag.NewFlagSet(cmd.Argv0+" "+argv0, flag.ExitOnError)
	fs.StringVar(&category, "c", "", "the category to put the revision under")
	fs.Parse(args[1:])

	args = fs.Args()

	var comment string

	if len(args) >= 1 {
		comment = args[0]
	}

	dir := revisionsDir

	if category != "" {
		dir = filepath.Join(revisionsDir, category)
	}

	if err := os.MkdirAll(dir, os.FileMode(0755)); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to create %s directory: %s", cmd.Argv0, args[0], revisionsDir, err)
		os.Exit(1)
	}

	author, err := mgrtAuthor()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to get mgrt author: %s", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	var rev *mgrt.Revision

	if category != "" {
		rev = mgrt.NewRevisionCategory(category, author, comment)
	} else {
		rev = mgrt.NewRevision(author, comment)
	}

	path := filepath.Join(dir, rev.ID+".sql")

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, os.FileMode(0644))

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to create revision: %s", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	defer f.Close()

	f.WriteString(rev.String())

	if err := openInEditor(path); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to open revision file: %s", cmd.Argv0, args[0], err)
		os.Exit(1)
	}
	fmt.Println("revision created", rev.Slug())
}
