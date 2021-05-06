#!/bin/sh

set -e

_version() {
	git log --decorate=full --format=format:%d |
		head -1 |
		tr ',' '\n' |
		grep tag: |
		cut -d / -f 3 |
		tr -d ',)'
}

[ ! -d bin ] && mkdir bin

bin="$(basename $(pwd))"
module="$(head -1 go.mod | awk '{ print $2 }')"
version="$(_version)"

[ "$version" = "" ] && {
	version="devel $(git log -n 1 --format='format: +%h %cd' HEAD)"
}

default_tags="netgo osusergo"
default_ldflags=$(printf -- "-X '%s/cmd/mgrt.Build=%s'" "$module" "$version")

tags="$TAGS $default_tags"
ldflags="$LDFLAGS $default_ldflags"

set -x
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags "$ldflags" -tags "$tags" -o bin/"$bin" ./cmd/mgrt
