package internal

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewpillar/mgrt"
)

var RunCmd = &Command{
	Usage: "run <revisions,...>",
	Short: "run the given revisions",
	Long: `Run will perform the given revisions against the given database. The database
to connect to is specified via the -type and -dsn flags.

The -type flag specifies the type of database to connect to, it will be one of,

    mysql
    postgresql
    sqlite3

The -dsn flag specifies the data source name for the database. This will vary
depending on the type of database you're connecting to.

mysql and postgresql both allow for the URI connection string, such as,

    type://[user[:password]@][host]:[port][,...][/dbname][?param1=value1&...]

where type would either be mysql or postgresql. The postgresql type also allows
for the DSN string such as,

    host=localhost port=5432 dbname=mydb connect_timeout=10

sqlite3 however will accept a filepath, or the :memory: string, for example,

    -dsn :memory:`,
	Run: runCmd,
}

func runCmd(cmd *Command, args []string) {
	info, err := os.Stat(revisionsDir)

	if err != nil {
		if os.IsNotExist(err) {
			return
		}

		fmt.Fprintf(os.Stderr, "%s %s: failed to run revisions: %s\n", cmd.Argv0, args[0], err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "%s %s: %s is not a directory\n", cmd.Argv0, args[0], revisionsDir)
		os.Exit(1)
	}

	argv0 := args[0]

	var (
		typ     string
		dsn     string
		verbose bool
	)

	fs := flag.NewFlagSet(cmd.Argv0+" "+argv0, flag.ExitOnError)
	fs.StringVar(&typ, "type", "", "the database type one of postgresql, sqlite3")
	fs.StringVar(&dsn, "dsn", "", "the dsn for the database to run the revisions against")
	fs.BoolVar(&verbose, "v", false, "display information about the revisions performed")
	fs.Parse(args[1:])

	if typ == "" {
		fmt.Fprintf(os.Stderr, "%s %s: missing -type flag\n", cmd.Argv0, argv0)
		os.Exit(1)
	}

	if dsn == "" {
		fmt.Fprintf(os.Stderr, "%s %s: missing -dsn flag\n", cmd.Argv0, argv0)
		os.Exit(1)
	}

	revs := make([]*mgrt.Revision, 0)

	for _, id := range fs.Args() {
		rev, err := mgrt.OpenRevision(revisionPath(id))

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: failed to open revision %s: %s\n", cmd.Argv0, argv0, id, err)
			os.Exit(1)
		}
		revs = append(revs, rev)
	}

	if len(revs) == 0 {
		err := filepath.Walk(revisionsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			rev, err := mgrt.OpenRevision(path)

			if err != nil {
				return err
			}

			revs = append(revs, rev)
			return nil
		})

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
			os.Exit(1)
		}
	}

	db, err := mgrt.Open(typ, dsn)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
		os.Exit(1)
	}

	defer db.Close()

	if err := mgrt.PerformRevisions(db, revs...); err != nil {
		if _, ok := err.(mgrt.Errors); ok {
			if verbose {
				fmt.Fprintf(os.Stderr, "%s", err)
			}
			return
		}

		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
		os.Exit(1)
	}
}
