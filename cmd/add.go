package cmd

import (
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

	r, err := revision.Add(c.Flags.GetString("message"))

	if err != nil {
		util.ExitError("failed to add revision", err)
	}

	if r.Message == "" {
		util.OpenInEditor(r.MessagePath)
	}

	fmt.Println("added new revision at:")
	fmt.Println(" ", r.UpPath)
	fmt.Println(" ", r.DownPath)
}
