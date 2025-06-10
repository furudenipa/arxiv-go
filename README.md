# arxiv-go

Go wrapper for the arXiv API

## Installation

```bash
go get github.com/furudenipa/arxiv-go
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/furudenipa/arxiv-go"
)

func main() {
    client := arxiv.NewClient()
    
    query := &arxiv.Query{
        SearchQuery: "quantum computing",
        MaxResults:  10,
    }
    
    results, err := client.Search(context.Background(), query)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, paper := range results.Papers {
        fmt.Printf("Title: %s\n", paper.Title)
        fmt.Printf("Authors: %v\n", paper.Authors)
        fmt.Printf("Abstract: %s\n\n", paper.Abstract)
    }
}
```

## Features

- Search papers by keywords, authors, categories
- Fetch paper details by arXiv ID
- Support for pagination
- Rate limiting
- Context support for cancellation

## License

MIT
