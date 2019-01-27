#!/bin/sh

set -ex

DIR=$(mktemp -d)

now=$(date +"%s")

revision="-- mgrt: revision: $now: Create users table
-- mgrt: up

CREATE TABLE users (
	email    TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL
);

-- mgrt: down

DROP TABLE users;
"

cd "$DIR"

mgrt init

sed -i "s/type:/type: sqlite3/g" mgrt.yml
sed -i "s/address:/address: db.sqlite/g" mgrt.yml

EDITOR=./tests/editor.sh mgrt add -m "Create users table"

printf "%b" "$revision" > revisions/"$now".sql

mgrt run | grep "up - performed revision"
mgrt log | grep "CREATE TABLE"
mgrt reset | grep "down - performed revision"
mgrt log | grep "DROP TABLE"

cd -

rm -rf "$DIR"
