package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlackSummary(t *testing.T) {
	t.Run("unlocked", func(t *testing.T) {
		s := slackSummary(false, []CommitInfo{})
		t.Log(s)
		assert.Contains(t, s, "unlocked")
	})

	t.Run("locked", func(t *testing.T) {
		s := slackSummary(true, []CommitInfo{
			{Commit: "a", Author: "bob", AuthorSlackID: "123"},
			{Commit: "b", Author: "alice", AuthorSlackID: "124"},
			{Commit: "c", Author: "no_slack", AuthorSlackID: ""},
		})
		t.Log(s)
		assert.Contains(t, s, "locked")
		// If Slack user is available, mention only the Slack user
		assert.Contains(t, s, "<@123>")
		assert.Contains(t, s, "<@124>")
		// If no Slack user is available, note the author (user not found is implicit)
		assert.Contains(t, s, "no_slack")
	})
}
