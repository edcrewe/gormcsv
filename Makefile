.PHONY: help build run test clean

help:
	@echo "build - build docker  image"
	@echo "run - build gormcsv and run demo import of csv file"
	@echo "test - unit and integration tests"
	@echo "clean - remove all build, test, coverage and build artifacts"

build: clean
	docker build -t gormcsv:0.1.1 .
	docker run -d --name gormcsv -v ${PWD}:/go/src/github.com/edcrewe/gormcsv gormcsv:0.1.1 bash -c "go build -v"
	docker logs -f gormcsv

clean:
	docker stop gormcsv || exit 0
	docker rm gormcsv || exit 0
	docker volume prune -f

run: clean 
	docker run -d --name gormcsv -v ${PWD}:/go/src/github.com/edcrewe/gormcsv gormcsv:0.1.1 bash -c "go build;chmod755 gormcsv;./gormcsv importcsv -f tests/fixtures/Country.csv"
	docker logs -f gormcsv

test: clean
	docker run -d --name gormcsv gormcsv:0.1.1 bash -c "go build;cd tests;go test -v --tags=u,i"
	docker logs -f gormcsv
