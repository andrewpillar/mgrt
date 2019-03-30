package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/revision"
	"github.com/andrewpillar/mgrt/util"
)

func Cat(c cli.Command) {
	config.Root = c.Flags.GetString("config")

	if err := config.Initialized(); err != nil {
		util.ExitError("not initialize", err)
	}

	if !c.Flags.IsSet("up") && !c.Flags.IsSet("down") {
		util.ExitError("missing one of two flags", errors.New("--up, --down"))
	}

	for _, id := range c.Args {
		r, err := revision.Find(id)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: failed to find revision: %s\n", os.Args[0], id)
			continue
		}

		if c.Flags.IsSet("up") {
			io.Copy(os.Stdout, r.Up)
		}

		if c.Flags.IsSet("down") {
			io.Copy(os.Stdout, r.Down)
		}
	}
}
