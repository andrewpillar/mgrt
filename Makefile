

GOPATH      := $(shell go env GOPATH)
GIT_VERSION := $(shell ./version)

TAGS    := netgo,osusergo
LDFLAGS := -s -w -X 'main.Build=$(GIT_VERSION)'

BIN := bin

.PHONY: mgrt

mgrt: gen fmt test
	mkdir -p $(BIN)
	go build -trimpath -ldflags "$(LDFLAGS)" -tags "$(TAGS)" -o $(BIN)/mgrt ./cmd/mgrt

gen:
	go generate ./...

fmt:
	gofmt -s -w .

test:
	go test -cover ./...

install: mgrt
	cp bin/mgrt $(GOPATH)/bin/
