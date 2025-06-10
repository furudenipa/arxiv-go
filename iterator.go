package arxiv

import (
	"context"
	"iter"
)

// IteratorState represents the current state of the iterator
type IteratorState int

const (
	StateInitial   IteratorState = iota // Initial state, no data fetched yet
	StateFetching                       // Currently fetching data from API
	StateReady                          // Data is ready and available for iteration
	StateExhausted                      // No more data available
	StateError                          // An error occurred
)

// String returns a string representation of the iterator state
func (s IteratorState) String() string {
	switch s {
	case StateInitial:
		return "initial"
	case StateFetching:
		return "fetching"
	case StateReady:
		return "ready"
	case StateExhausted:
		return "exhausted"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// State represents an immutable state of the iterator
type State struct {
	Current      IteratorState // Current state of the iterator
	CurrentPage  int           // Current page number
	CurrentIndex int           // Current index within the current page
	TotalFetched int           // Total number of papers fetched so far
	Error        error
	Results      *SearchResults
}

// Action represents a state transition action
type Action interface {
	Apply(state State) State
}

// FetchAction represents fetching new data
type FetchAction struct {
	Results *SearchResults
	Error   error
}

func (a FetchAction) Apply(state State) State {
	if a.Error != nil {
		return State{
			Current:      StateError,
			CurrentPage:  state.CurrentPage,
			CurrentIndex: state.CurrentIndex,
			TotalFetched: state.TotalFetched,
			Error:        a.Error,
			Results:      state.Results,
		}
	}

	if a.Results == nil || len(a.Results.Papers) == 0 {
		return State{
			Current:      StateExhausted,
			CurrentPage:  state.CurrentPage,
			CurrentIndex: state.CurrentIndex,
			TotalFetched: state.TotalFetched,
			Error:        nil,
			Results:      a.Results,
		}
	}

	return State{
		Current:      StateReady,
		CurrentPage:  state.CurrentPage + 1,
		CurrentIndex: 0,
		TotalFetched: state.TotalFetched,
		Error:        nil,
		Results:      a.Results,
	}
}

// ConsumeAction represents consuming a paper from current results
type ConsumeAction struct{}

func (a ConsumeAction) Apply(state State) State {
	return State{
		Current:      state.Current,
		CurrentPage:  state.CurrentPage,
		CurrentIndex: state.CurrentIndex + 1,
		TotalFetched: state.TotalFetched + 1,
		Error:        state.Error,
		Results:      state.Results,
	}
}

// StateManager manages state transitions
type StateManager struct {
	state State
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		state: State{Current: StateInitial},
	}
}

// GetState returns the current state (immutable)
func (sm *StateManager) GetState() State {
	return sm.state
}

// Transition applies an action and updates the state
func (sm *StateManager) Transition(action Action) State {
	sm.state = action.Apply(sm.state)
	return sm.state
}

// Reset resets the state to initial
func (sm *StateManager) Reset() {
	sm.state = State{Current: StateInitial}
}

// Paginator handles pagination logic
type Paginator struct {
	query *Query
}

// NewPaginator creates a new paginator
func NewPaginator(query *Query) *Paginator {
	return &Paginator{query: query}
}

// CalculateStartIndex calculates the start index for the next page
func (p *Paginator) CalculateStartIndex(currentPage int, results *SearchResults) int {
	if results != nil {
		return results.StartIndex + len(results.Papers)
	}
	return currentPage * p.query.MaxResults
}

// CalculateMaxResults calculates how many results to fetch considering the limit
func (p *Paginator) CalculateMaxResults(totalFetched int) int {
	maxResults := p.query.MaxResults
	if p.query.Limit > 0 {
		remaining := p.query.Limit - totalFetched
		if remaining < maxResults {
			maxResults = remaining
		}
	}
	return maxResults
}

// HasMoreData checks if more data might be available
func (p *Paginator) HasMoreData(state State) bool {
	// If we haven't fetched anything yet, there might be data
	if state.Results == nil {
		return true
	}

	// Check user-specified limit
	if p.query.Limit > 0 && state.TotalFetched >= p.query.Limit {
		return false
	}

	// Check total count if known
	expectedTotal := state.Results.StartIndex + len(state.Results.Papers)
	if state.Results.TotalCount > 0 && expectedTotal >= state.Results.TotalCount {
		return false
	}

	// If we got fewer results than requested, probably no more
	if len(state.Results.Papers) < p.query.MaxResults {
		return false
	}

	return true
}

// Fetcher handles API requests
type Fetcher struct {
	client *Client
	ctx    context.Context
}

// NewFetcher creates a new fetcher
func NewFetcher(client *Client, ctx context.Context) *Fetcher {
	return &Fetcher{client: client, ctx: ctx}
}

// Fetch fetches data from the API
func (f *Fetcher) Fetch(query *Query) (*SearchResults, error) {
	if query == nil {
		return nil, NewAPIError(ErrorTypeInvalidQuery, "query is nil", nil)
	}
	return f.client.Search(f.ctx, query)
}

// WithContext creates a new fetcher with a different context
func (f *Fetcher) WithContext(ctx context.Context) *Fetcher {
	return &Fetcher{client: f.client, ctx: ctx}
}

// Iterator provides a clean interface for iterating through paginated search results
type Iterator struct {
	paginator    *Paginator
	fetcher      *Fetcher
	stateManager *StateManager
	query        *Query
}

// NewIterator creates a new iterator
func NewIterator(client *Client, query *Query, ctx context.Context) *Iterator {
	return &Iterator{
		paginator:    NewPaginator(query),
		fetcher:      NewFetcher(client, ctx),
		stateManager: NewStateManager(),
		query:        query,
	}
}

// needsMoreData checks if we need to fetch more data
func (it *Iterator) needsMoreData(state State) bool {
	// No results yet
	if state.Results == nil {
		return true
	}

	// Consumed all current papers
	if state.CurrentIndex >= len(state.Results.Papers) {
		return it.paginator.HasMoreData(state)
	}

	return false
}

// nextPaper returns the next paper, handling all state transitions
func (it *Iterator) nextPaper() (*Paper, error) {
	state := it.stateManager.GetState()

	switch state.Current {
	case StateError:
		return nil, state.Error

	case StateExhausted:
		return nil, nil

	case StateInitial, StateReady:
		// Check if we need to fetch more data
		if it.needsMoreData(state) {
			// Check if there's more data available
			if !it.paginator.HasMoreData(state) {
				it.stateManager.Transition(FetchAction{Results: state.Results, Error: nil})
				return nil, nil
			}

			// Create query for next page
			nextQuery := *it.query
			nextQuery.Start = it.paginator.CalculateStartIndex(state.CurrentPage, state.Results)
			nextQuery.MaxResults = it.paginator.CalculateMaxResults(state.TotalFetched)

			// Fetch data
			results, err := it.fetcher.Fetch(&nextQuery)
			newState := it.stateManager.Transition(FetchAction{Results: results, Error: err})

			if newState.Current == StateError {
				return nil, newState.Error
			}
			if newState.Current == StateExhausted {
				return nil, nil
			}

			state = newState
		}

		// Check if we have papers available
		if state.Results != nil && state.CurrentIndex < len(state.Results.Papers) {
			// Check limit before yielding
			if it.query.Limit > 0 && state.TotalFetched >= it.query.Limit {
				it.stateManager.Transition(FetchAction{Results: state.Results, Error: nil})
				return nil, nil
			}

			paper := &state.Results.Papers[state.CurrentIndex]
			it.stateManager.Transition(ConsumeAction{})
			return paper, nil
		}

		// No papers available
		it.stateManager.Transition(FetchAction{Results: state.Results, Error: nil})
		return nil, nil

	default:
		return nil, NewAPIError(ErrorTypeUnknown, "unknown iterator state", nil)
	}
}

// All returns an iterator that yields papers one by one using Go 1.23+ iter pattern
func (it *Iterator) All() iter.Seq[*Paper] {
	return func(yield func(*Paper) bool) {
		for {
			paper, err := it.nextPaper()
			if err != nil || paper == nil {
				return
			}
			if !yield(paper) {
				return
			}
		}
	}
}

// AllWithError returns an iterator that yields papers with error handling
func (it *Iterator) AllWithError() iter.Seq2[*Paper, error] {
	return func(yield func(*Paper, error) bool) {
		for {
			paper, err := it.nextPaper()
			if err != nil {
				yield(nil, err)
				return
			}
			if paper == nil {
				return
			}
			if !yield(paper, nil) {
				return
			}
		}
	}
}

// Values is an alias for All() for compatibility with standard naming conventions
func (it *Iterator) Values() iter.Seq[*Paper] {
	return it.All()
}

// Error returns any error that occurred during iteration
func (it *Iterator) Error() error {
	return it.stateManager.GetState().Error
}

// TotalFetched returns the total number of papers fetched so far
func (it *Iterator) TotalFetched() int {
	return it.stateManager.GetState().TotalFetched
}

// TotalCount returns the total number of results available (if known)
func (it *Iterator) TotalCount() int {
	state := it.stateManager.GetState()
	if state.Results == nil {
		return -1
	}
	return state.Results.TotalCount
}

// CurrentPage returns the current page number (0-based)
func (it *Iterator) CurrentPage() int {
	return it.stateManager.GetState().CurrentPage
}

// Reset resets the iterator to the beginning
func (it *Iterator) Reset() {
	it.stateManager.Reset()
	if it.query != nil {
		it.query.Start = 0
	}
}

// WithContext creates a new iterator with a different context
func (it *Iterator) WithContext(ctx context.Context) *Iterator {
	return &Iterator{
		paginator:    it.paginator,
		fetcher:      it.fetcher.WithContext(ctx),
		stateManager: NewStateManager(),
		query:        it.query,
	}
}

// ForEach iterates through all remaining papers
func (it *Iterator) ForEach(fn func(*Paper) error) error {
	for paper := range it.All() {
		if err := fn(paper); err != nil {
			return err
		}
	}
	return it.Error()
}

// Collect returns all remaining papers as a slice
func (it *Iterator) Collect() ([]*Paper, error) {
	var papers []*Paper
	for paper := range it.All() {
		papers = append(papers, paper)
	}
	return papers, it.Error()
}

// CollectN returns up to n papers as a slice
func (it *Iterator) CollectN(n int) ([]*Paper, error) {
	var papers []*Paper
	count := 0
	for paper := range it.All() {
		if count >= n {
			break
		}
		papers = append(papers, paper)
		count++
	}
	return papers, it.Error()
}

// Package-level helper functions for working with iter.Seq

// ForEachSeq applies a function to each element in an iter.Seq
func ForEachSeq[T any](seq iter.Seq[T], fn func(T) error) error {
	for item := range seq {
		if err := fn(item); err != nil {
			return err
		}
	}
	return nil
}

// CollectSeq collects all elements from an iter.Seq into a slice
func CollectSeq[T any](seq iter.Seq[T]) []T {
	var result []T
	for item := range seq {
		result = append(result, item)
	}
	return result
}

// CollectNSeq collects up to n elements from an iter.Seq into a slice
func CollectNSeq[T any](seq iter.Seq[T], n int) []T {
	var result []T
	count := 0
	for item := range seq {
		if count >= n {
			break
		}
		result = append(result, item)
		count++
	}
	return result
}

// TakeSeq returns an iterator that yields at most n elements
func TakeSeq[T any](seq iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		count := 0
		for item := range seq {
			if count >= n {
				return
			}
			if !yield(item) {
				return
			}
			count++
		}
	}
}

// FilterSeq returns an iterator that yields only elements that satisfy the predicate
func FilterSeq[T any](seq iter.Seq[T], predicate func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for item := range seq {
			if predicate(item) {
				if !yield(item) {
					return
				}
			}
		}
	}
}
