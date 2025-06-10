package arxiv

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// XML structures for parsing arXiv API responses
type atomFeed struct {
	XMLName xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string   `xml:"title"`
	ID      string   `xml:"id"`
	Link    []struct {
		Href string `xml:"href,attr"`
		Rel  string `xml:"rel,attr"`
		Type string `xml:"type,attr"`
	} `xml:"link"`
	Updated      string      `xml:"updated"`
	TotalCount   int         `xml:"http://a9.com/-/spec/opensearch/1.1/ totalResults"`
	StartIndex   int         `xml:"http://a9.com/-/spec/opensearch/1.1/ startIndex"`
	ItemsPerPage int         `xml:"http://a9.com/-/spec/opensearch/1.1/ itemsPerPage"`
	Entries      []atomEntry `xml:"entry"`
}

type atomEntry struct {
	XMLName   xml.Name `xml:"entry"`
	ID        string   `xml:"id"`
	Updated   string   `xml:"updated"`
	Published string   `xml:"published"`
	Title     string   `xml:"title"`
	Summary   string   `xml:"summary"`
	Authors   []struct {
		Name string `xml:"name"`
	} `xml:"author"`
	DOI        string `xml:"http://arxiv.org/schemas/atom doi"`
	Comment    string `xml:"http://arxiv.org/schemas/atom comment"`
	JournalRef string `xml:"http://arxiv.org/schemas/atom journal_ref"`
	Categories []struct {
		Term   string `xml:"term,attr"`
		Scheme string `xml:"scheme,attr"`
	} `xml:"category"`
	Links []struct {
		Href  string `xml:"href,attr"`
		Rel   string `xml:"rel,attr"`
		Type  string `xml:"type,attr"`
		Title string `xml:"title,attr"`
	} `xml:"link"`
}

// parseSearchResponse parses the XML response from arXiv API
func (c *Client) parseSearchResponse(data []byte) (*SearchResults, error) {
	var feed atomFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	papers := make([]Paper, len(feed.Entries))
	for i, entry := range feed.Entries {
		paper, err := c.convertEntryToPaper(entry)
		if err != nil {
			return nil, fmt.Errorf("failed to convert entry %d: %w", i, err)
		}
		papers[i] = *paper
	}

	return &SearchResults{
		Papers:       papers,
		TotalCount:   feed.TotalCount,
		StartIndex:   feed.StartIndex,
		ItemsPerPage: feed.ItemsPerPage,
	}, nil
}

// convertEntryToPaper converts an XML entry to a Paper struct
func (c *Client) convertEntryToPaper(entry atomEntry) (*Paper, error) {
	// Parse dates
	publishedAt, err := time.Parse(time.RFC3339, entry.Published)
	if err != nil {
		return nil, fmt.Errorf("failed to parse published date: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, entry.Updated)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated date: %w", err)
	}

	// Extract arXiv ID from the full ID URL
	id := extractArxivID(entry.ID)

	// Convert authors
	authors := make([]Author, len(entry.Authors))
	for i, author := range entry.Authors {
		authors[i] = Author{
			Name: strings.TrimSpace(author.Name),
		}
	}

	// Convert categories
	categories := make([]string, len(entry.Categories))
	for i, cat := range entry.Categories {
		categories[i] = cat.Term
	}

	// Convert links
	links := make([]Link, len(entry.Links))
	for i, link := range entry.Links {
		links[i] = Link{
			Href:  link.Href,
			Rel:   link.Rel,
			Type:  link.Type,
			Title: link.Title,
		}
	}

	return &Paper{
		ID:          id,
		Title:       strings.TrimSpace(entry.Title),
		Abstract:    strings.TrimSpace(entry.Summary),
		Authors:     authors,
		Categories:  categories,
		PublishedAt: publishedAt,
		UpdatedAt:   updatedAt,
		DOI:         entry.DOI,
		JournalRef:  entry.JournalRef,
		Comment:     entry.Comment,
		Links:       links,
	}, nil
}

// extractArxivID extracts the arXiv ID from the full ID URL
// Example: "http://arxiv.org/abs/1234.5678v1" -> "1234.5678v1"
func extractArxivID(fullID string) string {
	// Remove the URL prefix to get just the ID
	if strings.HasPrefix(fullID, "http://arxiv.org/abs/") {
		return strings.TrimPrefix(fullID, "http://arxiv.org/abs/")
	}
	return fullID
}
