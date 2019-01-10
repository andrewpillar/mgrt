package cmd

import (
	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/revision"
	"github.com/andrewpillar/mgrt/util"
)

func Add(c cli.Command) {
	if err := config.Initialized(); err != nil {
		util.ExitError("not initialized", err)
	}

	r, err := revision.Add(c.Flags.GetString("message"))

	if err != nil {
		util.ExitError("failed to add revision", err)
	}

	util.OpenInEditor(r.Path)
}
