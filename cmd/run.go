package cmd

import (
	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/database"
	"github.com/andrewpillar/mgrt/revision"
	"github.com/andrewpillar/mgrt/util"
)

func Run(c cli.Command) {
	if err := config.Initialized(); err != nil {
		util.ExitError("not initialized", err)
	}

	cfg, err := config.Open()

	if err != nil {
		util.ExitError("failed to open config", err)
	}

	defer cfg.Close()

	db, err := database.Open(cfg)

	if err != nil {
		util.ExitError("failed to open database", err)
	}

	defer db.Close()

	if err := db.Init(); err != nil && err != database.ErrInitialized {
		util.ExitError("failed to initialize database", err)
	}

	revisions, err := revision.Oldest()

	if err != nil {
		util.ExitError("failed to load revisions", err)
	}

	direction := revision.Up

	if c.Flags.IsSet("down") {
		direction = revision.Down
	}

	for _, r := range revisions {
		if err := db.Run(r, direction); err != nil {
			util.ExitError("failed to run revision", err)
		}

		if err := db.Log(r, direction); err != nil {
			util.ExitError("failed to log revision", err)
		}
	}
}
