package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/andrewpillar/cli"

	"github.com/andrewpillar/mgrt/config"
	"github.com/andrewpillar/mgrt/database"
	"github.com/andrewpillar/mgrt/revision"
	"github.com/andrewpillar/mgrt/util"
)

func Log(c cli.Command) {
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

	var revisions []*revision.Revision

	if c.Flags.IsSet("reverse") {
		revisions, err = db.ReadLogReverse(c.Args...)
	} else {
		revisions, err = db.ReadLog(c.Args...)
	}

	if err != nil {
		util.ExitError("failed to read revisions log", err)
	}

	for _, r := range revisions {

		hash := make([]byte, len(r.Hash), len(r.Hash))

		for i := range hash {
			hash[i] = r.Hash[i]
		}

		fmt.Printf("Revision: %d: %x", r.ID, hash)

		if r.Forced {
			fmt.Printf(" [FORCED]")
		}

		fmt.Printf("\nAuthor:   %s\n", r.Author)
		fmt.Printf("Date:     %s\n", r.CreatedAt.Format("Mon Jan 02 15:04:05 2006"))

		if r.Message != "" {
			subject, body := r.SplitMessage()

			fmt.Printf("\n  %s\n", subject)

			s := bufio.NewScanner(strings.NewReader(body))

			for s.Scan() {
				fmt.Printf("  %s\n", s.Text())
			}
		}

		s := bufio.NewScanner(strings.NewReader(r.Query()))

		fmt.Printf("\n")

		for s.Scan() {
			fmt.Printf("    %s\n", s.Text())
		}

		fmt.Printf("\n")
	}
}
