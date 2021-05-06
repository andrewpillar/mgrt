package internal

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/andrewpillar/mgrt"
)

var LogCmd = &Command{
	Usage: "log",
	Short: "log the performed revisions",
	Long:  `Log displays all of the revisions that have been performed in the given
database. The database to connect to is specified via the -type and -dsn flags.

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
	Run: logCmd,
}

func logCmd(cmd *Command, args []string) {
	argv0 := args[0]

	var (
		typ string
		dsn string
	)

	fs := flag.NewFlagSet(cmd.Argv0+" "+argv0, flag.ExitOnError)
	fs.StringVar(&typ, "type", "", "the database type one of postgresql, sqlite3")
	fs.StringVar(&dsn, "dsn", "", "the dsn for the database to run the revisions against")
	fs.Parse(args[1:])

	db, err := openDB(typ, dsn)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.Argv0, argv0, err)
		os.Exit(1)
	}

	defer db.Close()

	revs, err := mgrt.GetRevisions(db)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: failed to get revisions: %s\n", cmd.Argv0, argv0, err)
		os.Exit(1)
	}

	for _, rev := range revs {
		fmt.Println("revision", rev.ID)
		fmt.Println("Author:    ", rev.Author)
		fmt.Println("Performed: ", rev.PerformedAt.Format(time.ANSIC))
		fmt.Println()

		lines := strings.Split(rev.SQL, "\n")

		for _, line := range lines {
			fmt.Println("   ", line)
		}
	}
}
