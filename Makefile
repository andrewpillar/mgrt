BUILD := `git rev-parse HEAD`
TAG   := `git describe --abbrev=0`

TAGS    := "netgo osusergo"
LFLAGS  := -ldflags "-X main.Build=$(BUILD) -X main.Tag=$(TAG)"

.PHONY: build test install

all: build

build:
	go build $(LFLAGS) -tags $(TAGS) -o mgrt.out
	go build $(LFLAGS) -tags sqlite3 -o mgrt-sqlite3.out

test:
	go test -v -cover ./... -tags sqlite3
