.PHONY: help build run test clean

help:
	@echo "build - build docker  image"
	@echo "run - build gormcsv and run demo import of csv file"
	@echo "test - unit and integration tests"
	@echo "clean - remove all build, test, coverage and build artifacts"
	@echo "lint - lint the code"

build: clean
	docker build -t gormcsv:0.1.1 .
	docker run -a stdout --name gormcsv -v ${PWD}:/go/src/github.com/edcrewe/gormcsv gormcsv:0.1.1 bash -c "go build -v"

clean:
	docker stop gormcsv || exit 0
	docker rm gormcsv || exit 0
	docker volume prune -f

run: clean 
	docker run -a stdout --name gormcsv -v ${PWD}:/go/src/github.com/edcrewe/gormcsv gormcsv:0.1.1 bash -c "go build;chmod755 gormcsv;./gormcsv importcsv -f static/fixtures/Country.csv"

test: clean
	docker run -a stdout --name gormcsv gormcsv:0.1.1 bash -c "go build;cd tests;go test -v --tags=u,i"

lint: clean
	docker run -d --name gormcsv -v ${PWD}:/go/src/github.com/edcrewe/gormcsv gormcsv:0.1.1 bash -c "golangci-lint run ./... -c golangci-lint.yml -v --timeout 5m"
	docker logs -f gormcsv
