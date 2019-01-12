package cmd

import (
	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/mgrt/revision"
)

func Reset(c cli.Command) {
	perform(c, revision.Down)
}
