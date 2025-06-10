package arxiv

import (
	"fmt"
	"time"
)

// Paper represents an arXiv paper
type Paper struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Abstract    string    `json:"abstract"`
	Authors     []Author  `json:"authors"`
	Categories  []string  `json:"categories"`
	PublishedAt time.Time `json:"published_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DOI         string    `json:"doi,omitempty"`
	JournalRef  string    `json:"journal_ref,omitempty"`
	Comment     string    `json:"comment,omitempty"`
	Links       []Link    `json:"links"`
}

// Author represents a paper author
type Author struct {
	Name        string `json:"name"`
	Affiliation string `json:"affiliation,omitempty"`
}

// Link represents a link associated with a paper
type Link struct {
	Href  string `json:"href"`
	Rel   string `json:"rel"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
}

// Query represents search parameters for arXiv API
type Query struct {
	// Search query string (e.g., "quantum computing", "au:Einstein")
	SearchQuery string

	// arXiv ID list (alternative to SearchQuery)
	IDList []string

	// Start index for pagination (0-based)
	Start int

	// Maximum number of results per API request (default: 100, max: 30000)
	MaxResults int

	// Maximum total number of results to fetch across all requests (0 = unlimited)
	Limit int

	// Sort criteria: "relevance", "lastUpdatedDate", "submittedDate"
	SortBy string

	// Sort order: "ascending", "descending"
	SortOrder string

	// Date range filtering
	SubmittedDateFrom *time.Time
	SubmittedDateTo   *time.Time
}

// SearchResults represents the response from arXiv API
type SearchResults struct {
	Papers       []Paper `json:"papers"`         // List of papers returned by the search
	TotalCount   int     `json:"total_count"`    // Total number of papers matching the query (not fetched papers)
	StartIndex   int     `json:"start_index"`    // Start index of the current page (0-based)
	ItemsPerPage int     `json:"items_per_page"` // Number of papers in the current page
}

// ErrorType represents the type of error that occurred
type ErrorType int

const (
	ErrorTypeRateLimit ErrorType = iota
	ErrorTypeTimeout
	ErrorTypeParsing
	ErrorTypeNetwork
	ErrorTypeNotFound
	ErrorTypeInvalidQuery
	ErrorTypeNoEntry // Check https://github.com/lukasschwab/arxiv.py/issues/129
	ErrorTypeUnknown
)

// String returns a string representation of the error type
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeRateLimit:
		return "rate_limit"
	case ErrorTypeTimeout:
		return "timeout"
	case ErrorTypeParsing:
		return "parsing"
	case ErrorTypeNetwork:
		return "network"
	case ErrorTypeNotFound:
		return "not_found"
	case ErrorTypeInvalidQuery:
		return "invalid_query"
	default:
		return "unknown"
	}
}

// APIError represents a detailed arXiv API error
type APIError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Code    int       `json:"code,omitempty"`
	Retry   bool      `json:"retry"`
	Err     error     `json:"-"`
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type.String(), e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type.String(), e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// NewAPIError creates a new APIError
func NewAPIError(errorType ErrorType, message string, err error) *APIError {
	retry := errorType == ErrorTypeRateLimit || errorType == ErrorTypeTimeout || errorType == ErrorTypeNetwork || errorType == ErrorTypeNoEntry
	return &APIError{
		Type:    errorType,
		Message: message,
		Retry:   retry,
		Err:     err,
	}
}

// Legacy Error type for backward compatibility
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}
