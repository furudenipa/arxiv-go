package arxiv

import (
	"testing"
	"time"
)

func TestQueryBuilder_SearchQuery(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().SearchQuery("quantum computing")

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	if query.SearchQuery != "(quantum computing)" {
		t.Errorf("Expected search query '(quantum computing)', got '%s'", query.SearchQuery)
	}
}

func TestQueryBuilder_Category(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().Category(CategoryCSAI)

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	if query.SearchQuery != "cat:cs.AI" {
		t.Errorf("Expected search query 'cat:cs.AI', got '%s'", query.SearchQuery)
	}
}

func TestQueryBuilder_Categories(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().Categories(CategoryCSAI, CategoryCSLG)

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	expected := "(cat:cs.AI OR cat:cs.LG)"
	if query.SearchQuery != expected {
		t.Errorf("Expected search query '%s', got '%s'", expected, query.SearchQuery)
	}
}

func TestQueryBuilder_Author(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().Author("Einstein")

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	if query.SearchQuery != "au:Einstein" {
		t.Errorf("Expected search query 'au:Einstein', got '%s'", query.SearchQuery)
	}
}

func TestQueryBuilder_Authors(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().Authors("Einstein", "Bohr")

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	expected := "(au:Einstein OR au:Bohr)"
	if query.SearchQuery != expected {
		t.Errorf("Expected search query '%s', got '%s'", expected, query.SearchQuery)
	}
}

func TestQueryBuilder_Title(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().Title("relativity")

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	if query.SearchQuery != "ti:relativity" {
		t.Errorf("Expected search query 'ti:relativity', got '%s'", query.SearchQuery)
	}
}

func TestQueryBuilder_Abstract(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().Abstract("machine learning")

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	if query.SearchQuery != "abs:machine learning" {
		t.Errorf("Expected search query 'abs:machine learning', got '%s'", query.SearchQuery)
	}
}

func TestQueryBuilder_DateRange(t *testing.T) {
	client := NewClient()
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

	qb := client.NewQuery().
		SearchQuery("quantum computing").
		DateRange(from, to)

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	if query.SubmittedDateFrom == nil || !query.SubmittedDateFrom.Equal(from) {
		t.Errorf("Expected SubmittedDateFrom to be %v, got %v", from, query.SubmittedDateFrom)
	}

	if query.SubmittedDateTo == nil || !query.SubmittedDateTo.Equal(to) {
		t.Errorf("Expected SubmittedDateTo to be %v, got %v", to, query.SubmittedDateTo)
	}
}

func TestQueryBuilder_SortBy(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().
		SearchQuery("quantum computing").
		SortBy(SortBySubmittedDate, SortOrderAscending)

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	if query.SortBy != string(SortBySubmittedDate) {
		t.Errorf("Expected SortBy to be '%s', got '%s'", SortBySubmittedDate, query.SortBy)
	}

	if query.SortOrder != string(SortOrderAscending) {
		t.Errorf("Expected SortOrder to be '%s', got '%s'", SortOrderAscending, query.SortOrder)
	}
}

func TestQueryBuilder_MaxResults(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().
		SearchQuery("quantum computing").
		MaxResults(50)

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	if query.MaxResults != 50 {
		t.Errorf("Expected MaxResults to be 50, got %d", query.MaxResults)
	}
}

func TestQueryBuilder_Start(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().
		SearchQuery("quantum computing").
		Start(10)

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	if query.Start != 10 {
		t.Errorf("Expected Start to be 10, got %d", query.Start)
	}
}

func TestQueryBuilder_IDList(t *testing.T) {
	client := NewClient()
	ids := []string{"1234.5678", "9876.5432"}
	qb := client.NewQuery().IDList(ids...)

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	if len(query.IDList) != 2 {
		t.Errorf("Expected IDList length to be 2, got %d", len(query.IDList))
	}

	if query.IDList[0] != "1234.5678" || query.IDList[1] != "9876.5432" {
		t.Errorf("Expected IDList to be %v, got %v", ids, query.IDList)
	}
}

func TestQueryBuilder_ComplexQuery(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().
		SearchQuery("quantum computing").
		Category(CategoryCSAI).
		Author("Einstein").
		Title("relativity")

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	expected := "(quantum computing) AND cat:cs.AI AND au:Einstein AND ti:relativity"
	if query.SearchQuery != expected {
		t.Errorf("Expected search query '%s', got '%s'", expected, query.SearchQuery)
	}
}

func TestQueryBuilder_ComplexQueryWithOperators(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().
		SearchQuery("quantum").
		AND().
		SearchQuery("computing").
		OR().
		SearchQuery("machine learning")

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	expected := "(quantum AND computing OR machine learning)"
	if query.SearchQuery != expected {
		t.Errorf("Expected search query '%s', got '%s'", expected, query.SearchQuery)
	}
}

func TestQueryBuilder_Validation(t *testing.T) {
	client := NewClient()

	// Test empty query
	qb := client.NewQuery()
	err := qb.Validate()
	if err == nil {
		t.Error("Expected validation error for empty query")
	}

	// Test negative max results
	qb = client.NewQuery().SearchQuery("test").MaxResults(-1)
	err = qb.Validate()
	if err == nil {
		t.Error("Expected validation error for negative max results")
	}

	// Test negative start
	qb = client.NewQuery().SearchQuery("test").Start(-1)
	err = qb.Validate()
	if err == nil {
		t.Error("Expected validation error for negative start")
	}

	// Test valid query
	qb = client.NewQuery().SearchQuery("test")
	err = qb.Validate()
	if err != nil {
		t.Errorf("Expected no validation error for valid query, got: %v", err)
	}
}

func TestQueryBuilder_ErrorAccumulation(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().
		MaxResults(-1). // This should add an error
		Start(-1)       // This should also add an error

	_, err := qb.buildQuery()
	if err == nil {
		t.Error("Expected error from buildQuery due to accumulated errors")
	}
}

func TestQueryBuilder_EmptyValues(t *testing.T) {
	client := NewClient()
	qb := client.NewQuery().
		SearchQuery("").                 // Empty string should be ignored
		Category("").                    // Empty category should be ignored
		Author("").                      // Empty author should be ignored
		SearchQuery("quantum computing") // This should be used

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	if query.SearchQuery != "(quantum computing)" {
		t.Errorf("Expected search query '(quantum computing)', got '%s'", query.SearchQuery)
	}
}

func TestQueryBuilder_ChainedCalls(t *testing.T) {
	client := NewClient()
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

	// Test full chain of method calls
	qb := client.NewQuery().
		SearchQuery("quantum computing").
		Category(CategoryCSAI).
		Author("Einstein").
		DateRange(from, to).
		SortBy(SortBySubmittedDate, SortOrderDescending).
		MaxResults(50).
		Start(10)

	query, err := qb.buildQuery()
	if err != nil {
		t.Fatalf("buildQuery failed: %v", err)
	}

	// Verify all parameters are set correctly
	expectedSearchQuery := "(quantum computing) AND cat:cs.AI AND au:Einstein"
	if query.SearchQuery != expectedSearchQuery {
		t.Errorf("Expected search query '%s', got '%s'", expectedSearchQuery, query.SearchQuery)
	}

	if query.MaxResults != 50 {
		t.Errorf("Expected MaxResults to be 50, got %d", query.MaxResults)
	}

	if query.Start != 10 {
		t.Errorf("Expected Start to be 10, got %d", query.Start)
	}

	if query.SortBy != string(SortBySubmittedDate) {
		t.Errorf("Expected SortBy to be '%s', got '%s'", SortBySubmittedDate, query.SortBy)
	}

	if query.SortOrder != string(SortOrderDescending) {
		t.Errorf("Expected SortOrder to be '%s', got '%s'", SortOrderDescending, query.SortOrder)
	}
}
