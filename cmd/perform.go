package cmd

import (
	"errors"
	"fmt"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/database"
	"github.com/andrewpillar/mgrt/revision"
	"github.com/andrewpillar/mgrt/util"
)

func loadRevisions(c cli.Command, d revision.Direction) ([]*revision.Revision, error) {
	var revisions []*revision.Revision

	if len(c.Args) > 0 {
		for _, id := range c.Args {
			r, err := revision.Find(id)

			if err != nil {
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
		util.ExitError("not initialize", err)
	}

	cfg, err := config.Open()

	if err != nil {
		util.ExitError("failed to open config", err)
	}

	defer cfg.Close()

	if cfg.Type == "" {
		util.ExitError("failed to perform revisions", errors.New("database type not specified"))
	}

	if cfg.Address == "" {
		util.ExitError("failed to perform revisions", errors.New("database address not specified"))
	}

	db, err := database.Open(cfg)

	if err != nil {
		util.ExitError("failed to open database", err)
	}

	defer db.Close()

	if c.Flags.IsSet("force") {
		_, err := db.Exec("DROP TABLE mgrt_revisions")

		if err != nil {
			util.ExitError("failed to force revisions", err)
		}

		if err := db.Init(); err != nil && err != database.ErrInitialized {
			util.ExitError("failed to force revisions", err)
		}

		forceDirection := d.Invert()

		revisions, err := loadRevisions(c, forceDirection)

		if err != nil {
			util.ExitError("failed to force revisions", err)
		}

		for _, r := range revisions {
			r.Direction = forceDirection

			if err := db.Perform(r); err != nil && err != database.ErrAlreadyPerformed {
				util.ExitError("failed to force revisions", err)
			}
		}
	}

	if err := db.Init(); err != nil && err != database.ErrInitialized {
		util.ExitError("failed to initialize database", err)
	}

	revisions, err := loadRevisions(c, d)

	if err != nil {
		util.ExitError("failed to load revisions", err)
	}

	for _, r := range revisions {
		r.Direction = d

		if err := db.Perform(r); err != nil {
			if err != database.ErrAlreadyPerformed {
				util.ExitError("failed to perform revision", fmt.Errorf("%s: %d", err, r.ID))
			}

			fmt.Printf("%s - %s: %d", d, err, r.ID)

			if r.Message != "" {
				fmt.Printf(": %s", r.Message)
			}

			fmt.Printf("\n")
			continue
		}

		if err := db.Log(r); err != nil {
			util.ExitError("failed to log revision", err)
		}

		fmt.Printf("%s - performed revision: %d", d, r.ID)

		if r.Message != "" {
			fmt.Printf(": %s", r.Message)
		}

		fmt.Printf("\n")
	}
}
