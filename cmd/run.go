package cmd

import (
	"fmt"

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

	var revisions []*revision.Revision

	if len(c.Args) == 0 {
		revisions, err = revision.Oldest()
	} else {
		for _, id := range c.Args {
			r, err := revision.Find(id)

			if err != nil {
				continue
			}

			revisions = append(revisions, r)
		}
	}

	if err != nil {
		util.ExitError("failed to load revisions", err)
	}

	for _, r := range revisions {
		if err := db.Run(r, revision.Up); err != nil {
			if err != database.ErrAlreadyRan && err != database.ErrChecksumFailed {
				util.ExitError("failed to run revision", err)
			}

			fmt.Printf("[ WARN ] %s: %s", err, r.ID)

			if r.Message != "" {
				fmt.Printf(": %s", r.Message)
			}

			fmt.Printf("\n")
			continue
		}

		if err := db.Log(r, revision.Up); err != nil {
			util.ExitError("failed to log revision", err)
		}

		fmt.Printf("[  OK  ] ran revision: %s", r.ID)

		if r.Message != "" {
			fmt.Printf(": %s", r.Message)
		}

		fmt.Printf("\n")
	}
}
