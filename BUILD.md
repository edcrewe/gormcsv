# Build

The normal development workflow uses the local Go toolchain:

```sh
make test
make test-postgres
make lint
make build
```

Go 1.25 or newer and a C compiler are required because the SQLite driver uses
CGO.

`make test-postgres` is optional and downloads an embedded PostgreSQL binary on
its first run. Later runs reuse the archive cached in
`~/.embedded-postgres-go`.

Build the production container separately:

```sh
make docker
```

The multi-stage Docker build runs the test suite and copies only the compiled
binary into the unprivileged runtime image.
