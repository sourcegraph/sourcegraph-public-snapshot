package main

import (
	"testing"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intPtr(v int) *int {
	return &v
}

func TestGenerateDeploymentTrace(t *testing.T) {
	trace, err := GenerateDeploymentTrace(&DeploymentReport{
		DeployedAt: time.RFC822Z,
		PullRequests: []*github.PullRequest{
			{Number: intPtr(32996)},
			{Number: intPtr(32871)},
			{Number: intPtr(32767)},
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
}
