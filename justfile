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

# serve the documentation locally
[group('documentation')]
serve-docs: docs
	python3 -m http.server -d site

# build the documentation site
[group('documentation')]
docs:
	rm -rf site
	pandoc docs/docs.md \
		--standalone \
		--from markdown \
		--to chunkedhtml \
		--variable toc \
		--toc-depth 2 \
		--chunk-template "%i.html" \
		--template docs/template.html \
		--output "site"

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
