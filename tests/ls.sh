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
EDITOR=./tests/editor.sh mgrt add -m "Create users table"

printf "%b" "$revision" > revisions/"$now".sql

mgrt ls | grep "$now"

cd -

rm -rf "$DIR"
