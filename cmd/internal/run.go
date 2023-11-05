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

		fmt.Fprintf(os.Stderr, "%s: failed to run revisions: %s\n", cmd.Argv0, err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "%s: %s is not a directory\n", cmd.Argv0, revisionsDir)
		os.Exit(1)
	}

	var (
		typ      string
		dsn      string
		category string
		dbname   string
		verbose  bool
	)

	fs := flag.NewFlagSet(cmd.Argv0, flag.ExitOnError)
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
				fmt.Fprintf(os.Stderr, "%s: database %s does not exist\n", cmd.Argv0, dbname)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "%s: %s\n", cmd.Argv0, err)
			os.Exit(1)
		}

		typ = it.Type
		dsn = it.DSN
	}

	if typ == "" {
		fmt.Fprintf(os.Stderr, "%s: database not specified\n", cmd.Argv0)
		os.Exit(1)
	}

	if dsn == "" {
		fmt.Fprintf(os.Stderr, "%s: database not specified\n", cmd.Argv0)
		os.Exit(1)
	}

	revs := make([]*mgrt.Revision, 0)

	for _, id := range fs.Args() {
		rev, err := mgrt.OpenRevision(revisionPath(id))

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: failed to open revision %s: %s\n", cmd.Argv0, id, err)
			os.Exit(1)
		}
		revs = append(revs, rev)
	}

	if len(revs) == 0 {
		dir := revisionsDir

		if category != "" {
			dir = filepath.Join(revisionsDir, category)
		}

		revs, err = mgrt.LoadRevisions(dir)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", cmd.Argv0, err)
			os.Exit(1)
		}
	}

	db, err := mgrt.Open(typ, dsn)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmd.Argv0, err)
		os.Exit(1)
	}

	defer db.Close()

	var c mgrt.Collection

	for _, rev := range revs {
		c.Put(rev)
	}

	code := 0

	for _, rev := range c.Slice() {
		if err := rev.Perform(db); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			code = 1
			continue
		}

		if verbose {
			fmt.Println(rev.ID, rev.Title())
		}
	}

	if code != 0 {
		os.Exit(code)
	}
}
