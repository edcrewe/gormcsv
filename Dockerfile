# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from golang v1.12 base image
FROM golang:1.12-alpine AS build_base

# Add Maintainer Info
LABEL maintainer="Ed Crewe <edmundcrewe@gmail.com>"
# Install all build dependencies

# Install all build dependencies for modules
# Add bash for running tests and debugging purposes
RUN apk update \
    && apk add --virtual build-dependencies \
        build-base \
        gcc \
        wget \
        git \
    && apk add \
        bash
	
# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/edcrewe/gormcsv

ENV GO111MODULE on

COPY go.mod .
COPY go.sum .

RUN go mod download

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Run the unit tests
CMD ["go test -v"]