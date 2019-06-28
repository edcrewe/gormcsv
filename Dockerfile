# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from golang v1.12 base image
FROM golang:1.12

# Add Maintainer Info
LABEL maintainer="Ed Crewe <edmundcrewe@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/edcrewe/gormcsv

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

ENV GO111MODULE on

# Download all the dependencies
# https://stackoverflow.com/questions/28031603/what-do-three-dots-mean-in-go-command-line-invocations
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

# Run the unit and integration tests
CMD ["go test"]