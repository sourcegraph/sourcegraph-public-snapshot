package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sourcegraph/sitemap"
)

func main() {
	lastMod := time.Now()

	var minScore, maxScore int
	var urls []sitemap.URL
	for {
		var url string
		var score int

		n, err := fmt.Scanf("%s %d", &url, &score)
		if n < 2 {
			break
		}
		if err != nil {
			panic(err)
		}

		if score < minScore || len(urls) == 0 {
			minScore = score
		}
		if score > maxScore || len(urls) == 0 {
			maxScore = score
		}

		urls = append(urls, sitemap.URL{
			Loc:        url,
			LastMod:    &lastMod,
			ChangeFreq: sitemap.Monthly,
			Priority:   float64(score),
		})
	}

	// second pass to map priorities into [0, 1]
	for i := range urls {
		urls[i].Priority = (urls[i].Priority - float64(minScore)) / float64(maxScore-minScore)
	}

	xml, err := sitemap.Marshal(&sitemap.URLSet{URLs: urls})
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(xml)
	os.Stdout.WriteString("\n")
}
