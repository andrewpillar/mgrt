# mgrt

mgrt is a simple tool for managing revisions across SQL databases. It takes SQL scripts, runs them against the database, and keeps a log of them. mgrt is not intrinsically tied to an ORM, since it only takes SQL, and sends it straight to the database.

* [Quick Start](#quick-start)

## Quick Start

If you have Go installed then you can simply clone this repository and run `go install`, assuming you have `~/go/bin` in your `$PATH`.

```
$ git clone https://github.com/andrewpillar/mgrt.git
$ cd mgrt
$ go install
```

Once installed you can create a new mgrt instance by running `mgrt init`.

```
$ mgrt init
```

We can now begin writing up revisions for mgrt to perform with the `mgrt add`. command.

```
$ mgrt add -m "Create users table"
```

mgrt will then drop you into an editor, as specified via `$EDITOR`, for editing the newly created revision. This file will be pre-populated with some directives which are interpreted by mgrt to determine how the revision should be run.

```sql
-- mgrt: revision: 1136214245: Create users table
-- mgrt: up

-- mgrt: down

```

To write our revision, we simply put the SQL code we want to be run beneath the necessary directive. For our current revision, we want to create users table.

```sql
-- mgrt: revision: 1136214245: Create users table
-- mgrt: up

CREATE TABLE users (
    email    TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
);

-- mgrt: down

DROP TABLE users;

```

As you can see, we write the creation logic beneath the `-- mgrt: up` directive, and the destruction logic beneath the `-- mgrt: down` directive.

We now have a revision. Before we can run it however, we need to configure the database connectivity. This is done in the `mgrt.yml` file. For our purposes we will be running the revisions against an SQLite database, so we only need to configure two properties, `type`, and `address`.

```yaml
# The type of database, one of:
#   - postgres
#   - mysql
#   - sqlite3
type: sqlite3

# The database address, if SQLite then the filepath instead.
address: db.sqlite

# Login credentials for the user that will run the revisions.
username:
password:

# Database to run the revisions against, if using SQLite then leave empty.
database:
```

We can now perform our revisions against the database by running `mgrt run`.

```
$ mgrt run
up - performed revision: 1136214245: Create users table

```

This will perform the `-- mgrt: up` directive on each revision we have. If we try running `mgrt run` again, the revision will not be performed because it already once has.

```
$ mgrt run
up - already performed revision: 1136214245
```

The revision we just performed can be undone by running `mgrt reset`. This will perform the `-- mgrt: down` directive on each revision.

```
$ mgrt reset
down - performed revision: 1136214245: Create users table
```

Each revision that has been performed will be logged in the database. This log can be read with `mgrt log`.

```
$ mgrt log
Revision: 1136214245 - Create users table
Performed At: Mon Jan 02 15:04:05 2006

DROP TABLE users;

Revision: 1136214245 - Create users table
Performed At: Mon Jan 02 15:04:05 2006

CREATE TABLE users (
    email    TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
);

```
