# mgrt

mgrt is a simple tool for managing revisions across SQL databases. It takes SQL scripts, runs them against the database, and keeps a log of them.

* [Quick Start](#quick-start)
* [Initialization](#initialization)
* [Configuration](#configuration)
* [Performing a Revision](#performing-a-revision)
* [Revision Log](#revision-log)
* [Viewing Revisions](#viewing-revisions)
* [Working with Multiple Databases](#working-with-multiple-databases)

## Quick Start

mgrt can be installed via Go. Simply clone the repository and run `go install`.

```
$ git clone https://github.com/andrewpillar/mgrt.git
$ cd mgrt
$ go install
```

Once installed you can create a new mgrt instance by running `mgrt init`.

```
$ mgrt init
```

Before we can start writing revisions, we need to set the author information in the `mgrt.yml` file.

```yaml
# The type of database, one of:
#   - postgres
#   - mysql
#   - sqlite3
type:

# The database address, if SQLite then the filepath instead.
address:

# Login credentials for the user that will run the revisions.
username:
password:

# Database to run the revisions against, if using SQLite then leave empty.
database:

# Details about the person creating the database revisions.
author:
  name: Andrew Pillar
  email: andrewjohnpillar@gmail.com
```

We can now begin with writing up revisions for mgrt to perform with the `mgrt add` command.

```
$ mgrt add -m "Create users table"
added new revision at:
  revisions/1136214245/up.sql
  revisions/1136214245/down.sql
```

mgrt will create a directory named for the revision's ID, and populate it with some files, a `_message` file containing the message we passed with `-m`, and the SQL files that will contain the up/down logic for the revision.

Writing the revision is as simple as editing the newly created SQL files.

```sql
CREATE TABLE users (
    email    TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
);
```

```sql
DROP TABLE users;
```

We now have a revision. Before we can run it however, we need to configure the database connectivity. This is done in the `mgrt.yml` file. For our purposes we will be running the revisions against an SQLite database, so we only need to configure two properties, `type` and `address`.

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

# Details about the person creating the database revisions.
author:
  name: Andrew Pillar
  email: andrewjohnpillar@gmail.com
```

We can now perform our revisions against the database by running `mgrt run`.

```
$ mgrt run
up - performed revision: 1136214245: Create users table
```

This will read in the contents of the `up.sql` file on each revision we have, and run it against the database. If we try running `mgrt run` again, the revision will not be performed because it was already run once.


```
$ mgrt run
up - already performed revision: 1136214245: Create users table
```

The revision we just performed can be undone by running `mgrt reset`. This will read in the contents of the `down.sql` file on each revision we have.

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
| `address` | The database address formatted as `<host>:<port>`. If using SQLite then a path to the file. |
| `username` | The username of the user performing the revisions. |
| `password` | The password of the user performing the revisions. |
| `database` | The database to perform the revisions on. |
| `author.name` | The name of the person authoring the revisions. |
| `author.email` | The email of the person authoring the revisions. |

>**Note:** When preparing your database for mgrt to run against, ensure you do not have a table with the name of `mgrt_revisions`. This is the table used by mgrt for logging information about the performed revisions. You do not need to worry about creating this table, mgrt will do it for you when a revision is performed for the first time.

## Revisions

mgrt works by performing revisions against the given database. Each time a revision is performed a hash of that revision is stored in the database. This is to ensure that no modifications of that revision cannot be run. Revisions in mgrt are deliberately immutable like this, so as to ensure that a log of all changes made against the database can be kept.

A new revision can be created by running `mgrt add`. This command takes the optional `-m` flag for specifying a message for the revision. If the `-m` flag is not specified, then mgrt will drop you into an editor, as specified via `$EDITOR`, for writing out the revision's message. This file will be pre-populated with your author information

```
Author: Andrew Pillar <andrewjohnpillar@gmail.com>
```

## Performing a Revision

Revisions can be performed in two different ways. When performing a revision with the `mgrt run` command, mgrt will take the SQL code from the `up.sql` file, and run it against the database.

Revisions can then be reset with the `mgrt reset` command. mgrt will take the SQL code from the `down.sql` file, and run it against the database.

Both the `mgrt run`, and `mgrt reset` commands accept a list of revision IDs as their arguments, allowing for finer control over which revisions can be performed.

The typical convention to follow when writing revisions, is to have the `down.sql` directive do the opposite of the `up.sql` directive.

Each time a revision is performed, a checksum is done to ensure that a revision that has been modified cannot be performed again. This is deliberate, as revisions are supposed to be treated as immutable, however sometimes you would want to bypass this check if you want faster iterations on the revisions you are writing. This checksum can be bypassed by passing the `-f` flag to either the `mgrt run`, or `mgrt reset` command.

## Revision Log

Each time a revision is performed a log will be made of that revision. This log is stored in the database itself, and contains the ID of the revision, the direction, the time it happened, and the hash of the revision.

This log can be viewed with the `mgrt log` command.

```
$ mgrt log
Revision: 1136214245 99121b9c2c88efdf77a0da709476e9f57b08d8423fa8af5046c140950ecbc18a
Author:   Andrew Pillar <andrewjohnpillar@gmail.com>
Date:     Mon Jan 02 15:04:05 2006

  Create users table

    DROP TABLE users;

Revision: 1136214245 2d9d97a7e76b07c4636b45a7d3dfaa5a2586c2b0b6734cad4dd05438c96276d9
Author:   Andrew Pillar <andrewjohnpillar@gmail.com>
Date:     Mon Jan 02 15:04:05 2006

  Create users table

    CREATE TABLE users (
        email    TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL
    );
```

Upon being run, mgrt will search the `revisions` directory for the IDs of the revisions that were performed, and display the exact SQL queries that were performed for that revision in the log.

By default, `mgrt log` will display the performed revisions latest first. This can be reversed however by passing the `-r` flag to the command.

## Viewing Revisions

Revisions can be viewed via the `mgrt cat` command. This command must be provided either the `--up`, or `--down` flags, or both in order to view the contents of a revision.

```
$ mgrt cat 1136214245 --up
CREATE TABLE users (
    email    TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
);
$ mgrt cat 1136214245 --down
DROP TABLE users;
```

This can be useful if you want to debug any erroneously written queries you have in a revision. Of course, since they are just plain SQL files on disk you could just `cat` the file normally, this command exists just as a helper.

```
$ mgrt cat 1136214245 --up | mysql ...
```

`mgrt cat` takes a list of revision IDs for its arguments.

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

And to have that revision performed you would pass the same flag to the `run` command.

```
$ mgrt run -c uat
```
