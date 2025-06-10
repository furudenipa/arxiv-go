package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/furudenipa/arxiv-go"
)

func main() {
	// Create a new client
	client := arxiv.NewClient()
	ctx := context.Background()

	fmt.Println("=== Go 1.23+ Iterator Pattern Examples ===")

	// Example 1: Basic range-over-func iteration with Limit
	fmt.Println("Example 1: Basic range-over-func iteration with Limit")
	fmt.Println("------------------------------------------------------")
	iter := client.NewQuery().
		SearchQuery("quantum computing").
		Category(arxiv.CategoryQuantPh).
		Limit(3).
		Iterator(ctx)

	fmt.Println("Papers on quantum computing:")
	for paper := range iter.All() {
		fmt.Printf("• %s\n", paper.Title)
		fmt.Printf("  Authors: %s\n", getAuthorNames(paper.Authors))
		fmt.Printf("  Published: %s\n\n", paper.PublishedAt.Format("2006-01-02"))
	}

	if err := iter.Error(); err != nil {
		log.Printf("Error: %v", err)
	}

	// Example 2: Error handling with AllWithError
	fmt.Println("Example 2: Error handling with AllWithError")
	fmt.Println("--------------------------------------------")
	iter = client.NewQuery().
		SearchQuery("machine learning").
		Category(arxiv.CategoryCSLG).
		Limit(3).
		Iterator(ctx)

	fmt.Println("Papers on machine learning (with error handling):")
	for paper, err := range iter.AllWithError() {
		if err != nil {
			log.Printf("Error occurred: %v", err)
			break
		}
		fmt.Printf("• %s\n", paper.Title)
	}

	// Example 3: Using helper functions with iter.Seq
	fmt.Println("\nExample 3: Using helper functions")
	fmt.Println("---------------------------------")
	iter = client.NewQuery().
		SearchQuery("transformer").
		Category(arxiv.CategoryCSAI).
		Iterator(ctx)
	// Collect first 5 papers
	fmt.Println("First 5 papers using TakeSeq:")
	first5 := arxiv.CollectSeq(arxiv.TakeSeq(iter.All(), 5))
	for i, paper := range first5 {
		fmt.Printf("%d. %s\n", i+1, paper.Title)
	}
	fmt.Println("total fetched:", iter.TotalFetched(), "total count:", iter.TotalCount())
	// Reset and filter papers
	iter.Reset()
	fmt.Println("\nFiltered papers (containing 'deep' in title):")
	filtered := arxiv.CollectSeq(arxiv.FilterSeq(iter.All(), func(paper *arxiv.Paper) bool {
		return strings.Contains(strings.ToLower(paper.Title), "deep")
	}))
	fmt.Println("total fetched:", iter.TotalFetched(), "total count:", iter.TotalCount())

	if len(filtered) == 0 {
		fmt.Println("No papers found with 'deep' in title")
	} else {
		for i, paper := range filtered {
			fmt.Printf("%d. %s\n", i+1, paper.Title)
		}
	}

	// Example 4: Processing with ForEachSeq
	fmt.Println("\nExample 4: Processing with ForEachSeq")
	fmt.Println("-------------------------------------")
	iter = client.NewQuery().
		SearchQuery("computer vision").
		Category(arxiv.CategoryCSCV).
		Limit(5).
		Iterator(ctx)

	fmt.Println("Processing papers with ForEachSeq:")
	err := arxiv.ForEachSeq(iter.All(), func(paper *arxiv.Paper) error {
		fmt.Printf("Processing: %s\n", truncateTitle(paper.Title, 50))
		fmt.Printf("  Categories: %v\n", paper.Categories)

		// Simulate some processing
		if len(paper.Authors) == 0 {
			return fmt.Errorf("paper has no authors")
		}
		return nil
	})

	if err != nil {
		log.Printf("Processing error: %v", err)
	}

	// Example 5: Chaining operations
	fmt.Println("\nExample 5: Chaining multiple operations")
	fmt.Println("---------------------------------------")
	iter = client.NewQuery().
		SearchQuery("artificial intelligence").
		Categories(arxiv.CategoryCSAI, arxiv.CategoryCSLG).
		Limit(15).
		Iterator(ctx)

	fmt.Println("Recent AI papers (filtered, limited, and processed):")

	// Chain: Filter by recent years -> Take first 3 -> Process each
	recentPapers := arxiv.TakeSeq(
		arxiv.FilterSeq(iter.All(), func(paper *arxiv.Paper) bool {
			return paper.PublishedAt.Year() >= 2020
		}),
		3,
	)

	count := 0
	for paper := range recentPapers {
		count++
		fmt.Printf("%d. %s\n", count, truncateTitle(paper.Title, 60))
		fmt.Printf("   Year: %d, Authors: %d\n",
			paper.PublishedAt.Year(),
			len(paper.Authors))
	}

	// Example 6: Early termination and resource management
	fmt.Println("\nExample 6: Early termination")
	fmt.Println("-----------------------------")
	iter = client.NewQuery().
		SearchQuery("deep learning").
		Category(arxiv.CategoryCSLG).
		MaxResults(20).
		Iterator(ctx)

	fmt.Println("Looking for the first paper with 'attention' in title:")
	found := false
	for paper := range iter.All() {
		fmt.Printf("Checking: %s\n", truncateTitle(paper.Title, 50))
		if strings.Contains(strings.ToLower(paper.Title), "attention") {
			fmt.Printf("✓ Found! %s\n", paper.Title)
			found = true
			break // Early termination
		}
	}

	if !found {
		fmt.Println("No paper with 'attention' in title found in the first 20 results")
	}

	fmt.Printf("Total papers fetched: %d\n", iter.TotalFetched())

	// Example 7: Demonstrating Limit vs MaxResults
	fmt.Println("\nExample 7: Demonstrating Limit vs MaxResults")
	fmt.Println("--------------------------------------------")
	fmt.Println("Limit(5) with MaxResults(50) - efficient fetching:")
	iter = client.NewQuery().
		SearchQuery("machine learning").
		Limit(5).       // Total: 5 papers
		MaxResults(50). // Per request: 50 papers (efficient)
		Iterator(ctx)

	count = 0
	for paper := range iter.All() {
		count++
		fmt.Printf("%d. %s\n", count, truncateTitle(paper.Title, 60))
	}
	fmt.Printf("Total fetched: %d papers\n\n", iter.TotalFetched())

	// Example 8: Using Values() alias
	fmt.Println("Example 8: Using Values() alias")
	fmt.Println("-------------------------------")
	iter = client.NewQuery().
		SearchQuery("robotics").
		Category(arxiv.CategoryCSRO).
		Limit(3).
		Iterator(ctx)

	fmt.Println("Robotics papers using Values() alias:")
	for paper := range iter.Values() {
		fmt.Printf("• %s\n", truncateTitle(paper.Title, 50))
	}

	fmt.Println("\n=== Iterator Pattern Examples Complete ===")
}

// Helper function to get author names as a string
func getAuthorNames(authors []arxiv.Author) string {
	if len(authors) == 0 {
		return "No authors"
	}

	names := make([]string, len(authors))
	for i, author := range authors {
		names[i] = author.Name
	}
	return strings.Join(names, ", ")
}

// Helper function to truncate long titles
func truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen-3] + "..."
}
