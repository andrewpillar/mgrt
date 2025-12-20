package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"

	"github.com/andrewpillar/mgrt/v4"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

func performRevisions(source, dir string, d mgrt.Direction, verbose bool) error {
	url, err := url.Parse(source)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	defer db.Close()

	revs, err := mgrt.LoadDir(os.DirFS("."), dir)

	if err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	done := make(chan *mgrt.Revision)
	errs := make(chan error)

	go func() {
		if err := mgrt.Perform(ctx, db, done, d, revs...); err != nil {
			errs <- err
		}
		close(done)
	}()

	performed := 0

loop:
	for {
		select {
		case <-ch:
			cancel()
			break loop
		case rev, ok := <-done:
			if !ok {
				break loop
			}

			performed++

			if verbose {
				fmt.Printf("%-4s %s %s\n", rev.Direction, rev.Ref[:9], rev.Name)
			}
		case err := <-errs:
			cancel()
			return err
		}
	}

	if verbose {
		if performed == 0 {
			fmt.Println("no new revisions to perform")
		}
	}
	return nil
}

var UpCmd = &Command{
	Usage: "up [-v] <source> <revisions>",
	Short: "perform the revisions annotated as +up",
	Long: `up will perform the SQL files in the given location, only executing the SQL code
that immediately follows after an "+up" annotation, for example,

    $ cat table.sql
    -- +up
    
    CREATE TABLE t (
        id VARCHAR UNIQUE NOT NULL,
        PRIMARY KEY (id)
    );
    
    -- +down
    
    DROP TABLE t;

the above CREATE statement is the only thing that would be executed from running
"mgrt up".

The revisions are expected to be in a directory named for the database backend
being used. For example, if using sqlite, then the revisions should be in a
directory named sqlite.

If the revision has already been performed, then it will not be performed again.

The given source should be either a valid path or URL. If a URL, then the scheme
should denote the database backend being used, either postgresql:// or
sqlite://. If no scheme is provided then SQLite is used as the backend.

Pass -v to display the revisions that are performed, if any.`,
	Run: performCmd(mgrt.Up),
}

var DownCmd = &Command{
	Usage: "down [-v] <source> <revisions>",
	Short: "perform the revisions annotated as +down",
	Long: `down will perform the SQL files in the given location, only executing the SQL
code that immediately follows after a "+down" annotation, for example,

    $ cat table.sql
    -- +up
    
    CREATE TABLE t (
        id VARCHAR UNIQUE NOT NULL,
        PRIMARY KEY (id)
    );
    
    -- +down
    
    DROP TABLE t;

the above DROP statement is the only thing that would be executed from running
"mgrt down".

The revisions are expected to be in a directory named for the database backend
being used. For example, if using sqlite, then the revisions should be in a
directory named sqlite.

If the revision has already been performed, then it will not be performed again.

The given source should be either a valid path or URL. If a URL, then the scheme
should denote the database backend being used, either postgresql:// or
sqlite://. If no scheme is provided then SQLite is used as the backend.

Pass -v to display the revisions that are performed, if any.`,
	Run: performCmd(mgrt.Down),
}

func performCmd(dir mgrt.Direction) func(*Command, []string) error {
	return func(cmd *Command, args []string) error {
		var verbose bool

		fs := flag.NewFlagSet(cmd.Argv0, flag.ExitOnError)
		fs.BoolVar(&verbose, "v", false, "display revisions that are performed")
		fs.Parse(args)

		args = fs.Args()

		if len(args) != 2 {
			return ErrUsage
		}

		if err := performRevisions(args[0], args[1], dir, verbose); err != nil {
			return err
		}
		return nil
	}
}
