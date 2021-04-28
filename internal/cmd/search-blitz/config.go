package main

import (
	"fmt"
	"time"
)

type Config struct {
	Groups []QueryGroupConfig
}

type QueryGroupConfig struct {
	Name    string
	Queries []QueryConfig
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

func (s *Protocol) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	switch v {
	case "stream", "streaming":
		*s = Stream
	case "batch":
		*s = Batch
	default:
		return fmt.Errorf("invalid search type %s", v)
	}

	return nil
}
