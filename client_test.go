package arxiv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// Mock XML response for testing
const mockXMLResponse = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <link href="http://arxiv.org/api/query?search_query=quantum+computing&amp;id_list=&amp;start=0&amp;max_results=1" rel="self" type="application/atom+xml"/>
  <title type="html">ArXiv Query: search_query=quantum computing&amp;id_list=&amp;start=0&amp;max_results=1</title>
  <id>http://arxiv.org/api/query?search_query=quantum+computing&amp;id_list=&amp;start=0&amp;max_results=1</id>
  <updated>2023-01-01T00:00:00-05:00</updated>
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">50000</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">1</opensearch:itemsPerPage>
  <entry>
    <id>http://arxiv.org/abs/1234.5678v1</id>
    <updated>2023-01-01T00:00:00-05:00</updated>
    <published>2023-01-01T00:00:00-05:00</published>
    <title>Test Paper on Quantum Computing</title>
    <summary>This is a test abstract for quantum computing research.</summary>
    <author>
      <name>John Doe</name>
    </author>
    <author>
      <name>Jane Smith</name>
    </author>
    <arxiv:doi xmlns:arxiv="http://arxiv.org/schemas/atom">10.1234/test.doi</arxiv:doi>
    <link href="http://arxiv.org/abs/1234.5678v1" rel="alternate" type="text/html"/>
    <link title="pdf" href="http://arxiv.org/pdf/1234.5678v1.pdf" rel="related" type="application/pdf"/>
    <arxiv:comment xmlns:arxiv="http://arxiv.org/schemas/atom">Test comment</arxiv:comment>
    <category term="quant-ph" scheme="http://arxiv.org/schemas/atom"/>
    <category term="cs.ET" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
</feed>`

// =============================================================================
// Client Creation Tests
// =============================================================================

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.httpClient == nil {
		t.Fatal("HTTP client is nil")
	}

	if client.baseURL != baseURL {
		t.Fatalf("Expected base URL %s, got %s", baseURL, client.baseURL)
	}

	// Verify default options are set
	opts := client.options
	if opts.RetryAttempts != defaultRetryAttempts {
		t.Errorf("Expected default RetryAttempts %d, got %d", defaultRetryAttempts, opts.RetryAttempts)
	}
	if opts.UserAgent != defaultUserAgent {
		t.Errorf("Expected default UserAgent %s, got %s", defaultUserAgent, opts.UserAgent)
	}
}

func TestNewClientWithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 10 * time.Second}
	client := NewClientWithHTTPClient(customClient)

	if client == nil {
		t.Fatal("NewClientWithHTTPClient() returned nil")
	}

	if client.httpClient != customClient {
		t.Fatal("Custom HTTP client not set correctly")
	}

	if client.baseURL != baseURL {
		t.Fatalf("Expected base URL %s, got %s", baseURL, client.baseURL)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	opts := ClientOptions{
		RetryAttempts: 5,
		RetryDelay:    2 * time.Second,
		RateLimit:     1 * time.Second,
		UserAgent:     "test-agent/1.0",
		Timeout:       10 * time.Second,
	}

	client := NewClientWithOptions(opts)

	if client == nil {
		t.Fatal("NewClientWithOptions() returned nil")
	}

	if client.options.RetryAttempts != 5 {
		t.Errorf("Expected RetryAttempts 5, got %d", client.options.RetryAttempts)
	}

	if client.options.UserAgent != "test-agent/1.0" {
		t.Errorf("Expected UserAgent 'test-agent/1.0', got '%s'", client.options.UserAgent)
	}

	if client.options.RetryDelay != 2*time.Second {
		t.Errorf("Expected RetryDelay 2s, got %v", client.options.RetryDelay)
	}

	if client.options.RateLimit != 1*time.Second {
		t.Errorf("Expected RateLimit 1s, got %v", client.options.RateLimit)
	}

	if client.options.Timeout != 10*time.Second {
		t.Errorf("Expected Timeout 10s, got %v", client.options.Timeout)
	}
}

func TestDefaultClientOptions(t *testing.T) {
	opts := DefaultClientOptions()

	if opts.RetryAttempts != defaultRetryAttempts {
		t.Errorf("Expected default RetryAttempts %d, got %d", defaultRetryAttempts, opts.RetryAttempts)
	}

	if opts.UserAgent != defaultUserAgent {
		t.Errorf("Expected default UserAgent '%s', got '%s'", defaultUserAgent, opts.UserAgent)
	}

	if opts.RateLimit != defaultRateLimit {
		t.Errorf("Expected default RateLimit %v, got %v", defaultRateLimit, opts.RateLimit)
	}

	if opts.RetryDelay != defaultRetryDelay {
		t.Errorf("Expected default RetryDelay %v, got %v", defaultRetryDelay, opts.RetryDelay)
	}

	if opts.Timeout != defaultTimeout {
		t.Errorf("Expected default Timeout %v, got %v", defaultTimeout, opts.Timeout)
	}
}

func TestNewClientWithOptionsZeroValues(t *testing.T) {
	// Test that zero values are replaced with defaults
	opts := ClientOptions{} // All zero values

	client := NewClientWithOptions(opts)

	if client.options.RetryAttempts != defaultRetryAttempts {
		t.Errorf("Expected default RetryAttempts for zero value, got %d", client.options.RetryAttempts)
	}

	if client.options.UserAgent != defaultUserAgent {
		t.Errorf("Expected default UserAgent for zero value, got '%s'", client.options.UserAgent)
	}

	if client.options.RateLimit != defaultRateLimit {
		t.Errorf("Expected default RateLimit for zero value, got %v", client.options.RateLimit)
	}
}

// =============================================================================
// HTTP Communication Tests
// =============================================================================

func TestSearch(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check User-Agent header
		if ua := r.Header.Get("User-Agent"); ua != defaultUserAgent {
			t.Errorf("Expected User-Agent '%s', got '%s'", defaultUserAgent, ua)
		}

		// Return mock XML response
		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockXMLResponse))
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient()
	client.baseURL = server.URL

	// Test search
	query := &Query{
		SearchQuery: "quantum computing",
		MaxResults:  1,
	}

	results, err := client.Search(context.Background(), query)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Verify results
	if results == nil {
		t.Fatal("Results is nil")
	}

	if results.TotalCount != 50000 {
		t.Errorf("Expected total count 50000, got %d", results.TotalCount)
	}

	if len(results.Papers) != 1 {
		t.Errorf("Expected 1 paper, got %d", len(results.Papers))
	}

	paper := results.Papers[0]
	if paper.ID != "1234.5678v1" {
		t.Errorf("Expected ID '1234.5678v1', got '%s'", paper.ID)
	}

	if paper.Title != "Test Paper on Quantum Computing" {
		t.Errorf("Expected title 'Test Paper on Quantum Computing', got '%s'", paper.Title)
	}

	if len(paper.Authors) != 2 {
		t.Errorf("Expected 2 authors, got %d", len(paper.Authors))
	}

	if paper.DOI != "10.1234/test.doi" {
		t.Errorf("Expected DOI '10.1234/test.doi', got '%s'", paper.DOI)
	}
}

func TestGetByID(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that id_list parameter is set correctly
		idList := r.URL.Query().Get("id_list")
		if idList != "1234.5678" {
			t.Errorf("Expected id_list '1234.5678', got '%s'", idList)
		}

		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockXMLResponse))
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient()
	client.baseURL = server.URL

	// Test GetByID
	paper, err := client.GetByID(context.Background(), "1234.5678")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if paper == nil {
		t.Fatal("Paper is nil")
	}

	if paper.ID != "1234.5678v1" {
		t.Errorf("Expected ID '1234.5678v1', got '%s'", paper.ID)
	}
}

func TestSearchWithCustomUserAgent(t *testing.T) {
	// Create a test server that checks User-Agent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ua := r.Header.Get("User-Agent"); ua != "custom-agent/2.0" {
			t.Errorf("Expected User-Agent 'custom-agent/2.0', got '%s'", ua)
		}

		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockXMLResponse))
	}))
	defer server.Close()

	// Create client with custom User-Agent
	client := NewClientWithOptions(ClientOptions{
		UserAgent: "custom-agent/2.0",
	})
	client.baseURL = server.URL

	query := &Query{
		SearchQuery: "test",
		MaxResults:  1,
	}

	_, err := client.Search(context.Background(), query)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestSearchWithNilQuery(t *testing.T) {
	client := NewClient()
	_, err := client.Search(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil query, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Error("Expected APIError type")
	}

	if apiErr.Type != ErrorTypeInvalidQuery {
		t.Errorf("Expected ErrorTypeInvalidQuery, got %v", apiErr.Type)
	}
}

func TestGetByIDWithEmptyID(t *testing.T) {
	client := NewClient()
	_, err := client.GetByID(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Error("Expected APIError type")
	}

	if apiErr.Type != ErrorTypeInvalidQuery {
		t.Errorf("Expected ErrorTypeInvalidQuery, got %v", apiErr.Type)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty feed
		emptyResponse := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:itemsPerPage>
</feed>`
		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(emptyResponse))
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	_, err := client.GetByID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent paper")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}

	if apiErr.Type != ErrorTypeNotFound {
		t.Errorf("Expected ErrorTypeNotFound, got %v", apiErr.Type)
	}
}

func TestSearchNetworkError(t *testing.T) {
	// Create client with invalid URL to trigger network error
	client := NewClient()
	client.baseURL = "http://invalid-url-that-does-not-exist"

	query := &Query{
		SearchQuery: "test",
		MaxResults:  1,
	}

	_, err := client.Search(context.Background(), query)
	if err == nil {
		t.Error("Expected network error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Error("Expected APIError type")
	}

	if apiErr.Type != ErrorTypeNetwork {
		t.Errorf("Expected ErrorTypeNetwork, got %v", apiErr.Type)
	}
}

// =============================================================================
// Retry Mechanism Tests
// =============================================================================

func TestSearchWithRetryRateLimit(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockXMLResponse))
	}))
	defer server.Close()

	client := NewClientWithOptions(ClientOptions{
		RetryAttempts: 3,
		RetryDelay:    10 * time.Millisecond, // Fast for testing
		RateLimit:     1 * time.Millisecond,  // Fast for testing
	})
	client.baseURL = server.URL

	query := &Query{
		SearchQuery: "test",
		MaxResults:  1,
	}

	_, err := client.Search(context.Background(), query)
	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestSearchWithRetryServerError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockXMLResponse))
	}))
	defer server.Close()

	client := NewClientWithOptions(ClientOptions{
		RetryAttempts: 3,
		RetryDelay:    10 * time.Millisecond,
		RateLimit:     1 * time.Millisecond,
	})
	client.baseURL = server.URL

	query := &Query{
		SearchQuery: "test",
		MaxResults:  1,
	}

	_, err := client.Search(context.Background(), query)
	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestSearchRetryExhaustion(t *testing.T) {
	// Server always returns rate limit error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClientWithOptions(ClientOptions{
		RetryAttempts: 2,
		RetryDelay:    1 * time.Millisecond,
		RateLimit:     1 * time.Millisecond,
	})
	client.baseURL = server.URL

	query := &Query{
		SearchQuery: "test",
		MaxResults:  1,
	}

	_, err := client.Search(context.Background(), query)
	if err == nil {
		t.Error("Expected error after retry exhaustion")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Error("Expected APIError type")
	}

	if apiErr.Type != ErrorTypeRateLimit {
		t.Errorf("Expected ErrorTypeRateLimit, got %v", apiErr.Type)
	}
}

func TestSearchContextCancellation(t *testing.T) {
	// Server with long delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockXMLResponse))
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	// Create context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	query := &Query{
		SearchQuery: "test",
		MaxResults:  1,
	}

	_, err := client.Search(ctx, query)
	if err == nil {
		t.Error("Expected context cancellation error")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

// =============================================================================
// Rate Limiting Tests
// =============================================================================

func TestRateLimiting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockXMLResponse))
	}))
	defer server.Close()

	client := NewClientWithOptions(ClientOptions{
		RateLimit: 100 * time.Millisecond,
	})
	client.baseURL = server.URL

	query := &Query{
		SearchQuery: "test",
		MaxResults:  1,
	}

	// Make first request
	start := time.Now()
	_, err := client.Search(context.Background(), query)
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}

	// Make second request immediately
	_, err = client.Search(context.Background(), query)
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}
	elapsed := time.Since(start)

	// Second request should be delayed by rate limit
	if elapsed < 100*time.Millisecond {
		t.Errorf("Expected delay due to rate limiting, elapsed time: %v", elapsed)
	}
}

func TestRateLimitingDisabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockXMLResponse))
	}))
	defer server.Close()

	client := NewClientWithOptions(ClientOptions{
		RateLimit: 1 * time.Millisecond, // Very fast rate limiting
	})
	client.baseURL = server.URL

	query := &Query{
		SearchQuery: "test",
		MaxResults:  1,
	}

	// Make two requests quickly
	start := time.Now()
	_, err := client.Search(context.Background(), query)
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}

	_, err = client.Search(context.Background(), query)
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}
	elapsed := time.Since(start)

	// Should be very fast with minimal rate limiting
	if elapsed > 100*time.Millisecond {
		t.Errorf("Unexpected delay with minimal rate limiting, elapsed time: %v", elapsed)
	}
}

func TestConcurrentRateLimiting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockXMLResponse))
	}))
	defer server.Close()

	client := NewClientWithOptions(ClientOptions{
		RateLimit: 50 * time.Millisecond,
	})
	client.baseURL = server.URL

	query := &Query{
		SearchQuery: "test",
		MaxResults:  1,
	}

	// Test concurrent requests
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	start := time.Now()
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.Search(context.Background(), query)
			mu.Lock()
			if err != nil {
				errors = append(errors, err)
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	if len(errors) > 0 {
		t.Errorf("Got %d errors in concurrent requests: %v", len(errors), errors[0])
	}

	// With 3 requests and 50ms rate limit, should take at least 100ms
	expectedMinDuration := 100 * time.Millisecond
	if elapsed < expectedMinDuration {
		t.Errorf("Expected at least %v for 3 concurrent requests with rate limiting, got %v", expectedMinDuration, elapsed)
	}
}

// =============================================================================
// Query Building Tests
// =============================================================================

func TestBuildQueryParams(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name     string
		query    *Query
		expected map[string]string
	}{
		{
			name: "basic search query",
			query: &Query{
				SearchQuery: "quantum computing",
				MaxResults:  5,
			},
			expected: map[string]string{
				"search_query": "quantum computing",
				"max_results":  "5",
				"sortBy":       "relevance",
				"sortOrder":    "descending",
			},
		},
		{
			name: "ID list query",
			query: &Query{
				IDList:     []string{"1234.5678", "9876.5432"},
				MaxResults: 2,
			},
			expected: map[string]string{
				"id_list":     "1234.5678,9876.5432",
				"max_results": "2",
				"sortBy":      "relevance",
				"sortOrder":   "descending",
			},
		},
		{
			name: "pagination and sorting",
			query: &Query{
				SearchQuery: "machine learning",
				Start:       10,
				MaxResults:  20,
				SortBy:      "submittedDate",
				SortOrder:   "ascending",
			},
			expected: map[string]string{
				"search_query": "machine learning",
				"start":        "10",
				"max_results":  "20",
				"sortBy":       "submittedDate",
				"sortOrder":    "ascending",
			},
		},
		{
			name: "default values",
			query: &Query{
				SearchQuery: "test",
			},
			expected: map[string]string{
				"search_query": "test",
				"max_results":  "500",
				"sortBy":       "relevance",
				"sortOrder":    "descending",
			},
		},
		{
			name: "zero max results uses default",
			query: &Query{
				SearchQuery: "test",
				MaxResults:  0,
			},
			expected: map[string]string{
				"search_query": "test",
				"max_results":  "500",
				"sortBy":       "relevance",
				"sortOrder":    "descending",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := client.buildQueryParams(tt.query)

			for key, expected := range tt.expected {
				if got := params.Get(key); got != expected {
					t.Errorf("Parameter %s: expected %s, got %s", key, expected, got)
				}
			}
		})
	}
}

func TestBuildDateRangeFilter(t *testing.T) {
	client := NewClient()

	// Test date range
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

	filter := client.buildDateRangeFilter(&from, &to)
	expected := "submittedDate:[20200101 TO 20231231]"
	if filter != expected {
		t.Errorf("Expected '%s', got '%s'", expected, filter)
	}

	// Test from only
	filter = client.buildDateRangeFilter(&from, nil)
	expected = "submittedDate:[20200101 TO *]"
	if filter != expected {
		t.Errorf("Expected '%s', got '%s'", expected, filter)
	}

	// Test to only
	filter = client.buildDateRangeFilter(nil, &to)
	expected = "submittedDate:[* TO 20231231]"
	if filter != expected {
		t.Errorf("Expected '%s', got '%s'", expected, filter)
	}

	// Test neither
	filter = client.buildDateRangeFilter(nil, nil)
	if filter != "" {
		t.Errorf("Expected empty string, got '%s'", filter)
	}
}

func TestSearchWithDateRange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		searchQuery := r.URL.Query().Get("search_query")
		expectedQuery := "(quantum computing) AND submittedDate:[20200101 TO 20231231]"
		if searchQuery != expectedQuery {
			t.Errorf("Expected search query '%s', got '%s'", expectedQuery, searchQuery)
		}

		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockXMLResponse))
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

	query := &Query{
		SearchQuery:       "quantum computing",
		MaxResults:        1,
		SubmittedDateFrom: &from,
		SubmittedDateTo:   &to,
	}

	_, err := client.Search(context.Background(), query)
	if err != nil {
		t.Fatalf("Search with date range failed: %v", err)
	}
}

func TestSearchDateRangeOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		searchQuery := r.URL.Query().Get("search_query")
		expectedQuery := "submittedDate:[20200101 TO 20231231]"
		if searchQuery != expectedQuery {
			t.Errorf("Expected search query '%s', got '%s'", expectedQuery, searchQuery)
		}

		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockXMLResponse))
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

	query := &Query{
		MaxResults:        1,
		SubmittedDateFrom: &from,
		SubmittedDateTo:   &to,
	}

	_, err := client.Search(context.Background(), query)
	if err != nil {
		t.Fatalf("Search with date range only failed: %v", err)
	}
}

// =============================================================================
// Factory Method Tests
// =============================================================================

func TestNewQuery(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery()

	if qb == nil {
		t.Fatal("NewQuery() returned nil")
	}

	if qb.client != client {
		t.Error("QueryBuilder client reference not set correctly")
	}

	// Verify default values are set
	if qb.maxResults != defaultMaxResults {
		t.Errorf("Expected default maxResults %d, got %d", defaultMaxResults, qb.maxResults)
	}

	if qb.sortBy != SortByRelevance {
		t.Errorf("Expected default sortBy %s, got %s", SortByRelevance, qb.sortBy)
	}

	if qb.sortOrder != SortOrderDescending {
		t.Errorf("Expected default sortOrder %s, got %s", SortOrderDescending, qb.sortOrder)
	}
}

func TestIterator(t *testing.T) {
	client := NewClient()
	query := &Query{
		SearchQuery: "test",
		MaxResults:  10,
	}

	ctx := context.Background()
	iter := client.Iterator(ctx, query)

	if iter == nil {
		t.Fatal("Iterator() returned nil")
	}

	// Test that iterator is properly initialized
	if iter.query != query {
		t.Error("Iterator query reference not set correctly")
	}

	// Test that we can get the current state
	state := iter.stateManager.GetState()
	if state.Current != StateInitial {
		t.Errorf("Expected initial state, got %s", state.Current.String())
	}
}

// =============================================================================
// Utility Function Tests
// =============================================================================

func TestExtractArxivID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "http://arxiv.org/abs/1234.5678v1",
			expected: "1234.5678v1",
		},
		{
			input:    "1234.5678v1",
			expected: "1234.5678v1",
		},
		{
			input:    "http://arxiv.org/abs/quant-ph/0301001",
			expected: "quant-ph/0301001",
		},
		{
			input:    "quant-ph/0301001",
			expected: "quant-ph/0301001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractArxivID(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
