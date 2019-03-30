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

	c.AddFlag(&cli.Flag{
		Name:     "config",
		Short:    "-c",
		Long:     "--config",
		Argument: true,
		Default:  ".",
	})

	reverseFlag := &cli.Flag{
		Name:  "reverse",
		Short: "-r",
		Long:  "--reverse",
	}

	forceFlag := &cli.Flag{
		Name:  "force",
		Short: "-f",
		Long:  "--force",
	}

	c.MainCommand(usageHandler)

	c.Command("init", cmd.Init)

	addCmd := c.Command("add", cmd.Add)

	addCmd.AddFlag(&cli.Flag{
		Name:     "message",
		Short:    "-m",
		Long:     "--message",
		Argument: true,
	})

	catCmd := c.Command("cat", cmd.Cat)

	catCmd.AddFlag(&cli.Flag{
		Name: "up",
		Long: "--up",
	})

	catCmd.AddFlag(&cli.Flag{
		Name: "down",
		Long: "--down",
	})

	c.Command("run", cmd.Run).AddFlag(forceFlag)
	c.Command("reset", cmd.Reset).AddFlag(forceFlag)

	logCmd := c.Command("log", cmd.Log)

	logCmd.AddFlag(reverseFlag)

	lsCmd := c.Command("ls", cmd.Ls)

	lsCmd.AddFlag(reverseFlag)

	if err := c.Run(os.Args[1:]); err != nil {
		util.ExitError("", err)
	}
}
