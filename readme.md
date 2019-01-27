# mgrt

mgrt is a simple tool for managing revisions across SQL databases. It takes SQL scripts, runs them against the database, and keeps a log of them. mgrt is not intrinsically tied to an ORM, since it only takes SQL, and sends it straight to the database.

* [Quick Start](#quick-start)
* [Initialization](#initialization)
* [Configuration](#configuration)
* [Revisions](#revision)
  * [Directives](#directives)
  * [Performing a Revision](#performing-a-revision)
  * [Revision Log](#revision-log)
* [Working with Multiple Databases](#working-with-multiple-databases)

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

## Initialization

A new mgrt instance can be initialized with `mgrt init`. This command takes on optional argument for the directory to initialize mgrt in. If that directory does not exist then it will be created.

Once initialized there will be a directory, and a configuration file within the directory. The `revisions` directory stores the revisions, and the `mgrt.yml` configuration file contains information about the database you want to perform revisions against.

## Configuration

Configuring mgrt is simple. Depending on the database type you're running against, will depend on the configuration options that need to be set in the `mgrt.yml` file.

| Property | Purpose |
|----------|---------|
| `type` | The type of database the revisions will be performed against. |
| `address` | The database address formatted as `<host>:<port>`. If using SQLite then the path to the file. |
| `username` | The username of the user performing the revisions. |
| `password` | The password of the user performing the revisions. |
| `database` | The database to perform the revisions on. |

>**Note:** When preparing your database for mgrt to run against, ensure you do not have a table with the name of `mgrt_revisions`. This is the table used by mgrt for logging information about the performed revisions. You do not need to worry about creating this table, mgrt will do it for you when a revision is performed for the first time.

## Revisions

mgrt works by performing revisions against the given database. Each time a revision is performed a hash of that revision is stored in the database. This is to ensure that no modifications of that revision cannot be run. Revisions in mgrt are deliberately immutable like this, so as to ensure that a log of all changes made against the database can be kept.

A new revision can be created by running `mgrt add`. This command does take the optional `-m` flag for specifying a message for the revision. mgrt will then open up the newly created revision with the editor you have set in the `$EDITOR` environment variable.

```sql
-- mgrt: revision: 1136214245: Create users table
-- mgrt: up

-- mgrt: down


```

Above is what a new revision would look like. It is an SQL file with some comments contained information about the revision. To mgrt these comments are known as [directives](#directives), and tell mgrt how to interpret the revision. Each revision that is created will have an ID set to the UNIX timestamp of the time the revision was created. If a message was provided to the `mgrt add` command via the `-m` flag then the message will be in the file too.

### Directives

Directives in mgrt are delineated with an SQL comment that follows the below format:

```sql
--- mgrt: [directive]: [args...]
```

Right now, there are three different directives that can be used in mgrt, `revision`, `up`, and `down`.

The `revision` directive tells mgrt about the revision itself. This will hold the ID of the revision, and the revision message if there is one.

The `up`, and `down` directives tell mgrt what SQL code should be performed for a revision depending on whether a revision is being run, or reset respectively.

### Performing a Revision

Revisions can be performed in two different ways. When performing a revision with the `mgrt run` command, mgrt will take the SQL code from the `-- mgrt: up` directive, and run it against the database.

Revisions can then be reset with the `mgrt reset` command. mgrt will take the SQL code from the `-- mgrt: down` directive, and run it against the database.

The typical convention to follow when writing revisions, is to have the `-- mgrt: down` directive do the opposite of the `-- mgrt: up` directive.

### Revision Log

Each time a revision is performed a log will be made of that revision. This log is stored in the database itself, and contains the ID of the revision, the direction, the time it happened, and the hash of the revision.

This log can be viewed with the `mgrt log` command.

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

Upon being run, mgrt will search the `revisions/` directory for the IDs of the revisions that were performed, and display the exact SQL queries that were performed for that revision in the log.


## Working with Multiple Databases

mgrt allows the ability to perform revisions against multiple databases. The `-c` flag can be passed to the `add`, `run`, `reset`, and `log` commands. This tells mgrt which config directory to look into when performing these commands. So, you could initialize mgrt multiple times for each database you might have, like so:

```
$ mgrt init uat
$ mgrt init prod
```

Then to add a revision to a specific database you would pass the `-c` flag to the `add` command.

```
$ mgrt add -m "Create users table" -c uat
```

And to have that revisions performed you would also pass the same flag to `run`.

```
$ mgrt run -c uat
```
