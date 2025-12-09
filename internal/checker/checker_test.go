package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	timeout := 5 * time.Second
	maxWorkers := 50

	checker := New(timeout, maxWorkers)

	assert.NotNil(t, checker)
	assert.NotNil(t, checker.client)
	assert.Equal(t, maxWorkers, checker.maxWorkers)
	assert.Equal(t, timeout, checker.client.Timeout)
}

func TestCheckURLSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := New(5*time.Second, 10)
	ctx := context.Background()

	result := checker.CheckURL(ctx, server.URL)

	assert.Equal(t, server.URL, result.URL)
	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.True(t, result.Available)
	assert.Empty(t, result.Error)
	assert.Greater(t, result.ResponseTimeMs, int64(0))
}

func TestCheckURLNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	checker := New(5*time.Second, 10)
	ctx := context.Background()

	result := checker.CheckURL(ctx, server.URL)

	assert.Equal(t, http.StatusNotFound, result.StatusCode)
	assert.False(t, result.Available)
	assert.Empty(t, result.Error)
}

func TestCheckURLTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := New(100*time.Millisecond, 10)
	ctx := context.Background()

	result := checker.CheckURL(ctx, server.URL)

	assert.NotEmpty(t, result.Error)
	assert.False(t, result.Available)
	assert.Contains(t, result.Error, "request failed")
}

func TestCheckURLInvalidURL(t *testing.T) {
	checker := New(5*time.Second, 10)
	ctx := context.Background()

	result := checker.CheckURL(ctx, "://invalid-url")

	assert.NotEmpty(t, result.Error)
	assert.False(t, result.Available)
}

func TestCheckURLsMultiple(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server2.Close()

	checker := New(5*time.Second, 10)
	ctx := context.Background()

	urls := []string{server1.URL, server2.URL}
	results := checker.CheckURLs(ctx, urls)

	require.Len(t, results, 2)

	var ok, notFound int
	for _, result := range results {
		if result.StatusCode == http.StatusOK {
			ok++
			assert.True(t, result.Available)
		} else if result.StatusCode == http.StatusNotFound {
			notFound++
			assert.False(t, result.Available)
		}
	}

	assert.Equal(t, 1, ok)
	assert.Equal(t, 1, notFound)
}

func TestCheckURLsConcurrency(t *testing.T) {
	var mu sync.Mutex
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	urls := make([]string, 10)
	for i := range urls {
		urls[i] = server.URL
	}

	checker := New(5*time.Second, 5)
	ctx := context.Background()

	start := time.Now()
	results := checker.CheckURLs(ctx, urls)
	duration := time.Since(start)

	require.Len(t, results, 10)
	assert.Equal(t, 10, callCount)
	assert.Less(t, duration, 500*time.Millisecond, "Should complete faster with concurrency")
}

func TestCheckURLsContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	urls := []string{server.URL, server.URL, server.URL}

	checker := New(5*time.Second, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	results := checker.CheckURLs(ctx, urls)

	for _, result := range results {
		if result.Error != "" {
			assert.Contains(t, result.Error, "request failed")
		}
	}
}

func BenchmarkCheckURL(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := New(5*time.Second, 10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.CheckURL(ctx, server.URL)
	}
}

func BenchmarkCheckURLs10URLs(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	urls := make([]string, 10)
	for i := range urls {
		urls[i] = server.URL
	}

	checker := New(5*time.Second, 10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.CheckURLs(ctx, urls)
	}
}
