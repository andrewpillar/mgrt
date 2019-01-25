package main

import (
	"fmt"
	"os"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/mgrt/cmd"
	"github.com/andrewpillar/mgrt/usage"
	"github.com/andrewpillar/mgrt/util"
)

func usageHandler(c cli.Command) {
	if c.Name == "" {
		fmt.Println(usage.Main)
		return
	}

	fmt.Println(usage.Commands[c.FullName()])
}

func main() {
	c := cli.New()

	c.AddFlag(&cli.Flag{
		Name:      "help",
		Long:      "--help",
		Exclusive: true,
		Handler:   func(f cli.Flag, c cli.Command) {
			usageHandler(c)
		},
	})

	configFlag := &cli.Flag{
		Name:     "config",
		Short:    "-c",
		Long:     "--config",
		Argument: true,
		Default:  ".",
	}

	c.NilHandler(usageHandler)
	c.Main(nil)

	c.Command("init", cmd.Init)

	addCmd := c.Command("add", cmd.Add)

	addCmd.AddFlag(&cli.Flag{
		Name:     "message",
		Short:    "-m",
		Long:     "--message",
		Argument: true,
	})

	addCmd.AddFlag(configFlag)

	c.Command("run", cmd.Run).AddFlag(configFlag)
	c.Command("reset", cmd.Reset).AddFlag(configFlag)
	c.Command("log", cmd.Log).AddFlag(configFlag)

	lsCmd := c.Command("ls", cmd.Ls)

	lsCmd.AddFlag(configFlag)

	lsCmd.AddFlag(&cli.Flag{
		Name:  "reverse",
		Short: "-r",
		Long:  "--reverse",
	})

	if err := c.Run(os.Args[1:]); err != nil {
		util.ExitError("", err)
	}
}
