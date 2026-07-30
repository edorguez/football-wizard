package scraper

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"
)

type JobResult struct {
	Label string
	Err   error
}

type WorkerPool struct {
	client         *Client
	cache          *Cache
	logger         *slog.Logger
	workers        int
	rateLimitDelay time.Duration
	jitter         float64
}

func NewWorkerPool(client *Client, cache *Cache, logger *slog.Logger, workers int, rateLimitDelay time.Duration, jitter float64) *WorkerPool {
	if workers < 1 {
		workers = 1
	}
	return &WorkerPool{
		client:         client,
		cache:          cache,
		logger:         logger,
		workers:        workers,
		rateLimitDelay: rateLimitDelay,
		jitter:         jitter,
	}
}

type FetchJob struct {
	Label      string
	URL        string
	Season     int
	RequiresJS bool
	CacheParts []string
	ParseFn    func(season int, html string) error
}

func (wp *WorkerPool) Run(jobs []FetchJob) []JobResult {
	var wg sync.WaitGroup
	jobCh := make(chan FetchJob, len(jobs))
	resultCh := make(chan JobResult, len(jobs))

	for range wp.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				result := wp.process(job)
				resultCh <- result
			}
		}()
	}

	for _, job := range jobs {
		jobCh <- job
	}
	close(jobCh)

	wg.Wait()
	close(resultCh)

	var results []JobResult
	for r := range resultCh {
		results = append(results, r)
	}

	return results
}

func (wp *WorkerPool) process(job FetchJob) JobResult {
	if len(job.CacheParts) > 0 {
		if cached, ok := wp.cache.Read(job.Season, job.CacheParts...); ok {
			wp.logger.Debug("using cached HTML", "label", job.Label)
			if err := job.ParseFn(job.Season, cached); err != nil {
				return JobResult{Label: job.Label, Err: fmt.Errorf("parsing cached %s: %w", job.Label, err)}
			}
			wp.throttle()
			return JobResult{Label: job.Label}
		}
	}

	var html string
	var err error

	if job.RequiresJS {
		html, err = wp.client.FetchHTMLWithJS(job.URL)
	} else {
		html, err = wp.client.FetchHTML(job.URL)
	}

	if err != nil {
		return JobResult{Label: job.Label, Err: fmt.Errorf("fetching %s: %w", job.Label, err)}
	}

	if len(job.CacheParts) > 0 {
		wp.cache.Write(job.Season, html, job.CacheParts...)
	}

	if err := job.ParseFn(job.Season, html); err != nil {
		return JobResult{Label: job.Label, Err: fmt.Errorf("parsing %s: %w", job.Label, err)}
	}

	wp.throttle()

	return JobResult{Label: job.Label}
}

func (wp *WorkerPool) throttle() {
	if wp.rateLimitDelay <= 0 {
		return
	}

	jitterMs := time.Duration(float64(wp.rateLimitDelay) * wp.jitter * (rand.Float64()*2 - 1))
	sleepTime := wp.rateLimitDelay + jitterMs

	if sleepTime < 0 {
		sleepTime = 0
	}

	time.Sleep(sleepTime)
}
