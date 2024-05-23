package main

import (
	"context"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/buildkite/go-buildkite/v3/buildkite"
)

type BigQueryWriter interface {
	Write(ctx context.Context, values ...bigquery.ValueSaver) error
}

type BuildkiteAgentEvent struct {
	event     string
	timestamp time.Time
	buildkite.Agent
}

// Save implements bigquery.ValueSaver.
func (b *BuildkiteAgentEvent) Save() (row map[string]bigquery.Value, insertID string, err error) {
	var queues []string
	for _, meta := range b.Metadata {
		if strings.HasPrefix(meta, "queue=") {
			queues = append(queues, strings.TrimPrefix(meta, "queue="))
		}
	}

	return map[string]bigquery.Value{
		"event":      strings.TrimPrefix(b.event, "agent."),
		"uuid":       *b.ID,
		"name":       *b.Name,
		"hostname":   *b.Hostname,
		"version":    *b.Version,
		"ip_address": *b.IPAddress,
		"queues":     queues,
		"user_agent": *b.UserAgent,
		"timestamp":  b.timestamp,
	}, "", nil
}

var _ (bigquery.ValueSaver) = (*BuildkiteAgentEvent)(nil)
