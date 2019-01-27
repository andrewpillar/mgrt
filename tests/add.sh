#!/bin/sh

set -ex

DIR=$(mktemp -d)

now=$(date +"%s")

cd "$DIR"

mgrt init
EDITOR=./tests/editor.sh mgrt add -m "Create users table"

grep "\-\- mgrt: revision: $now" revisions/$(ls revisions)

cd -

rm -rf "$DIR"
