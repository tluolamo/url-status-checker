package checker

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/tluolamo/url-status-checker/internal/models"
)

// Checker handles concurrent URL availability checking.
type Checker struct {
	client     *http.Client
	maxWorkers int
}

// New creates a new Checker instance.
func New(timeout time.Duration, maxWorkers int) *Checker {
	return &Checker{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		maxWorkers: maxWorkers,
	}
}

// CheckURLs checks multiple URLs concurrently using goroutines and channels.
func (c *Checker) CheckURLs(ctx context.Context, urls []string) []models.CheckResult {
	jobs := make(chan string, len(urls))
	results := make(chan models.CheckResult, len(urls))

	workerCount := c.maxWorkers
	if len(urls) < workerCount {
		workerCount = len(urls)
	}
	if workerCount == 0 {
		return []models.CheckResult{}
	}

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go c.worker(ctx, jobs, results, &wg)
	}

	go func() {
		defer close(jobs)
		for _, url := range urls {
			select {
			case jobs <- url:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var checkResults []models.CheckResult
	for result := range results {
		checkResults = append(checkResults, result)
	}

	return checkResults
}

func (c *Checker) worker(ctx context.Context, jobs <-chan string, results chan<- models.CheckResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			results <- c.checkURL(ctx, url)
		}
	}
}

func (c *Checker) checkURL(ctx context.Context, url string) models.CheckResult {
	result := models.CheckResult{
		URL:       url,
		CheckedAt: time.Now(),
	}

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		return result
	}

	req.Header.Set("User-Agent", "URL-Status-Checker/1.0")

	resp, err := c.client.Do(req)

	duration := time.Since(start)
	result.ResponseTimeMs = duration.Milliseconds()

	if err != nil {
		result.Error = fmt.Sprintf("request failed: %v", err)
		return result
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log would require logger field; silently ignore for now
			_ = closeErr
		}
	}()

	result.StatusCode = resp.StatusCode
	result.Available = resp.StatusCode >= 200 && resp.StatusCode < 400

	return result
}

// CheckURL is a convenience method to check a single URL.
func (c *Checker) CheckURL(ctx context.Context, url string) models.CheckResult {
	return c.checkURL(ctx, url)
}
