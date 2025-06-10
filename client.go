package arxiv

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// ArXiv API base URL
	baseURL = "https://export.arxiv.org/api/query"

	// Default values
	defaultMaxResults = 500
	defaultLimit      = 0
	defaultSortBy     = "relevance"
	defaultSortOrder  = "descending"

	defaultRetryAttempts = 3
	defaultRetryDelay    = 1 * time.Second
	defaultRateLimit     = 1000 * time.Millisecond
	defaultUserAgent     = "arxiv-go/1.0"
	defaultTimeout       = 30 * time.Second
)

// ClientOptions represents configuration options for the arXiv client
type ClientOptions struct {

	// RetryAttempts specifies the number of retry attempts for failed requests
	RetryAttempts int

	// RetryDelay specifies the initial delay between retry attempts
	RetryDelay time.Duration

	// RateLimit specifies the minimum delay between requests
	RateLimit time.Duration

	// UserAgent specifies the User-Agent header to use
	UserAgent string

	// Timeout specifies the request timeout
	Timeout time.Duration
}

// DefaultClientOptions returns the default client options
func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		RetryAttempts: defaultRetryAttempts,
		RetryDelay:    defaultRetryDelay,
		RateLimit:     defaultRateLimit,
		UserAgent:     defaultUserAgent,
		Timeout:       defaultTimeout,
	}
}

// Client represents an arXiv API client
type Client struct {
	httpClient  *http.Client
	baseURL     string
	options     ClientOptions
	lastRequest time.Time

	rlMu sync.Mutex // Mutex for rate limiting
}

// NewClient creates a new arXiv API client
func NewClient() *Client {
	return NewClientWithOptions(DefaultClientOptions())
}

// NewClientWithHTTPClient creates a new arXiv API client with custom HTTP client
func NewClientWithHTTPClient(httpClient *http.Client) *Client {
	opts := DefaultClientOptions()
	return &Client{
		httpClient:  httpClient,
		baseURL:     baseURL,
		options:     opts,
		lastRequest: time.Time{},
	}
}

// NewClientWithOptions creates a new arXiv API client with custom options
// TODO: do not use magic values for defaults, use constants or config
func NewClientWithOptions(opts ClientOptions) *Client {
	// Set defaults for zero values
	if opts.RetryAttempts == 0 {
		opts.RetryAttempts = defaultRetryAttempts
	}
	if opts.RetryDelay == 0 {
		opts.RetryDelay = defaultRetryDelay
	}
	if opts.RateLimit == 0 {
		opts.RateLimit = defaultRateLimit
	}
	if opts.UserAgent == "" {
		opts.UserAgent = defaultUserAgent
	}
	if opts.Timeout == 0 {
		opts.Timeout = defaultTimeout
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: opts.Timeout,
		},
		baseURL:     baseURL,
		options:     opts,
		lastRequest: time.Time{},
	}
}

// Search searches for papers using the arXiv API with retry and rate limiting
func (c *Client) Search(ctx context.Context, query *Query) (*SearchResults, error) {
	if query == nil {
		return nil, NewAPIError(ErrorTypeInvalidQuery, "query cannot be nil", nil)
	}

	var result *SearchResults
	err := c.retryWithBackoff(ctx, func() error {
		// Build URL
		params := c.buildQueryParams(query)
		reqURL := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())

		// Create HTTP request
		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return NewAPIError(ErrorTypeNetwork, "failed to create request", err)
		}

		userAgent := c.options.UserAgent
		if userAgent == "" {
			userAgent = defaultUserAgent
		}
		req.Header.Set("User-Agent", userAgent)

		// Apply rate limiting and update last request time
		err = c.applyRateLimit(ctx)
		if err != nil {
			return err
		}

		// Make request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return NewAPIError(ErrorTypeNetwork, "failed to make request", err)
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			// Continue
		case http.StatusTooManyRequests, http.StatusServiceUnavailable:
			return NewAPIError(ErrorTypeRateLimit, "rate limit exceeded", fmt.Errorf("rate limit exceeded, status %d", resp.StatusCode))
		default:
			return NewAPIError(ErrorTypeNetwork, "API error", fmt.Errorf("unexpected status code %d", resp.StatusCode))
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return NewAPIError(ErrorTypeNetwork, "failed to read response body", err)
		}

		// Parse XML response
		// TODO: implement ErrorTypeNoEntry retry
		parsedResult, err := c.parseSearchResponse(body)
		if err != nil {
			return NewAPIError(ErrorTypeParsing, "failed to parse response", err)
		}

		result = parsedResult
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetByID retrieves a paper by its arXiv ID with retry logic
func (c *Client) GetByID(ctx context.Context, id string) (*Paper, error) {
	if id == "" {
		return nil, NewAPIError(ErrorTypeInvalidQuery, "id cannot be empty", nil)
	}

	query := &Query{
		IDList:     []string{id},
		MaxResults: 1,
	}

	results, err := c.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(results.Papers) == 0 {
		return nil, NewAPIError(ErrorTypeNotFound, fmt.Sprintf("paper with ID %s not found", id), nil)
	}

	return &results.Papers[0], nil
}

// NewQuery creates a new QueryBuilder instance
func (c *Client) NewQuery() *QueryBuilder {
	return &QueryBuilder{
		client:     c,
		maxResults: defaultMaxResults,
		limit:      defaultLimit,
		sortBy:     SortByRelevance,
		sortOrder:  SortOrderDescending,
	}
}

// Iterator returns an iterator for paginated results
func (c *Client) Iterator(ctx context.Context, query *Query) *Iterator {
	return NewIterator(c, query, ctx)
}

// WithRetryConfig returns a new client with updated retry configuration
// func (c *Client) WithRetryConfig(retryAttempts int, retryDelay time.Duration) *Client {
// 	newClient := *c
// 	newClient.options.RetryAttempts = retryAttempts
// 	newClient.options.RetryDelay = retryDelay
// 	return &newClient
// }

// // WithRateLimitConfig returns a new client with updated rate limit configuration
// func (c *Client) WithRateLimitConfig(rateLimit time.Duration) *Client {
// 	newClient := *c
// 	newClient.options.RateLimit = rateLimit
// 	return &newClient
// }

// buildQueryParams builds URL query parameters with enhanced date range support
func (c *Client) buildQueryParams(query *Query) url.Values {
	params := url.Values{}

	// Build search query with date range if specified
	searchQuery := query.SearchQuery
	if query.SubmittedDateFrom != nil || query.SubmittedDateTo != nil {
		dateFilter := c.buildDateRangeFilter(query.SubmittedDateFrom, query.SubmittedDateTo)
		if searchQuery != "" {
			searchQuery = fmt.Sprintf("(%s) AND %s", searchQuery, dateFilter)
		} else {
			searchQuery = dateFilter
		}
	}

	// Search query or ID list
	if len(query.IDList) > 0 {
		params.Set("id_list", strings.Join(query.IDList, ","))
	} else if searchQuery != "" {
		params.Set("search_query", searchQuery)
	}

	// Pagination
	if query.Start > 0 {
		params.Set("start", strconv.Itoa(query.Start))
	}

	// Max results
	maxResults := query.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}
	params.Set("max_results", strconv.Itoa(maxResults))

	// Sorting
	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = defaultSortBy
	}
	params.Set("sortBy", sortBy)

	sortOrder := query.SortOrder
	if sortOrder == "" {
		sortOrder = defaultSortOrder
	}
	params.Set("sortOrder", sortOrder)

	return params
}

// buildDateRangeFilter builds a date range filter for the search query
func (c *Client) buildDateRangeFilter(from, to *time.Time) string {
	const dateFormat = "20060102"

	if from != nil && to != nil {
		return fmt.Sprintf("submittedDate:[%s TO %s]",
			from.Format(dateFormat),
			to.Format(dateFormat))
	} else if from != nil {
		return fmt.Sprintf("submittedDate:[%s TO *]", from.Format(dateFormat))
	} else if to != nil {
		return fmt.Sprintf("submittedDate:[* TO %s]", to.Format(dateFormat))
	}
	return ""
}

// retryWithBackoff executes a function with exponential backoff retry logic
func (c *Client) retryWithBackoff(ctx context.Context, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt < c.options.RetryAttempts; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		var apiErr *APIError
		if !errors.As(err, &apiErr) || !apiErr.Retry {
			return err
		}

		// Don't delay after the last attempt
		if attempt < c.options.RetryAttempts-1 {
			var delay time.Duration
			if attempt != 0 {
				delay = c.options.RetryDelay
			}
			// Wait before retrying
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return lastErr
}

// applyRateLimit ensures we don't exceed the configured rate limit and updates lastRequest
func (c *Client) applyRateLimit(ctx context.Context) error {
	c.rlMu.Lock()
	defer c.rlMu.Unlock()

	c.lastRequest = time.Now()

	if c.options.RateLimit <= 0 {
		return nil
	}

	elapsed := time.Since(c.lastRequest)
	if elapsed >= c.options.RateLimit {
		return nil
	}

	wait := c.options.RateLimit - elapsed
	t := time.NewTimer(wait)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
