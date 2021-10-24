package main

import (
	"flag"
	"fmt"
	"sort"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
)

var token string
var date string
var pipeline string

func init() {
	flag.StringVar(&token, "token", "", "mandatory buildkite token")
	flag.StringVar(&date, "date", "", "date for builds")
	flag.StringVar(&pipeline, "pipeline", "sourcegraph", "name of the pipeline to inspect")
}

type event struct {
	at       time.Time
	state    string
	buildURL string
}

func main() {
	flag.Parse()

	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}

	config, err := buildkite.NewTokenConfig(token, false)
	if err != nil {
		panic(err)
	}
	client := buildkite.NewClient(config.Client())

	var builds []buildkite.Build
	nextPage := 0
	for {
		bs, resp, err := client.Builds.ListByPipeline("sourcegraph", pipeline, &buildkite.BuildsListOptions{
			Branch:      "main",
			CreatedFrom: BoD(t).Add(-1 * 24 * time.Hour), // day before
			CreatedTo:   EoD(t),
			ListOptions: buildkite.ListOptions{Page: nextPage},
		})
		if err != nil {
			panic(err)
		}
		nextPage = resp.NextPage
		builds = append(builds, bs...)

		if nextPage == 0 {
			break
		}
	}

	if len(builds) <= 0 {
		panic("no builds")
	}

	ends := []*event{}
	for _, b := range builds {
		if b.FinishedAt != nil {
			ends = append(ends, &event{
				at:       b.FinishedAt.Time,
				state:    *b.State,
				buildURL: *b.WebURL,
			})
		}
	}
	sort.Slice(ends, func(i, j int) bool { return ends[i].at.Before(ends[j].at) })

	var lastRed *event
	red := time.Duration(0)
	for _, event := range ends {
		if event.state == "failed" {
			// if a build failed, compute how much time until the next green
			lastRed = event
			fmt.Printf("failure at %s, %s\n", event.at.Format(time.RFC822), event.buildURL)
		}
		if event.state == "passed" && lastRed != nil {
			// if a build passed and we previously were red, stop recording the duration.
			red += event.at.Sub(lastRed.at)
			lastRed = nil
			fmt.Printf("fixed at   %s, %s\n", event.at.Format(time.RFC822), event.buildURL)
		}
	}
	fmt.Printf("On %s, the pipeline was red for %s\n", date, red.String())
}

func BoD(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func EoD(t time.Time) time.Time {
	return BoD(t).Add(time.Hour * 24).Add(-1 * time.Nanosecond)
}
