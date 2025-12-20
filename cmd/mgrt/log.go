package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/andrewpillar/mgrt/v4"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

var LogCmd = &Command{
	Usage: "log <source>",
	Short: "log the revisions performed on the database",
	Long: `log will display all of the revisions performed agains the database in order
of most recent.

The given source should be either a valid path or URL. If a URL, then the scheme
should denote the database backend being used, either postgresql:// or
sqlite://. If no scheme is provided then SQLite is used as the backend.`,
	Run: logCmd,
}

func logCmd(cmd *Command, args []string) error {
	if len(args) == 0 {
		return ErrUsage
	}

	source := args[0]

	url, err := url.Parse(source)

	if err != nil {
		return err
	}

	var driver string

	switch url.Scheme {
	case "":
		fallthrough
	case "sqlite":
		driver = "sqlite"
	case "postgresql":
		driver = "pgx"
	default:
		return errors.New("unsupported database")
	}

	db, err := sql.Open(driver, url.String())

	if err != nil {
		return err
	}

	ctx := context.Background()

	revs, err := mgrt.Log(ctx, db)

	if err != nil {
		return err
	}

	for i, rev := range revs {
		fmt.Printf("%-4s %s\n", rev.Direction, rev.Ref)
		fmt.Println("Revision:    ", rev.Name)
		fmt.Println("Performed at:", rev.PerformedAt.Format(time.ANSIC))

		if rev.Comment.Valid {
			fmt.Println(strings.TrimSpace(rev.Comment.V))
		}

		fmt.Println()

		lines := strings.Split(strings.TrimSpace(rev.SQL()), "\n")

		for _, line := range lines {
			fmt.Println("   ", line)
		}

		if i != len(revs)-1 {
			fmt.Println()
		}
	}
	return nil
}
