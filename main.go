package main

import (
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/mgrt/cmd"
	"github.com/andrewpillar/mgrt/util"
)

func main() {
	c := cli.New()

	c.Command("init", cmd.Init)

	addCmd := c.Command("add", cmd.Add)

	addCmd.AddFlag(&cli.Flag{
		Name:     "message",
		Short:    "-m",
		Long:     "--message",
		Argument: true,
	})

	c.Command("run", cmd.Run)
	c.Command("reset", cmd.Reset)
	c.Command("log", cmd.Log)

	if err := c.Run(os.Args[1:]); err != nil {
		util.ExitError("", err)
	}
}
