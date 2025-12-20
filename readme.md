# mgrt

mgrt is a simple tool for managing revisions across SQL databases. It takes SQL
scripts, runs them against the database, and keeps a log of them.

* [Quick start](#quick-start)
* [Library usage](#library-usage)

## Quick start

To install mgrt, clone the repository and run `make install`,

```shell
$ git clone https://github.com/andrewpillar/mgrt
$ cd mgrt/
$ make install
```

Once installed, you can start using mgrt right away, there is nothing to
initialize. To begin writing revisions simply run the `mgrt add` command,

```shell
$ mgrt add "My first revision"
```

this will create a new revision file in the `revisions`directory and open it up
for editting,

```
/*
 * My first revision
 */
-- +up

-- +down
```

the file will be pre-populated with the comment given to the `add` command and
the `-- +up` and `-- +down` annotations which delineate which SQL code should be
performed for the `up` and `down` commands.

Write some SQL code that will create a table for both the up and down
annotations,

```sql
/*
 * My first revision
 */
-- +up

CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY
);

-- +down

DROP TABLE IF EXISTS users;
```

once the revision has been saved the revision name will be printed out,

```shell
$ mgrt add "My first revision"
revision created revisions/2006-01-02T15-04-05-My-first-revision.sql
```

This can now be performed against a database via the `up` command. This
command takes a single argument which is the DSN to the database.

```shell
$ mgrt up -v example-db.sqlite revisions/
up   3718effeb revisions/2006-01-02T15-04-05-My-first-revision.sql
```

The first argument the `up` command takes is the DSN to the database. This takes
the format of a URL, and the scheme can be used to specify the type of database
to connect to, which can either be `sqlite` or `postgresql`. If no scheme is
given, then `sqlite` is used.

For example, to perform a revision against a PostgreSQL database you would run,

```shell
$ mgrt up "postgresql://user:password@db.example.com:5432/dbname" revisions/
```

If the `up` command is performed twice on the same set of revisions already
performed then no further changes are made.

```shell
$ mgrt up -v example-db.sqlite revisions/
no new revisions to perform
```

If we want to perform the revisions again, we can perform them via the `down`
command which will execute the SQL code for the `-- +down` annotation,

```
$ mgrt down -v example-db.sqlite revisions/
down 3718effeb revisions/2006-01-02T15-04-05-My-first-revision.sql
```

Now that some revisions have been performed, they can be viewed via the `log`
command. The `log` command will display the revisions that have been performed
in order of most recent,

```
$ mgrt log example-db.sqlite
down 3718effeb2d7350c862b86fe56c000b3f91b92c095476580f0d0ec66c68173bd
Revision:     revisions/2006-01-02T15-04-05-My-first-revision.sql
Performed at: Mon Jan 6 15:04:05 2006
My first revision

    DROP TABLE IF EXISTS users;

up   3718effeb2d7350c862b86fe56c000b3f91b92c095476580f0d0ec66c68173bd
Revision:     revisions/2006-01-02T15-04-05-My-first-revision.sql
Performed at: Mon Jan 6 15:04:05 2006
My first revision

    CREATE TABLE IF NOT EXISTS users (
        id INT PRIMARY KEY
    );

```

## Library usage

mgrt can be used as a library for have revisions performed in code. To do this,
first import the library,

```go
import (
    "github.com/andrewpillar/mgrt/v4"
)
```

Next, load in the revisions. This can be done via the [mgrt.Load][] function,
which takes an [fs.FS][] interface and the path to load from.

[mgrt.Load]: https://pkg.go.dev/github.com/andrewpillar/mgrt#Load
[fs.FS]: https://pkg.go.dev/io/fs/fs#FS

```go
revs, err := mgrt.Load(os.DirFS("revisions/"), ".")

if err != nil {
    // Handle error.
}
```

Because it makes use of the [fs.FS][] interface, this means that revisions can
be embedded directly into the code itself via [embed][].

[embed]: https://pkg.go.dev/embed

With the revisions now loaded, the [mgrt.Perform][] function can be called to
peform them. This takes a database connection from [database/sql][] to perform
the revisions against,

[mgrt.Perform]: https://pkg.go.dev/github.com/andrewpillar/mgrt#Perform
[database/sql]: https://pkg.go.dev/database/sql

```go
db, err := sql.Open(driver, dsn)

if err != nil {
    // Handle error.
}

defer db.Close()

ctx := context.Background()

if err := mgrt.Perform(ctx, db, nil, mgrt.Up, revs...); err != nil {
    // Handle error.
}
```

The [mgrt.Perform][] function takes a channel. Each revisions which is performed
will be sent down this channel, if given to the function. This can be used to
provide reporting on which revisions have been performed,

```go
done := make(chan *mgrt.Revision)

go func() {
    defer close(done)

    if err := mgrt.Perform(ctx, db, done, mgrt.Up, revs...); err != nil {
        // Handle error.
    }
}()

for rev := range done {
    fmt.Printf("%-4s %s %s\n", rev.Direction, rev.Ref[:9], rev.Name)
}
```
