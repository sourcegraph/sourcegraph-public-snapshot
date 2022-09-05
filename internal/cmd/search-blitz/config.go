package main

import (
	"bytes"
	"embed"
	_ "embed"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//go:embed queries*.txt
var queriesFS embed.FS

type Config struct {
	Groups []*QueryGroupConfig
}

type QueryGroupConfig struct {
	Name    string
	Queries []*QueryConfig
}

type QueryConfig struct {
	Query string
	Name  string

	// An unset interval defaults to 1m
	Interval time.Duration

	// An empty value for Protocols means "all"
	Protocols []Protocol
}

var allProtocols = []Protocol{Batch, Stream}

// Protocol represents either the graphQL Protocol or the streaming Protocol
type Protocol uint8

const (
	Batch Protocol = iota
	Stream
)

func loadQueries(env string) (_ *Config, err error) {
	if env == "" {
		env = "cloud"
	}

	queriesRaw, err := queriesFS.ReadFile("queries.txt")
	if err != nil {
		return nil, err
	}

	if env != "cloud" {
		extra, err := queriesFS.ReadFile("queries_" + env + ".txt")
		if err != nil {
			return nil, err
		}
		queriesRaw = append(queriesRaw, '\n')
		queriesRaw = append(queriesRaw, extra...)
	}

	var queries []*QueryConfig
	var current QueryConfig
	add := func() {
		q := &QueryConfig{
			Name:  strings.TrimSpace(current.Name),
			Query: strings.TrimSpace(current.Query),
		}
		current = QueryConfig{} // reset
		if q.Query == "" {
			return
		}
		if q.Name == "" {
			err = errors.Errorf("no name set for query %q", q.Query)
		}
		queries = append(queries, q)
	}
	for _, line := range bytes.Split(queriesRaw, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if line[0] == '#' {
			add()
			current.Name = string(line[1:])
		} else {
			current.Query += " " + string(line)
		}
	}
	add()

	return &Config{
		Groups: []*QueryGroupConfig{{
			Name:    "monitoring_queries",
			Queries: queries,
		}},
	}, err
}
