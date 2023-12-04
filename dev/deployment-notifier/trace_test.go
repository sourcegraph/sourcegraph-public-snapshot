package main

import (
	"testing"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestGenerateDeploymentTrace(t *testing.T) {
	trace, err := GenerateDeploymentTrace(&DeploymentReport{
		Environment: "preprepod",
		DeployedAt:  time.RFC822Z,
		PullRequests: []*github.PullRequest{
			{Number: pointers.Ptr(32996)},
			{Number: pointers.Ptr(32871)},
			{Number: pointers.Ptr(32767)},
		},
		ServicesPerPullRequest: map[int][]string{
			32996: {"frontend", "gitserver", "worker"},
			32871: {"frontend", "gitserver", "worker"},
			32767: {"gitserver"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, trace)

	const (
		expectPRSpans      = 3
		expectServiceSpans = 3 + 3 + 1
	)
	assert.NotEmpty(t, trace.ID)
	assert.NotNil(t, trace.Root)
	assert.Equal(t, expectPRSpans+expectServiceSpans, len(trace.Spans))

	// Assert fields every event should have
	for _, ev := range append(trace.Spans, trace.Root) {
		assert.Equal(t, ev.Fields()["environment"], "preprepod")
	}
}
