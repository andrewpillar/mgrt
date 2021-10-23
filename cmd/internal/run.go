package internal

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewpillar/mgrt/v3"
)

var RunCmd = &Command{
	Usage: "run <revisions,...>",
	Short: "run the given revisions",
	Long: `Run will perform the given revisions against the given database. The database
to connect to is specified via the -type and -dsn flags, or via the -db flag if a database
connection has been configured via the "mgrt db" command.

The -c flag specifies the category of revisions to run. If not given, then the
default revisions will be run.

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
		typ      string
		dsn      string
		category string
		dbname   string
		verbose  bool
	)

	fs := flag.NewFlagSet(cmd.Argv0+" "+argv0, flag.ExitOnError)
	fs.StringVar(&typ, "type", "", "the database type one of postgresql, sqlite3")
	fs.StringVar(&dsn, "dsn", "", "the dsn for the database to run the revisions against")
	fs.StringVar(&category, "c", "", "the category of revisions to run")
	fs.StringVar(&dbname, "db", "", "the database to connect to")
	fs.BoolVar(&verbose, "v", false, "display information about the revisions performed")
	fs.Parse(args[1:])

	if dbname != "" {
		it, err := getdbitem(dbname)

		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "%s %s: database %s does not exist\n", cmd.Argv0, argv0, dbname)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
			os.Exit(1)
		}

		typ = it.Type
		dsn = it.DSN
	}

	if typ == "" {
		fmt.Fprintf(os.Stderr, "%s %s: database not specified\n", cmd.Argv0, argv0)
		os.Exit(1)
	}

	if dsn == "" {
		fmt.Fprintf(os.Stderr, "%s %s: database not specified\n", cmd.Argv0, argv0)
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
		dir := revisionsDir

		if category != "" {
			dir = filepath.Join(revisionsDir, category)
		}

		ents, err := os.ReadDir(dir)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
			os.Exit(1)
		}

		for _, ent := range ents {
			if ent.IsDir() {
				continue
			}

			rev, err := mgrt.OpenRevision(filepath.Join(dir, ent.Name()))

			if err != nil {
				fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
				os.Exit(1)
			}
			revs = append(revs, rev)
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
