VERSION:=$(shell (git describe --tags 2>/dev/null || echo 'v0.0.0') | cut -c2-)

SOURCE:=$(wildcard internal/*.go internal/*/*.go cmd/*/*.go)

build: go.mod sagan

tidy: go.mod

clean:
	rm -f sagan

sagan: $(SOURCE) go.sum
	CGO_ENABLED=0 go build -trimpath -ldflags '-s -w -X github.com/kgaughan/sagan/internal/version.Version=$(VERSION)' -o sagan ./cmd/sagan

update:
	go get -u ./...
	go mod tidy

go.mod: $(SOURCE)
	go mod tidy

.DEFAULT: build

.PHONY: build clean tidy update
