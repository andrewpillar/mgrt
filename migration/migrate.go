package migrate

import (
	"fmt"

	"github.com/andrewpillar/mgrt/database"
	"github.com/andrewpillar/mgrt/revision"
	"github.com/andrewpillar/mgrt/util"
)

func Perform(db database.DB, revisions []*revision.Revision, d revision.Direction, force bool) {
	for _, r := range revisions {

		r.Direction = d

		if err := r.GenHash(); err != nil {
			util.ExitError("failed to perform revision", err)
		}

		if err := db.Perform(r, force); err != nil {
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

		if err := db.Log(r, force); err != nil {
			util.ExitError("failed to log revision", err)
		}

		fmt.Printf("%s - performed revision: %d", d, r.ID)

		if r.Message != "" {
			fmt.Printf(": %s", r.Message)
		}

		fmt.Printf("\n")
	}
}
