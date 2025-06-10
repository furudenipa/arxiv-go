package arxiv

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// QueryBuilder provides a fluent interface for building arXiv queries
type QueryBuilder struct {
	client      *Client
	searchTerms []string
	categories  []Category
	authors     []string
	titles      []string
	abstracts   []string
	dateFrom    *time.Time
	dateTo      *time.Time
	sortBy      SortCriterion
	sortOrder   SortOrder
	maxResults  int
	limit       int
	start       int
	idList      []string
	errors      []error
}

// SearchQuery adds a general search term
func (qb *QueryBuilder) SearchQuery(query string) *QueryBuilder {
	if query != "" {
		qb.searchTerms = append(qb.searchTerms, query)
	}
	return qb
}

// Category adds a category filter
func (qb *QueryBuilder) Category(cat Category) *QueryBuilder {
	if cat != "" {
		qb.categories = append(qb.categories, cat)
	}
	return qb
}

// Categories adds multiple category filters
func (qb *QueryBuilder) Categories(cats ...Category) *QueryBuilder {
	for _, cat := range cats {
		if cat != "" {
			qb.categories = append(qb.categories, cat)
		}
	}
	return qb
}

// Author adds an author filter
func (qb *QueryBuilder) Author(author string) *QueryBuilder {
	if author != "" {
		qb.authors = append(qb.authors, author)
	}
	return qb
}

// Authors adds multiple author filters
func (qb *QueryBuilder) Authors(authors ...string) *QueryBuilder {
	for _, author := range authors {
		if author != "" {
			qb.authors = append(qb.authors, author)
		}
	}
	return qb
}

// Title adds a title filter
func (qb *QueryBuilder) Title(title string) *QueryBuilder {
	if title != "" {
		qb.titles = append(qb.titles, title)
	}
	return qb
}

// Abstract adds an abstract filter
func (qb *QueryBuilder) Abstract(abstract string) *QueryBuilder {
	if abstract != "" {
		qb.abstracts = append(qb.abstracts, abstract)
	}
	return qb
}

// DateRange sets the date range filter
func (qb *QueryBuilder) DateRange(from, to time.Time) *QueryBuilder {
	qb.dateFrom = &from
	qb.dateTo = &to
	return qb
}

// DateFrom sets the start date filter
func (qb *QueryBuilder) DateFrom(from time.Time) *QueryBuilder {
	qb.dateFrom = &from
	return qb
}

// DateTo sets the end date filter
func (qb *QueryBuilder) DateTo(to time.Time) *QueryBuilder {
	qb.dateTo = &to
	return qb
}

// SortBy sets the sort criteria and order
func (qb *QueryBuilder) SortBy(criterion SortCriterion, order SortOrder) *QueryBuilder {
	qb.sortBy = criterion
	qb.sortOrder = order
	return qb
}

// MaxResults sets the maximum number of results per API request
func (qb *QueryBuilder) MaxResults(max int) *QueryBuilder {
	if max > 0 {
		qb.maxResults = max
	} else {
		qb.errors = append(qb.errors, fmt.Errorf("max results must be positive, got %d", max))
	}
	return qb
}

// Limit sets the maximum total number of results to fetch across all requests (0 = unlimited)
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	if limit >= 0 {
		qb.limit = limit
	} else {
		qb.errors = append(qb.errors, fmt.Errorf("limit must be non-negative, got %d", limit))
	}
	return qb
}

// Start sets the starting index for pagination
func (qb *QueryBuilder) Start(start int) *QueryBuilder {
	if start >= 0 {
		qb.start = start
	} else {
		qb.errors = append(qb.errors, fmt.Errorf("start index must be non-negative, got %d", start))
	}
	return qb
}

// IDList sets the arXiv ID list (alternative to search query)
func (qb *QueryBuilder) IDList(ids ...string) *QueryBuilder {
	qb.idList = append(qb.idList, ids...)
	return qb
}

// AND adds an AND operation to the search query
func (qb *QueryBuilder) AND() *QueryBuilder {
	// This is a marker for complex query building
	qb.searchTerms = append(qb.searchTerms, "AND")
	return qb
}

// OR adds an OR operation to the search query
func (qb *QueryBuilder) OR() *QueryBuilder {
	// This is a marker for complex query building
	qb.searchTerms = append(qb.searchTerms, "OR")
	return qb
}

// ANDNOT adds an ANDNOT operation to the search query
func (qb *QueryBuilder) ANDNOT() *QueryBuilder {
	// This is a marker for complex query building
	qb.searchTerms = append(qb.searchTerms, "ANDNOT")
	return qb
}

// buildSearchQuery constructs the final search query string
func (qb *QueryBuilder) buildSearchQuery() string {
	var queryParts []string

	// Add search terms
	if len(qb.searchTerms) > 0 {
		// Handle complex queries with operators
		searchQuery := strings.Join(qb.searchTerms, " ")
		if searchQuery != "" {
			queryParts = append(queryParts, fmt.Sprintf("(%s)", searchQuery))
		}
	}

	// Add category filters
	if len(qb.categories) > 0 {
		var catQueries []string
		for _, cat := range qb.categories {
			catQueries = append(catQueries, fmt.Sprintf("cat:%s", string(cat)))
		}
		if len(catQueries) == 1 {
			queryParts = append(queryParts, catQueries[0])
		} else {
			queryParts = append(queryParts, fmt.Sprintf("(%s)", strings.Join(catQueries, " OR ")))
		}
	}

	// Add author filters
	if len(qb.authors) > 0 {
		var authQueries []string
		for _, author := range qb.authors {
			authQueries = append(authQueries, fmt.Sprintf("au:%s", author))
		}
		if len(authQueries) == 1 {
			queryParts = append(queryParts, authQueries[0])
		} else {
			queryParts = append(queryParts, fmt.Sprintf("(%s)", strings.Join(authQueries, " OR ")))
		}
	}

	// Add title filters
	if len(qb.titles) > 0 {
		var titleQueries []string
		for _, title := range qb.titles {
			titleQueries = append(titleQueries, fmt.Sprintf("ti:%s", title))
		}
		if len(titleQueries) == 1 {
			queryParts = append(queryParts, titleQueries[0])
		} else {
			queryParts = append(queryParts, fmt.Sprintf("(%s)", strings.Join(titleQueries, " OR ")))
		}
	}

	// Add abstract filters
	if len(qb.abstracts) > 0 {
		var absQueries []string
		for _, abstract := range qb.abstracts {
			absQueries = append(absQueries, fmt.Sprintf("abs:%s", abstract))
		}
		if len(absQueries) == 1 {
			queryParts = append(queryParts, absQueries[0])
		} else {
			queryParts = append(queryParts, fmt.Sprintf("(%s)", strings.Join(absQueries, " OR ")))
		}
	}

	return strings.Join(queryParts, " AND ")
}

// buildQuery constructs the Query object
func (qb *QueryBuilder) buildQuery() (*Query, error) {
	// Check for accumulated errors
	if len(qb.errors) > 0 {
		return nil, qb.errors[0] // Return the first error
	}

	query := &Query{
		Start:             qb.start,
		MaxResults:        qb.maxResults,
		Limit:             qb.limit,
		SortBy:            string(qb.sortBy),
		SortOrder:         string(qb.sortOrder),
		SubmittedDateFrom: qb.dateFrom,
		SubmittedDateTo:   qb.dateTo,
	}

	// Set ID list or search query
	if len(qb.idList) > 0 {
		query.IDList = qb.idList
	} else {
		searchQuery := qb.buildSearchQuery()
		if searchQuery == "" && len(qb.idList) == 0 {
			return nil, NewAPIError(ErrorTypeInvalidQuery, "either search query or ID list must be provided", nil)
		}
		query.SearchQuery = searchQuery
	}

	return query, nil
}

// Execute executes the query and returns the results using the unified Search method
func (qb *QueryBuilder) Execute(ctx context.Context) (*SearchResults, error) {
	query, err := qb.buildQuery()
	if err != nil {
		return nil, err
	}

	return qb.client.Search(ctx, query)
}

// Iterator returns an iterator for paginated results
func (qb *QueryBuilder) Iterator(ctx context.Context) *Iterator {
	query, err := qb.buildQuery()
	if err != nil {
		// Return an iterator in error state
		iter := NewIterator(qb.client, query, ctx)
		iter.stateManager.Transition(FetchAction{Results: nil, Error: err})
		return iter
	}

	return NewIterator(qb.client, query, ctx)
}

// Validate checks if the query builder configuration is valid
func (qb *QueryBuilder) Validate() error {
	if len(qb.errors) > 0 {
		return qb.errors[0]
	}

	if len(qb.idList) == 0 && qb.buildSearchQuery() == "" {
		return NewAPIError(ErrorTypeInvalidQuery, "either search query or ID list must be provided", nil)
	}

	if qb.maxResults <= 0 {
		return NewAPIError(ErrorTypeInvalidQuery, "max results must be positive", nil)
	}

	if qb.start < 0 {
		return NewAPIError(ErrorTypeInvalidQuery, "start index must be non-negative", nil)
	}

	return nil
}
