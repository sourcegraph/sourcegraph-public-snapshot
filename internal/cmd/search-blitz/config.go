package main

import "fmt"

type Config struct {
	Groups []QueryGroupConfig
}

type QueryGroupConfig struct {
	Name    string
	Queries []QueryConfig
}

type QueryConfig struct {
	Query      string
	Name       string
	SearchType SearchType `yaml:"search_type"`
}

// SearchType specifies whether to run the query against the batch
// graphQL API, the streaming API, or both.
type SearchType uint8

func (s SearchType) HasBatch() bool {
	return s == SearchTypeBoth || s == SearchTypeStream
}

func (s SearchType) HasStream() bool {
	return s == SearchTypeBoth || s == SearchTypeStream
}

func (s SearchType) String() string {
	switch s {
	case SearchTypeBoth:
		return "both"
	case SearchTypeStream:
		return "stream"
	case SearchTypeBatch:
		return "batch"
	default:
		panic("invalid search type")
	}
}

const (
	// The zero value defaults to "both"
	SearchTypeBoth SearchType = iota
	SearchTypeBatch
	SearchTypeStream
)

func (s *SearchType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	switch v {
	case "stream", "streaming":
		*s = SearchTypeStream
	case "batch":
		*s = SearchTypeBatch
	case "both", "":
		*s = SearchTypeBoth
	default:
		return fmt.Errorf("invalid search type %s", v)
	}

	return nil
}
