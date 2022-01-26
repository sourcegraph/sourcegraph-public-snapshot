package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// See rendered output with:
// 	go test -timeout 30s -run ^TestSlackSummary$ github.com/sourcegraph/sourcegraph/dev/buildchecker -v
func TestSlackSummary(t *testing.T) {
	t.Run("unlocked", func(t *testing.T) {
		s := slackSummary(false, "main", "#buildkite-main", []CommitInfo{})
		t.Log(s)
		assert.Contains(t, s, "unlocked")
		assert.Contains(t, s, "main")
	})

	t.Run("locked", func(t *testing.T) {
		s := slackSummary(true, "main", "#buildkite-main", []CommitInfo{
			{Commit: "a", Author: "bob", AuthorSlackID: "123", BuildNumber: 3, BuildURL: "https://sourcegraph.com", BuildCreated: time.Now()},
			{Commit: "b", Author: "alice", AuthorSlackID: "124", BuildNumber: 2, BuildURL: "https://sourcegraph.com", BuildCreated: time.Now().Add(-1)},
			{Commit: "c", Author: "no_slack", AuthorSlackID: "", BuildNumber: 1, BuildURL: "https://sourcegraph.com", BuildCreated: time.Now().Add(-2)},
		})
		t.Log(s)

		assert.Contains(t, s, "locked")
		assert.Contains(t, s, "main")
		// If Slack user is available, mention only the Slack user
		assert.Contains(t, s, "<@123>")
		assert.Contains(t, s, "<@124>")
		// If no Slack user is available, note the author (user not found is implicit)
		assert.Contains(t, s, "no_slack")
		// Includes build number and URL
		assert.Contains(t, s, "build 1")
		assert.Contains(t, s, "build 2")
		assert.Contains(t, s, "build 3")
		assert.Contains(t, s, "https://sourcegraph.com")
		// Mention the slack channel for discussion
		assert.Contains(t, s, "#buildkite-main")
	})
}
