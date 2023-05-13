NAME:=sagan

VERSION:=$(shell (git describe --tags 2>/dev/null || echo 'v0.0.0') | cut -c2-)

SOURCE:=$(wildcard internal/*.go internal/*/*.go cmd/*/*.go)

build: go.mod $(NAME)

tidy: go.mod fmt

clean:
	rm -f $(NAME)

$(NAME): $(SOURCE) go.sum
	CGO_ENABLED=0 go build -trimpath -ldflags '-s -w -X github.com/kgaughan/$(NAME)/internal/version.Version=$(VERSION)' -o $(NAME) ./cmd/$(NAME)

fmt:
	go fmt ./...

lint:
	go vet ./...

update:
	go get -u ./...
	go mod tidy

go.mod: $(SOURCE)
	go mod tidy

.DEFAULT: build

.PHONY: build clean tidy update fmt lint
