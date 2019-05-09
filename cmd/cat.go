package cmd

import (
	"errors"
	"fmt"
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

	code := 0

	for _, id := range c.Args {
		r, err := revision.Find(id)

		if err != nil {
			code = 1
			fmt.Fprintf(os.Stderr, "%s: failed to find revision: %s\n", os.Args[0], id)
			continue
		}

		if c.Flags.IsSet("up") {
			r.Direction = revision.Up
		} else if c.Flags.IsSet("down") {
			r.Direction = revision.Down
		}

		fmt.Println(r.Query())
	}

	os.Exit(code)
}
