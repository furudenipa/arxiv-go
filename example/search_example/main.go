package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/furudenipa/arxiv-go"
)

func main() {
	// Example 1: Create client with custom options (retry, rate limiting)
	fmt.Println("=== Enhanced Client with Custom Options ===")
	opts := arxiv.ClientOptions{
		RetryAttempts: 5,
		RetryDelay:    2 * time.Second,
		RateLimit:     1 * time.Second,
		UserAgent:     "my-app/1.0",
		Timeout:       60 * time.Second,
	}
	client := arxiv.NewClientWithOptions(opts)

	// Example 2: Search with date range filtering
	fmt.Println("\n=== Search with Date Range ===")
	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

	query := &arxiv.Query{
		SearchQuery:       "quantum computing",
		MaxResults:        3,
		SubmittedDateFrom: &from,
		SubmittedDateTo:   &to,
		SortBy:            string(arxiv.SortBySubmittedDate),
		SortOrder:         string(arxiv.SortOrderDescending),
	}

	results, err := client.Search(context.Background(), query)
	if err != nil {
		// Enhanced error handling with detailed error types
		if apiErr, ok := err.(*arxiv.APIError); ok {
			fmt.Printf("API Error Type: %s\n", apiErr.Type.String())
			fmt.Printf("Retryable: %v\n", apiErr.Retry)
			fmt.Printf("Error: %v\n", apiErr)
		} else {
			log.Fatalf("Search failed: %v", err)
		}
		return
	}

	fmt.Printf("Found %d papers from 2023:\n", len(results.Papers))
	for i, paper := range results.Papers {
		fmt.Printf("\n%d. %s\n", i+1, paper.Title)
		fmt.Printf("   ID: %s\n", paper.ID)
		fmt.Printf("   Published: %s\n", paper.PublishedAt.Format("2006-01-02"))
		fmt.Printf("   Categories: %v\n", paper.Categories)
	}

	// Example 3: Using category constants for type safety
	fmt.Println("\n\n=== Search by Category (Type-safe) ===")
	categoryQuery := &arxiv.Query{
		SearchQuery: fmt.Sprintf("cat:%s", arxiv.CategoryCSAI), // cs.AI category
		MaxResults:  2,
		SortBy:      string(arxiv.SortByRelevance),
		SortOrder:   string(arxiv.SortOrderDescending),
	}

	categoryResults, err := client.Search(context.Background(), categoryQuery)
	if err != nil {
		log.Printf("Category search failed: %v", err)
	} else {
		fmt.Printf("Found %d papers in %s category:\n", len(categoryResults.Papers), arxiv.CategoryCSAI)
		for i, paper := range categoryResults.Papers {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, paper.Title, paper.ID)
		}
	}

	// Example 4: Enhanced error handling with GetByID
	fmt.Println("\n\n=== Enhanced Get By ID ===")
	paper, err := client.GetByID(context.Background(), "1234.5678") // Non-existent ID
	if err != nil {
		if apiErr, ok := err.(*arxiv.APIError); ok {
			if apiErr.Type == arxiv.ErrorTypeNotFound {
				fmt.Printf("Paper not found (as expected): %s\n", apiErr.Message)
			} else {
				fmt.Printf("Other API error: %s\n", apiErr.Error())
			}
		} else {
			fmt.Printf("Unexpected error: %v\n", err)
		}
	} else {
		fmt.Printf("Found paper: %s\n", paper.Title)
	}

	// Example 5: Demonstrate all available sort criteria and orders
	fmt.Println("\n\n=== Available Constants ===")
	fmt.Printf("Sort Criteria:\n")
	fmt.Printf("- Relevance: %s\n", arxiv.SortByRelevance)
	fmt.Printf("- Last Updated: %s\n", arxiv.SortByLastUpdatedDate)
	fmt.Printf("- Submitted Date: %s\n", arxiv.SortBySubmittedDate)

	fmt.Printf("\nSort Orders:\n")
	fmt.Printf("- Ascending: %s\n", arxiv.SortOrderAscending)
	fmt.Printf("- Descending: %s\n", arxiv.SortOrderDescending)

	fmt.Printf("\nSample Categories:\n")
	fmt.Printf("- Computer Science AI: %s\n", arxiv.CategoryCSAI)
	fmt.Printf("- Quantum Physics: %s\n", arxiv.CategoryQuantPh)
	fmt.Printf("- Economics: %s\n", arxiv.CategoryEconEM)
	fmt.Printf("- Machine Learning: %s\n", arxiv.CategoryCSLG)

	// Example 6: Backward compatibility - original methods still work
	fmt.Println("\n\n=== Backward Compatibility ===")
	legacyClient := arxiv.NewClient() // Uses default options
	legacyQuery := &arxiv.Query{
		SearchQuery: "machine learning",
		MaxResults:  1,
	}

	legacyResults, err := legacyClient.Search(context.Background(), legacyQuery)
	if err != nil {
		log.Printf("Legacy search failed: %v", err)
	} else {
		fmt.Printf("Legacy search found %d papers\n", len(legacyResults.Papers))
		if len(legacyResults.Papers) > 0 {
			fmt.Printf("Title: %s\n", legacyResults.Papers[0].Title)
		}
	}
}
