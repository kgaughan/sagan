export CGO_ENABLED := "0"

# build the sagan binary
build:
	go build -v -tags netgo -trimpath -ldflags '-s -w' -o sagan ./cmd/sagan

# update dependencies
[group('maintenance')]
update:
	go get -u ./...
	go mod verify
	go mod tidy

# format the code
[group('maintenance')]
fmt:
	go fmt ./...

# lint the code
[group('maintenance')]
lint:
	go vet ./...
	golangci-lint run ./...

# clean build artifacts
[group('maintenance')]
clean:
	find . -name \*.orig -delete
	rm -rf sagan dist site .venv coverage.out coverage.html

[private]
venv:
	test -e .venv || uv venv
	uv pip install -r requirements.txt

# rebuild documentation reqirements
[group('documentation')]
requirements:
	uv pip compile -q requirements.in -o requirements.txt

# serve the documentation locally
[group('documentation')]
serve-docs: venv
	uv run mkdocs serve

# build the documentation site
[group('documentation')]
docs: venv
	uv run mkdocs build

# run the test suite
[group('testing')]
tests:
	go test -cover -coverprofile=coverage.out -v ./...

# generate HTML report from coverage data
[group('testing')]
coverage-html: tests
	go tool cover -html=coverage.out -o coverage.html

# run `goreleaser release` without publishing anything
[group('testing')]
test-release:
	goreleaser release --auto-snapshot --clean --skip docker --skip publish
