package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlackSummary(t *testing.T) {
	t.Run("unlocked", func(t *testing.T) {
		s := slackSummary(false, []CommitInfo{}, "slackToken")
		t.Log(s)
		assert.Contains(t, s, "unlocked")
	})

	t.Run("locked", func(t *testing.T) {
		s := slackSummary(true, []CommitInfo{
			{Commit: "a", Author: "bob"},
			{Commit: "b", Author: "alice"},
		}, "slackToken")
		t.Log(s)
		assert.Contains(t, s, "locked")
		assert.Contains(t, s, "bob")
		assert.Contains(t, s, "alice")
	})
}
