.PHONY: help build run test clean

help:
	@echo "build - build docker  image"
	@echo "run - start the flask service ready to schedule helm jobs"
	@echo "test - unit and integration tests"
	@echo "clean - remove all build, test, coverage and Python artifacts"

build:
	docker build -t gormcsv:0.1.0 .

clean:
	docker stop gormcsv || exit 0
	docker rm gormcsv || exit 0
	docker volume prune -f

run: clean 
	docker run -d --name gormcsv gormcsv:0.1.0 bash -c "go build" 

test: clean
	docker run -d --name gormcsv gormcsv:0.1.0 bash -c "cd tests;go test;go test --tags=integration"

