package importcsv

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"gorm.io/gorm"
)

const postgresTestEnvironment = "GORMCSV_TEST_POSTGRES"

var (
	usePostgresTests bool
	postgresTestDSN  string
	postgresAdmin    *gorm.DB
	postgresSchemaID atomic.Uint64
)

func TestMain(tests *testing.M) {
	enabled, err := strconv.ParseBool(os.Getenv(postgresTestEnvironment))
	if err != nil && os.Getenv(postgresTestEnvironment) != "" {
		fmt.Fprintf(os.Stderr, "%s must be a boolean: %v\n", postgresTestEnvironment, err)
		os.Exit(1)
	}
	if !enabled {
		os.Exit(tests.Run())
	}
	os.Exit(runPostgresTests(tests))
}

func runPostgresTests(tests *testing.M) (code int) {
	runtimePath, err := os.MkdirTemp("", "gormcsv-postgres-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create embedded PostgreSQL runtime: %v\n", err)
		return 1
	}
	defer func() {
		if err := os.RemoveAll(runtimePath); err != nil {
			fmt.Fprintf(os.Stderr, "remove embedded PostgreSQL runtime: %v\n", err)
			code = 1
		}
	}()

	port, err := availablePort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "select embedded PostgreSQL port: %v\n", err)
		return 1
	}
	var postgresLog bytes.Buffer
	config := embeddedpostgres.DefaultConfig().
		Port(port).
		RuntimePath(filepath.Join(runtimePath, "runtime")).
		StartTimeout(30 * time.Second).
		StartParameters(map[string]string{"max_connections": "50"}).
		Logger(&postgresLog)
	server := embeddedpostgres.NewDatabase(config)
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start embedded PostgreSQL: %v\n%s", err, postgresLog.String())
		return 1
	}

	postgresTestDSN = fmt.Sprintf(
		"host=localhost port=%d user=postgres password=postgres dbname=postgres sslmode=disable",
		port,
	)
	postgresAdmin, err = OpenDatabase("postgres", postgresTestDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to embedded PostgreSQL: %v\n", err)
		_ = server.Stop()
		return 1
	}
	usePostgresTests = true
	code = tests.Run()

	if sqlDB, dbErr := postgresAdmin.DB(); dbErr != nil {
		fmt.Fprintf(os.Stderr, "access embedded PostgreSQL connection: %v\n", dbErr)
		code = 1
	} else if closeErr := sqlDB.Close(); closeErr != nil {
		fmt.Fprintf(os.Stderr, "close embedded PostgreSQL connection: %v\n", closeErr)
		code = 1
	}
	if err := server.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "stop embedded PostgreSQL: %v\n%s", err, postgresLog.String())
		code = 1
	}
	return code
}

func availablePort() (uint32, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		return 0, err
	}
	return uint32(port), nil
}

func openPostgresTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	schema := fmt.Sprintf("gormcsv_test_%d", postgresSchemaID.Add(1))
	if err := postgresAdmin.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatalf("create PostgreSQL test schema: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresAdmin.Exec("DROP SCHEMA " + schema + " CASCADE").Error; err != nil {
			t.Errorf("drop PostgreSQL test schema: %v", err)
		}
	})

	db, err := OpenDatabase("postgres", postgresTestDSN+" search_path="+schema)
	if err != nil {
		t.Fatalf("open PostgreSQL test schema: %v", err)
	}
	return db
}
