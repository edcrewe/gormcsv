package importcsv

import (
	"context"
	"sync"
)

type batchJob struct {
	models any
	rows   int64
	endRow int
}

type batchResult struct {
	inserted   int64
	duplicates int64
	rejected   int64
	err        error
}

type batchProcessor func(context.Context, batchJob) batchResult

func runWorkers(ctx context.Context, count int, jobs <-chan batchJob, process batchProcessor) <-chan batchResult {
	results := make(chan batchResult, count)
	var workers sync.WaitGroup
	workers.Add(count)
	for range count {
		go func() {
			defer workers.Done()
			for job := range jobs {
				results <- process(ctx, job)
			}
		}()
	}
	go func() {
		workers.Wait()
		close(results)
	}()
	return results
}
