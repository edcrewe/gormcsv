# Tag when on one, otherwise latest tag, commits since and short hash, e.g. v0.3.1-2-ga6866f6.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo devel)
LDFLAGS := -X github.com/edcrewe/gormcsv/cmd.Version=$(VERSION)

.PHONY: help build test test-postgres lint docker clean

help:
	@printf '%s\n' \
		'build   build the gormcsv binary' \
		'test    run tests with the race detector' \
		'test-postgres  run importer tests against embedded PostgreSQL' \
		'lint    run go vet and golangci-lint' \
		'docker  build the container image' \
		'clean   remove local build artifacts'

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o gormcsv .

test:
	go test -race ./...

test-postgres:
	GORMCSV_TEST_POSTGRES=1 go test -race -count=1 ./importcsv

lint:
	test -z "$$(gofmt -l .)"
	go vet ./...
	golangci-lint run --config=golangci-lint.yml

docker:
	docker build --build-arg VERSION=$(VERSION) --tag gormcsv:$(VERSION) --tag gormcsv:latest .

clean:
	rm -f gormcsv coverage.out
