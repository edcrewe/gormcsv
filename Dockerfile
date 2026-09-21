FROM golang:1.25-alpine AS build

RUN apk add --no-cache build-base
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go test ./... \
    && go build -trimpath -ldflags="-s -w" -o /out/gormcsv .

FROM alpine:3.22

RUN addgroup -S gormcsv \
    && adduser -S -G gormcsv gormcsv \
    && mkdir /data \
    && chown gormcsv:gormcsv /data
COPY --from=build /out/gormcsv /usr/local/bin/gormcsv

USER gormcsv
WORKDIR /data
ENTRYPOINT ["gormcsv"]
