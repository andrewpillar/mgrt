package cmd

import (
	"fmt"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/revision"
	"github.com/andrewpillar/mgrt/util"
)

func Ls(c cli.Command) {
	config.Root = c.Flags.GetString("config")

	if err := config.Initialized(); err != nil {
		util.ExitError("not initialized", err)
	}

	var revisions []*revision.Revision
	var err error

	if c.Flags.IsSet("reverse") {
		revisions, err = revision.Latest()
	} else {
		revisions, err = revision.Oldest()
	}

	if err != nil {
		util.ExitError("failed to load revisions", err)
	}

	for _, r := range revisions {
		fmt.Printf("%d", r.ID)

		if r.Message != "" {
			fmt.Printf(" - %s", r.Message)
		}

		fmt.Printf("\n")
	}
}
