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
			{Commit: "a", Author: "bob", AuthorSlackUser: "123"},
			{Commit: "b", Author: "alice", AuthorSlackUser: "124"},
			{Commit: "c", Author: "no_github", AuthorSlackUser: ""},
		})
		t.Log(s)
		assert.Contains(t, s, "locked")
		assert.Contains(t, s, "bob")
		assert.Contains(t, s, "<@123>")
		assert.Contains(t, s, "alice")
		assert.Contains(t, s, "<@124>")
		assert.Contains(t, s, "no_github")
		assert.Contains(t, s, ":warning: Cannot find Slack user :warning:")
	})
}
