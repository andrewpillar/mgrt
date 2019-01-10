package cmd

import (
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/util"
)

func Init(c cli.Command) {
	if err := config.Initialized(); err == nil {
		util.ExitError("already initialized", nil)
	}

	if err := os.MkdirAll(config.RevisionsDir, config.DirMode); err != nil {
		util.ExitError("failed to initialize", err)
	}

	if err := config.Create(); err != nil {
		util.ExitError("failed to initialize", err)
	}
}
