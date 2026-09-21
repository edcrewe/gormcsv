VERSION := 0.3.0

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
	go build -trimpath -o gormcsv .

test:
	go test -race ./...

test-postgres:
	GORMCSV_TEST_POSTGRES=1 go test -race -count=1 ./importcsv

lint:
	test -z "$$(gofmt -l .)"
	go vet ./...
	golangci-lint run --config=golangci-lint.yml

docker:
	docker build --tag gormcsv:$(VERSION) .

clean:
	rm -f gormcsv coverage.out
