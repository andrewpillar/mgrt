BUILD := `git rev-parse HEAD`
TAG   := `git describe --abbrev=0`

LDFLAGS := -ldflags "-X main.Build=$(BUILD) -X main.Tag=$(TAG)"

.PHONY: all test install

all: install

test:
	go test -v -cover ./...

install: test
	go install $(LDFLAGS)
