package importcsv

import (
	"context"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWorkerPoolProcessesBatchesConcurrently(t *testing.T) {
	const (
		workerCount = 4
		jobCount    = 20
	)
	jobs := make(chan batchJob)
	var active atomic.Int32
	var maximum atomic.Int32
	results := runWorkers(context.Background(), workerCount, jobs, func(_ context.Context, job batchJob) batchResult {
		current := active.Add(1)
		for {
			observed := maximum.Load()
			if current <= observed || maximum.CompareAndSwap(observed, current) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
		active.Add(-1)
		return batchResult{inserted: job.rows}
	})

	go func() {
		defer close(jobs)
		for index := 0; index < jobCount; index++ {
			jobs <- batchJob{rows: 1}
		}
	}()

	var inserted int64
	for result := range results {
		inserted += result.inserted
	}
	if inserted != jobCount {
		t.Errorf("inserted %d jobs, want %d", inserted, jobCount)
	}
	if maximum.Load() < 2 {
		t.Errorf("maximum concurrency was %d, want at least 2", maximum.Load())
	}
}

func TestWorkerPoolDrainsMoreResultsThanItsBuffer(t *testing.T) {
	const jobCount = 250
	jobs := make(chan batchJob, 1)
	results := runWorkers(context.Background(), 2, jobs, func(_ context.Context, job batchJob) batchResult {
		return batchResult{inserted: job.rows}
	})
	go func() {
		defer close(jobs)
		for range jobCount {
			jobs <- batchJob{rows: 1}
		}
	}()

	var completed int64
	for result := range results {
		completed += result.inserted
	}
	if completed != jobCount {
		t.Errorf("completed %d jobs, want %d", completed, jobCount)
	}
}

func TestResolveWorkerCount(t *testing.T) {
	sqliteDB, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "sqlite.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	closeDB(t, sqliteDB)
	if got := resolveWorkerCount(sqliteDB, 100); got != 1 {
		t.Errorf("SQLite worker count = %d, want 1", got)
	}

	pooledDB, err := gorm.Open(namedDialector{
		Dialector: sqlite.Open(filepath.Join(t.TempDir(), "pooled.db")),
		name:      "postgres",
	}, &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	closeDB(t, pooledDB)
	sqlDB, err := pooledDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(3)
	if got := resolveWorkerCount(pooledDB, 10); got != 3 {
		t.Errorf("pooled worker count = %d, want 3", got)
	}
	if got := resolveWorkerCount(pooledDB, 2); got != 2 {
		t.Errorf("configured worker count = %d, want 2", got)
	}
}

type namedDialector struct {
	gorm.Dialector
	name string
}

func (dialector namedDialector) Name() string {
	return dialector.name
}
