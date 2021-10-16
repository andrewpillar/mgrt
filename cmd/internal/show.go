package internal

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/andrewpillar/mgrt/v3"
)

var ShowCmd = &Command{
	Usage: "show [revision]",
	Short: "show the given revision",
	Long:  `Show will show the SQL that was run in the given revision. If no revision is
specified, then the latest revision will be shown, if any. The database to connect to is
specified via the -type and -dsn flags, or via the -db flag if a database connection has
been configured via the "mgrt db" command.

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
	Run: showCmd,
}

func showCmd(cmd *Command, args []string) {
	argv0 := args[0]

	var (
		typ    string
		dsn    string
		dbname string
	)

	fs := flag.NewFlagSet(cmd.Argv0+" "+argv0, flag.ExitOnError)
	fs.StringVar(&typ, "type", "", "the database type one of postgresql, sqlite3")
	fs.StringVar(&dsn, "dsn", "", "the dsn for the database to run the revisions against")
	fs.StringVar(&dbname, "db", "", "the database to connect to")
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

	db, err := mgrt.Open(typ, dsn)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
		os.Exit(1)
	}

	defer db.Close()

	args = fs.Args()

	var rev *mgrt.Revision

	if len(args) >= 1 {
		rev, err = mgrt.GetRevision(db, args[0])

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: failed to show revision: %s\n", cmd.Argv0, argv0, err)
			os.Exit(1)
		}
	}

	if rev == nil {
		revs, err := mgrt.GetRevisions(db, 1)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %s: failed to show revision: %s\n", cmd.Argv0, argv0, err)
			os.Exit(1)
		}
		rev = revs[0]
	}

	fmt.Println("revision", rev.ID)
	fmt.Println("Author:    ", rev.Author)
	fmt.Println("Performed: ", rev.PerformedAt.Format(time.ANSIC))
	fmt.Println()

	lines := strings.Split(rev.Comment, "\n")

	for _, line := range lines {
		fmt.Println("   ", line)
	}
	fmt.Println()

	lines = strings.Split(rev.SQL, "\n")

	for _, line := range lines {
		fmt.Println("   ", line)
	}
	fmt.Println()
}
