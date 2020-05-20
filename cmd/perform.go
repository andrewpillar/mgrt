package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/database"
	"github.com/andrewpillar/mgrt/migration"
	"github.com/andrewpillar/mgrt/revision"
	"github.com/andrewpillar/mgrt/util"
)

func loadRevisions(c cli.Command, d revision.Direction) ([]*revision.Revision, error) {
	var revisions []*revision.Revision

	if len(c.Args) > 0 {
		for _, id := range c.Args {
			r, err := revision.Find(id)

			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: failed to find revision: %s\n", os.Args[0], id)
				continue
			}

			revisions = append(revisions, r)
		}

		return revisions, nil
	}

	var err error

	switch d {
	case revision.Up:
		revisions, err = revision.Oldest()
		break
	case revision.Down:
		revisions, err = revision.Latest()
		break
	default:
		err = errors.New("unknown direction")
		break
	}

	return revisions, err
}

func perform(c cli.Command, d revision.Direction) {
	config.Root = c.Flags.GetString("config")

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

	revisions, err := loadRevisions(c, d)

	if err != nil {
		util.ExitError("failed to load revisions", err)
	}

	force := c.Flags.IsSet("force")

	migration.Perform(db, revisions, d, force)
}
