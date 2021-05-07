# mgrt

mgrt is a simple tool for managing revisions across SQL databases. It takes SQL
scripts, runs them against the database, and keeps a log of them.

* [Quick start](#quick-start)
* [Database connection](#database-connection)
* [Revisions](#revisions)
* [Revision log](#revision-log)
* [Viewing revisions](#viewing-revisions)
* [Library usage](#library-usage)

## Quick start

To install mgrt, clone the repository and run the `./make.sh` script,

    $ git clone https://github.com/andrewpillar/mgrt
    $ cd mgrt
    $ ./make.sh

to build mgrt with SQLite3 support add `sqlite3` to the `TAGS` environment
variable,

    $ TAGS="sqlite3" ./make.sh

this will produce a binary at `bin/mgrt`, add this to your `PATH`.

Once installed you can start using mgrt right away, there is nothing to
initialize. To begin writing revisions simply invoke `mgrt add`,

    $ mgrt add "My first revision"

this will create a new revision file in the `migrations` directory, and open
it up for editting with the revision to write,

    /*
    Revision: 20060102150405
    Author:   Andrew Pillar <me@andrewpillar.com>
    
    My first revision
    */

    CREATE TABLE users (
        id INT NOT NULL UNIQUE
    );

once you've saved the revision and quit the editor, you will see the revision ID
printed out,

    $ mgrt add "My first revision"
    revision created 20060102150405

local revisions can be viewed with `mgrt ls`. This will display the ID, the
author of the revision, and its comment, if any,

    $ mgrt ls
    20060102150405: Andrew Pillar <me@andrewpillar.com> - My first revision

revisions can be applied to the database via `mgrt run`. This command takes two
flags, `-type` and `-dsn` to specify the type of database to run the revision
against, and the data source for that database. Let's run our revision against
an SQLite3 database,

    $ mgrt run -type sqlite3 -dsn acme.db

revisions can only be performed on a database once, and cannot be undone. We can
view the revisions that have been run against the database with `mgrt log`. Just
like `mgrt run`, we use the `-type` and `-dsn` flags to specify the database to
connect to,

    $ mgrt log -type sqlite3 -dsn acme.db
    revision 20060102150405
    Author:    Andrew Pillar <me@andrewpillar.com>
    Performed: Mon Jan  6 15:04:05 2006
    My first revision
    
        CREATE TABLE users (
            id INT NOT NULL UNIQUE
        );

this will list out the revisions that have been performed, along with the SQL
code that was executed as part of that revision.

mgrt also offers the ability to sync the revisions that have been performed on
a database against what you have locally. This is achieved with `mgrt sync`, and
just like before, this also takes the `-type` and `-dsn` flags. Lets delete the
`migrations` directory that was created for us and do a `mgrt sync`.

    $ rm -rf migrations
    $ mgrt ls
    $ mgrt sync -type sqlite3 -dsn acme.db
    $ mgrt ls
    20060102150405: Andrew Pillar <me@andrewpillar.com> - My first revision

with `mgrt sync` you can easily view the migrations that have been run against
different databases.

## Database connection

Database connection for mgrt is controlled via the `-type` and `-dsn` flags.
This allows for easy use of mgrt with multiple databases of different types.

The `-type` flag specifies the type of database to connect to, it will be one
of,

* mysql
* postgresql
* sqlite3

The `-dsn` flag specifies the data source name for the database. This will vary
depending on the type of database you're connecting to.

mysql and postgresql both allow for the URI connection string, such as,

    type://[user[:password]@][host]:[port][,...][/dbname][?param1=value1&...]

where type would either be mysql or postgresql. The postgresql type also allows
for the DSN string such as,

    host=localhost port=5432 dbname=mydb connect_timeout=10

sqlite3 however will accept a filepath, or the `:memory:` string, for example,

    -dsn :memory:

## Revisions

Revisions are SQL scripts that are performed against the given database. Each
revision can only be performed once, and cannot be undone. If you wish to undo
a revision, then it is recommended to write another revision that does the
inverse of the prior.

Revisions are stored in the `migrations` directory from where the `mgrt add`
command was run. Each revision file is prefixed with a comment block header
that contains metadata about the revision itself, such as the ID, the author and
a short comment about the revision.

## Revision log

Each time a revision is performed, a log will be made of that revision. This log
is stored in the database, in the `mgrt_revisions` table. This will contain the
ID, the author, the comment (if any), and the SQL code itself, along with the
time of execution.

The revisions performed against a database can be viewed with `mgrt log`,

    $ mgrt log -type sqlite3 -dsn acme.db
    revision 20060102150405
    Author:    Andrew Pillar <me@andrewpillar.com>
    Performed: Mon Jan  6 15:04:05 2006

        My first revision

## Viewing revisions

Local revisions can be viewed with `mgrt cat`. This simply takes a list of
revision IDs to view.

    $ mgrt cat 20060102150405
    /*
    Revision: 20060102150405
    Author:   Andrew Pillar <me@andrewpillar.com>
    
    My first revision
    */
    
    CREATE TABLE users (
            id INT NOT NULL UNIQUE
    );

The `-sql` flag can be passed to the command too to only display the SQL portion
of the revision,

    $ mgrt cat -sql 20060102150405
    CREATE TABLE users (
            id INT NOT NULL UNIQUE
    );

performed revisions can also be seen with `mgrt show`. You can pass a revision
ID to `mgrt show` to view an individual revision. If no revision ID is given,
then the latest revision is shown.

    $ mgrt show -type sqlite3 -dsn acme.db 20060102150405
    revision 20060102150405
    Author:    Andrew Pillar <me@andrewpillar.com>
    Performed: Mon Jan  6 15:04:05 2006

        My first revision

        CREATE TABLE users (
                id INT NOT NULL UNIQUE
        );

## Library usage

As well as a CLI application, mgrt can be used as a library should you want to
be able to have revisions performed directly in your application. To start using
it just import the repository into your code,

    import "github.com/andrewpillar/mgrt"

from here you will be able to start creating revisions and performing them
against any pre-existing database connection you may have,

    // mgrt.Open will wrap sql.Open from the stdlib, and initialize the database
    // for performing revisions.
    db, err := mgrt.Open("sqlite3", "acme.db")

    if err != nil {
        panic(err) // maybe acceptable here
    }

    rev := mgrt.NewRevision("Andrew", "This is being done from Go.")

    if err := rev.Perform(db); err != nil {
        if !errors.Is(err, mgrt.ErrPerformed) {
            panic(err) // not best practice
        }
    }

all pre-existing revisions can be retrieved via GetRevisions,

    revs, err := mgrt.GetRevisions(db)

    if err != nil {
        panic(err) // don't actually do this
    }

more information about using mgrt as a library can be found in the
[Go doc](https://pkg.go.dev/github.com/andrewpillar/mgrt) itself for mgrt.
