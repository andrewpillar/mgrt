package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	revisionsDir = "revisions"
	revisionTmpl = `/*
 * %s
 */
-- +up

-- +down
`

	AddCmd = &Command{
		Usage: "add [comment]",
		Short: "add a new revision",
		Long:  `add will open up the editor specified via EDITOR for creating the new revision.`,
		Run:   addCmd,
	}

	reSlug = regexp.MustCompile("[^a-zA-Z0-9]")
	reDupe = regexp.MustCompile("-{2,}")
)

func addCmd(_ *Command, args []string) error {
	if err := os.MkdirAll(revisionsDir, os.FileMode(0755)); err != nil {
		return err
	}

	name := time.Now().Format("2006-01-02T15-04-05")

	var comment string

	if len(args) > 0 {
		comment = args[0]

		s := strings.TrimSpace(comment)
		s = reSlug.ReplaceAllString(s, "-")
		s = reDupe.ReplaceAllString(s, "-")

		name += "-" + s
	}

	name += ".sql"

	f, err := os.Create(filepath.Join(revisionsDir, name))

	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := fmt.Fprintf(f, revisionTmpl, comment); err != nil {
		return err
	}

	editor := os.Getenv("EDITOR")

	if editor == "" {
		return errors.New("EDITOR not set")
	}

	cmd := exec.Command(editor, f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println("revision created", filepath.Join(revisionsDir, name))

	return nil
}
