#!/bin/sh

set -ex

DIR=$(mktemp -d)

cd "$DIR"

mgrt init

[ -f mgrt.yml ]
[ -d revisions ]

mgrt init database

[ -f database/mgrt.yml ]
[ -d database/revisions ]

cd -

rm -rf "$DIR"
