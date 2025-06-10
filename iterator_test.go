package arxiv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIterator_BasicIteration(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response with 2 papers
		response := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">100</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">2</opensearch:itemsPerPage>
  <entry>
    <id>http://arxiv.org/abs/1234.5678v1</id>
    <title>Test Paper 1</title>
    <summary>Abstract 1</summary>
    <published>2023-01-01T00:00:00Z</published>
    <updated>2023-01-01T00:00:00Z</updated>
    <author><name>Author One</name></author>
    <category term="cs.AI" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/9876.5432v1</id>
    <title>Test Paper 2</title>
    <summary>Abstract 2</summary>
    <published>2023-01-02T00:00:00Z</published>
    <updated>2023-01-02T00:00:00Z</updated>
    <author><name>Author Two</name></author>
    <category term="cs.LG" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
</feed>`
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with custom base URL
	client := NewClient()
	client.baseURL = server.URL

	// Create iterator
	iter := client.NewQuery().SearchQuery("test").Iterator(context.Background())

	// Test iteration
	count := 0
	for paper := range iter.All() {
		if paper == nil {
			t.Error("Expected paper, got nil")
		}
		count++
	}

	if err := iter.Error(); err != nil {
		t.Errorf("Iterator error: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 papers, got %d", count)
	}

	if iter.TotalFetched() != 2 {
		t.Errorf("Expected TotalFetched to be 2, got %d", iter.TotalFetched())
	}
}

func TestIterator_ErrorHandling(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server error"))
	}))
	defer server.Close()

	// Create client with custom base URL
	client := NewClient()
	client.baseURL = server.URL

	// Create iterator
	iter := client.NewQuery().SearchQuery("test").Iterator(context.Background())

	// Test that iteration handles errors correctly
	count := 0
	for paper, err := range iter.AllWithError() {
		if err != nil {
			break // Expected error
		}
		if paper != nil {
			count++
		}
	}

	if count > 0 {
		t.Error("Expected no papers due to server error")
	}

	// Check that error is captured
	if iter.Error() == nil {
		t.Error("Expected iterator to capture error")
	}
}

func TestIterator_EmptyResults(t *testing.T) {
	// Create a mock server that returns empty results
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:itemsPerPage>
</feed>`
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with custom base URL
	client := NewClient()
	client.baseURL = server.URL

	// Create iterator
	iter := client.NewQuery().SearchQuery("test").Iterator(context.Background())

	// Test that iteration handles empty results correctly
	count := 0
	for paper := range iter.All() {
		if paper != nil {
			count++
		}
	}

	if count != 0 {
		t.Error("Expected no papers for empty results")
	}

	if iter.Error() != nil {
		t.Errorf("Expected no error for empty results, got: %v", iter.Error())
	}

	if iter.TotalFetched() != 0 {
		t.Errorf("Expected TotalFetched to be 0, got %d", iter.TotalFetched())
	}
}

func TestIterator_Reset(t *testing.T) {
	// Create a simple iterator
	client := NewClient()
	iter := client.NewQuery().SearchQuery("test").Iterator(context.Background())

	// Simulate some state by manually setting state
	iter.stateManager.state = State{
		Current:      StateReady,
		CurrentPage:  2,
		CurrentIndex: 5,
		TotalFetched: 10,
		Error:        nil,
		Results:      &SearchResults{},
	}

	// Reset the iterator
	iter.Reset()

	// Check that state is reset
	state := iter.stateManager.GetState()
	if state.CurrentIndex != 0 {
		t.Errorf("Expected CurrentIndex to be 0 after reset, got %d", state.CurrentIndex)
	}

	if state.CurrentPage != 0 {
		t.Errorf("Expected CurrentPage to be 0 after reset, got %d", state.CurrentPage)
	}

	if state.TotalFetched != 0 {
		t.Errorf("Expected TotalFetched to be 0 after reset, got %d", state.TotalFetched)
	}

	if state.Results != nil {
		t.Error("Expected Results to be nil after reset")
	}

	if state.Error != nil {
		t.Error("Expected Error to be nil after reset")
	}

	if state.Current != StateInitial {
		t.Errorf("Expected state to be StateInitial after reset, got %s", state.Current.String())
	}
}

func TestIterator_ForEach(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">2</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">2</opensearch:itemsPerPage>
  <entry>
    <id>http://arxiv.org/abs/1234.5678v1</id>
    <title>Test Paper 1</title>
    <summary>Abstract 1</summary>
    <published>2023-01-01T00:00:00Z</published>
    <updated>2023-01-01T00:00:00Z</updated>
    <author><name>Author One</name></author>
    <category term="cs.AI" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/9876.5432v1</id>
    <title>Test Paper 2</title>
    <summary>Abstract 2</summary>
    <published>2023-01-02T00:00:00Z</published>
    <updated>2023-01-02T00:00:00Z</updated>
    <author><name>Author Two</name></author>
    <category term="cs.LG" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
</feed>`
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with custom base URL
	client := NewClient()
	client.baseURL = server.URL

	// Create iterator
	iter := client.NewQuery().SearchQuery("test").Iterator(context.Background())

	// Test ForEach
	count := 0
	err := iter.ForEach(func(paper *Paper) error {
		if paper == nil {
			t.Error("Expected paper, got nil")
		}
		count++
		return nil
	})

	if err != nil {
		t.Errorf("ForEach error: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected ForEach to process 2 papers, got %d", count)
	}
}

func TestIterator_Collect(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">2</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">2</opensearch:itemsPerPage>
  <entry>
    <id>http://arxiv.org/abs/1234.5678v1</id>
    <title>Test Paper 1</title>
    <summary>Abstract 1</summary>
    <published>2023-01-01T00:00:00Z</published>
    <updated>2023-01-01T00:00:00Z</updated>
    <author><name>Author One</name></author>
    <category term="cs.AI" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/9876.5432v1</id>
    <title>Test Paper 2</title>
    <summary>Abstract 2</summary>
    <published>2023-01-02T00:00:00Z</published>
    <updated>2023-01-02T00:00:00Z</updated>
    <author><name>Author Two</name></author>
    <category term="cs.LG" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
</feed>`
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with custom base URL
	client := NewClient()
	client.baseURL = server.URL

	// Create iterator
	iter := client.NewQuery().SearchQuery("test").Iterator(context.Background())

	// Test Collect
	papers, err := iter.Collect()
	if err != nil {
		t.Errorf("Collect error: %v", err)
	}

	if len(papers) != 2 {
		t.Errorf("Expected Collect to return 2 papers, got %d", len(papers))
	}

	// Verify paper data
	if papers[0].Title != "Test Paper 1" {
		t.Errorf("Expected first paper title 'Test Paper 1', got '%s'", papers[0].Title)
	}

	if papers[1].Title != "Test Paper 2" {
		t.Errorf("Expected second paper title 'Test Paper 2', got '%s'", papers[1].Title)
	}
}

func TestIterator_CollectN(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">3</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">3</opensearch:itemsPerPage>
  <entry>
    <id>http://arxiv.org/abs/1234.5678v1</id>
    <title>Test Paper 1</title>
    <summary>Abstract 1</summary>
    <published>2023-01-01T00:00:00Z</published>
    <updated>2023-01-01T00:00:00Z</updated>
    <author><name>Author One</name></author>
    <category term="cs.AI" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/9876.5432v1</id>
    <title>Test Paper 2</title>
    <summary>Abstract 2</summary>
    <published>2023-01-02T00:00:00Z</published>
    <updated>2023-01-02T00:00:00Z</updated>
    <author><name>Author Two</name></author>
    <category term="cs.LG" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/1111.2222v1</id>
    <title>Test Paper 3</title>
    <summary>Abstract 3</summary>
    <published>2023-01-03T00:00:00Z</published>
    <updated>2023-01-03T00:00:00Z</updated>
    <author><name>Author Three</name></author>
    <category term="cs.CV" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
</feed>`
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with custom base URL
	client := NewClient()
	client.baseURL = server.URL

	// Create iterator
	iter := client.NewQuery().SearchQuery("test").Iterator(context.Background())

	// Test CollectN with limit of 2
	papers, err := iter.CollectN(2)
	if err != nil {
		t.Errorf("CollectN error: %v", err)
	}

	if len(papers) != 2 {
		t.Errorf("Expected CollectN(2) to return 2 papers, got %d", len(papers))
	}

	// Verify paper data
	if papers[0].Title != "Test Paper 1" {
		t.Errorf("Expected first paper title 'Test Paper 1', got '%s'", papers[0].Title)
	}

	if papers[1].Title != "Test Paper 2" {
		t.Errorf("Expected second paper title 'Test Paper 2', got '%s'", papers[1].Title)
	}
}

func TestIterator_WithContext(t *testing.T) {
	client := NewClient()
	originalCtx := context.Background()
	newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := &Query{SearchQuery: "test", MaxResults: 10}
	iter := NewIterator(client, query, originalCtx)

	// Test WithContext
	newIter := iter.WithContext(newCtx)

	// Check that new iterator is different from original
	if newIter == iter {
		t.Error("Expected WithContext to return a new iterator")
	}

	// Check that new iterator has reset state
	newState := newIter.stateManager.GetState()
	if newState.Current != StateInitial {
		t.Errorf("Expected new iterator to have initial state, got %s", newState.Current.String())
	}

	// Check that original iterator's state is unchanged
	originalState := iter.stateManager.GetState()
	if originalState.Current != StateInitial {
		t.Errorf("Expected original iterator to maintain its state, got %s", originalState.Current.String())
	}
}

func TestIterator_LimitFunctionality(t *testing.T) {
	// Create a mock server that returns many papers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">1000</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">5</opensearch:itemsPerPage>
  <entry>
    <id>http://arxiv.org/abs/1234.5678v1</id>
    <title>Test Paper 1</title>
    <summary>Abstract 1</summary>
    <published>2023-01-01T00:00:00Z</published>
    <updated>2023-01-01T00:00:00Z</updated>
    <author><name>Author One</name></author>
    <category term="cs.AI" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/9876.5432v1</id>
    <title>Test Paper 2</title>
    <summary>Abstract 2</summary>
    <published>2023-01-02T00:00:00Z</published>
    <updated>2023-01-02T00:00:00Z</updated>
    <author><name>Author Two</name></author>
    <category term="cs.LG" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/1111.2222v1</id>
    <title>Test Paper 3</title>
    <summary>Abstract 3</summary>
    <published>2023-01-03T00:00:00Z</published>
    <updated>2023-01-03T00:00:00Z</updated>
    <author><name>Author Three</name></author>
    <category term="cs.CV" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/3333.4444v1</id>
    <title>Test Paper 4</title>
    <summary>Abstract 4</summary>
    <published>2023-01-04T00:00:00Z</published>
    <updated>2023-01-04T00:00:00Z</updated>
    <author><name>Author Four</name></author>
    <category term="cs.AI" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/5555.6666v1</id>
    <title>Test Paper 5</title>
    <summary>Abstract 5</summary>
    <published>2023-01-05T00:00:00Z</published>
    <updated>2023-01-05T00:00:00Z</updated>
    <author><name>Author Five</name></author>
    <category term="cs.LG" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
</feed>`
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with custom base URL
	client := NewClient()
	client.baseURL = server.URL

	// Test with limit of 3
	iter := client.NewQuery().
		SearchQuery("test").
		Limit(3).
		Iterator(context.Background())

	// Test iteration with limit
	count := 0
	for paper := range iter.All() {
		if paper == nil {
			t.Error("Expected paper, got nil")
		}
		count++
	}

	if err := iter.Error(); err != nil {
		t.Errorf("Iterator error: %v", err)
	}

	// Should stop at limit of 3, not fetch all 5 available papers
	if count != 3 {
		t.Errorf("Expected exactly 3 papers due to limit, got %d", count)
	}

	if iter.TotalFetched() != 3 {
		t.Errorf("Expected TotalFetched to be 3, got %d", iter.TotalFetched())
	}
}

func TestIterator_InvalidQuery(t *testing.T) {
	client := NewClient()

	// Create iterator with invalid query (this should set an error)
	iter := client.NewQuery().Iterator(context.Background()) // No search query or ID list

	// Test that iteration handles invalid query correctly
	count := 0
	for paper := range iter.All() {
		if paper != nil {
			count++
		}
	}

	if count > 0 {
		t.Error("Expected no papers for invalid query")
	}

	// Check that error is set
	if iter.Error() == nil {
		t.Error("Expected iterator to have error for invalid query")
	}
}

// TestIterator_Go123Iteration tests the new Go 1.23+ iter.Seq pattern
func TestIterator_Go123Iteration(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">3</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">3</opensearch:itemsPerPage>
  <entry>
    <id>http://arxiv.org/abs/1234.5678v1</id>
    <title>Quantum Computing Paper</title>
    <summary>Abstract about quantum computing</summary>
    <published>2023-01-01T00:00:00Z</published>
    <updated>2023-01-01T00:00:00Z</updated>
    <author><name>Alice Quantum</name></author>
    <category term="quant-ph" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/9876.5432v1</id>
    <title>Machine Learning Paper</title>
    <summary>Abstract about ML</summary>
    <published>2023-01-02T00:00:00Z</published>
    <updated>2023-01-02T00:00:00Z</updated>
    <author><name>Bob ML</name></author>
    <category term="cs.LG" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/1111.2222v1</id>
    <title>Computer Vision Paper</title>
    <summary>Abstract about CV</summary>
    <published>2023-01-03T00:00:00Z</published>
    <updated>2023-01-03T00:00:00Z</updated>
    <author><name>Carol CV</name></author>
    <category term="cs.CV" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
</feed>`
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with custom base URL
	client := NewClient()
	client.baseURL = server.URL

	// Create iterator
	iter := client.NewQuery().SearchQuery("test").Iterator(context.Background())

	// Test range-over-func pattern
	papers := make([]*Paper, 0)
	for paper := range iter.All() {
		papers = append(papers, paper)
	}

	// Verify results
	if len(papers) != 3 {
		t.Errorf("Expected 3 papers, got %d", len(papers))
	}

	expectedTitles := []string{
		"Quantum Computing Paper",
		"Machine Learning Paper",
		"Computer Vision Paper",
	}

	for i, paper := range papers {
		if paper.Title != expectedTitles[i] {
			t.Errorf("Expected paper %d title '%s', got '%s'", i, expectedTitles[i], paper.Title)
		}
	}

	// Test that TotalFetched is updated correctly
	if iter.TotalFetched() != 3 {
		t.Errorf("Expected TotalFetched to be 3, got %d", iter.TotalFetched())
	}
}

// TestIterator_AllWithError tests the iter.Seq2 pattern with error handling
func TestIterator_AllWithError(t *testing.T) {
	// Create a mock server that returns an error after first page
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			// First request succeeds
			response := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">100</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">1</opensearch:itemsPerPage>
  <entry>
    <id>http://arxiv.org/abs/1234.5678v1</id>
    <title>First Paper</title>
    <summary>Abstract 1</summary>
    <published>2023-01-01T00:00:00Z</published>
    <updated>2023-01-01T00:00:00Z</updated>
    <author><name>Author One</name></author>
    <category term="cs.AI" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
</feed>`
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		} else {
			// Subsequent requests fail
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server error"))
		}
	}))
	defer server.Close()

	// Create client with custom base URL
	client := NewClient()
	client.baseURL = server.URL

	// Create iterator with small batch size to trigger multiple requests
	iter := client.NewQuery().SearchQuery("test").MaxResults(1).Iterator(context.Background())

	// Test AllWithError pattern
	papers := make([]*Paper, 0)
	var lastError error
	for paper, err := range iter.AllWithError() {
		if err != nil {
			lastError = err
			break
		}
		if paper != nil {
			papers = append(papers, paper)
		}
	}

	// Should have gotten one paper before the error
	if len(papers) != 1 {
		t.Errorf("Expected 1 paper before error, got %d", len(papers))
	}

	if papers[0].Title != "First Paper" {
		t.Errorf("Expected first paper title 'First Paper', got '%s'", papers[0].Title)
	}

	// Should have captured the error
	if lastError == nil {
		t.Error("Expected to capture error from second request")
	}
}

// TestIterator_Values tests the Values() alias
func TestIterator_Values(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">1</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">1</opensearch:itemsPerPage>
  <entry>
    <id>http://arxiv.org/abs/1234.5678v1</id>
    <title>Test Paper</title>
    <summary>Test Abstract</summary>
    <published>2023-01-01T00:00:00Z</published>
    <updated>2023-01-01T00:00:00Z</updated>
    <author><name>Test Author</name></author>
    <category term="cs.AI" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
</feed>`
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with custom base URL
	client := NewClient()
	client.baseURL = server.URL

	// Create iterator
	iter := client.NewQuery().SearchQuery("test").Iterator(context.Background())

	// Test Values() alias
	count := 0
	for paper := range iter.Values() {
		if paper == nil {
			t.Error("Expected paper, got nil")
			break
		}
		if paper.Title != "Test Paper" {
			t.Errorf("Expected title 'Test Paper', got '%s'", paper.Title)
		}
		count++
	}

	if count != 1 {
		t.Errorf("Expected 1 paper, got %d", count)
	}
}

// TestIterator_HelperFunctions tests the package-level helper functions
func TestIterator_HelperFunctions(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">5</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">5</opensearch:itemsPerPage>
  <entry>
    <id>http://arxiv.org/abs/1001.0001v1</id>
    <title>Paper 1</title>
    <summary>Abstract 1</summary>
    <published>2023-01-01T00:00:00Z</published>
    <updated>2023-01-01T00:00:00Z</updated>
    <author><name>Author 1</name></author>
    <category term="cs.AI" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/1002.0002v1</id>
    <title>Paper 2</title>
    <summary>Abstract 2</summary>
    <published>2023-01-02T00:00:00Z</published>
    <updated>2023-01-02T00:00:00Z</updated>
    <author><name>Author 2</name></author>
    <category term="cs.LG" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/1003.0003v1</id>
    <title>Paper 3</title>
    <summary>Abstract 3</summary>
    <published>2023-01-03T00:00:00Z</published>
    <updated>2023-01-03T00:00:00Z</updated>
    <author><name>Author 3</name></author>
    <category term="cs.CV" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/1004.0004v1</id>
    <title>Paper 4</title>
    <summary>Abstract 4</summary>
    <published>2023-01-04T00:00:00Z</published>
    <updated>2023-01-04T00:00:00Z</updated>
    <author><name>Author 4</name></author>
    <category term="cs.NI" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/1005.0005v1</id>
    <title>Paper 5</title>
    <summary>Abstract 5</summary>
    <published>2023-01-05T00:00:00Z</published>
    <updated>2023-01-05T00:00:00Z</updated>
    <author><name>Author 5</name></author>
    <category term="cs.CL" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
</feed>`
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with custom base URL
	client := NewClient()
	client.baseURL = server.URL

	// Create iterator
	iter := client.NewQuery().SearchQuery("test").Iterator(context.Background())

	// Test CollectSeq
	papers := CollectSeq(iter.All())
	if len(papers) != 5 {
		t.Errorf("Expected CollectSeq to return 5 papers, got %d", len(papers))
	}

	// Reset iterator for next test
	iter.Reset()

	// Test CollectNSeq
	papers = CollectNSeq(iter.All(), 3)
	if len(papers) != 3 {
		t.Errorf("Expected CollectNSeq to return 3 papers, got %d", len(papers))
	}

	// Reset iterator for next test
	iter.Reset()

	// Test TakeSeq
	taken := CollectSeq(TakeSeq(iter.All(), 2))
	if len(taken) != 2 {
		t.Errorf("Expected TakeSeq to return 2 papers, got %d", len(taken))
	}

	// Reset iterator for next test
	iter.Reset()

	// Test FilterSeq - filter papers with "AI" or "LG" categories
	filtered := CollectSeq(FilterSeq(iter.All(), func(paper *Paper) bool {
		for _, cat := range paper.Categories {
			if cat == "cs.AI" || cat == "cs.LG" {
				return true
			}
		}
		return false
	}))
	if len(filtered) != 2 {
		t.Errorf("Expected FilterSeq to return 2 papers (cs.AI and cs.LG), got %d", len(filtered))
	}

	// Reset iterator for next test
	iter.Reset()

	// Test ForEachSeq
	count := 0
	err := ForEachSeq(iter.All(), func(paper *Paper) error {
		count++
		return nil
	})
	if err != nil {
		t.Errorf("ForEachSeq error: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected ForEachSeq to process 5 papers, got %d", count)
	}
}

// TestIterator_EarlyBreak tests that early breaking from iteration works correctly
func TestIterator_EarlyBreak(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <opensearch:totalResults xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">100</opensearch:totalResults>
  <opensearch:startIndex xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">0</opensearch:startIndex>
  <opensearch:itemsPerPage xmlns:opensearch="http://a9.com/-/spec/opensearch/1.1/">10</opensearch:itemsPerPage>
  <entry>
    <id>http://arxiv.org/abs/1234.5678v1</id>
    <title>Paper 1</title>
    <summary>Abstract 1</summary>
    <published>2023-01-01T00:00:00Z</published>
    <updated>2023-01-01T00:00:00Z</updated>
    <author><name>Author 1</name></author>
    <category term="cs.AI" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/2345.6789v1</id>
    <title>Paper 2</title>
    <summary>Abstract 2</summary>
    <published>2023-01-02T00:00:00Z</published>
    <updated>2023-01-02T00:00:00Z</updated>
    <author><name>Author 2</name></author>
    <category term="cs.LG" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/3456.7890v1</id>
    <title>Paper 3</title>
    <summary>Abstract 3</summary>
    <published>2023-01-03T00:00:00Z</published>
    <updated>2023-01-03T00:00:00Z</updated>
    <author><name>Author 3</name></author>
    <category term="cs.CV" scheme="http://arxiv.org/schemas/atom"/>
  </entry>
</feed>`
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with custom base URL
	client := NewClient()
	client.baseURL = server.URL

	// Create iterator
	iter := client.NewQuery().SearchQuery("test").Iterator(context.Background())

	// Test early break
	count := 0
	for paper := range iter.All() {
		count++
		if paper.Title == "Paper 2" {
			break // Early break
		}
	}

	if count != 2 {
		t.Errorf("Expected to process 2 papers before break, got %d", count)
	}

	// Verify that TotalFetched reflects only the papers we actually consumed
	if iter.TotalFetched() != 2 {
		t.Errorf("Expected TotalFetched to be 2 after early break, got %d", iter.TotalFetched())
	}
}
