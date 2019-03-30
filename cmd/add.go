package cmd

import (
	"errors"
	"fmt"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/revision"
	"github.com/andrewpillar/mgrt/util"
)

func Add(c cli.Command) {
	config.Root = c.Flags.GetString("config")

	if err := config.Initialized(); err != nil {
		util.ExitError("not initialized", err)
	}

	cfg, err := config.Open()

	if err != nil {
		util.ExitError("failed to open config", err)
	}

	defer cfg.Close()

	if cfg.Author.Name == "" || cfg.Author.Email == "" {
		util.ExitError("failed to perform revisions", errors.New("name and email not specified"))
	}

	r, err := revision.Add(c.Flags.GetString("message"), cfg.Author.Name, cfg.Author.Email)

	if err != nil {
		util.ExitError("failed to add revision", err)
	}

	if r.Message == "" {
		util.OpenInEditor(r.MessagePath)
	}

	fmt.Println("added new revision at:")
	fmt.Println("  ", r.UpPath)
	fmt.Println("  ", r.DownPath)
}
