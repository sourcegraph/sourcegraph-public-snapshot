package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestMakeSuggestions(t *testing.T) {
	cmds := []*cli.Command{
		{Name: "elloo"},
		{Name: "helo"},
		{Name: "totally unrelated"},
		{Name: "hello"},
		{Name: "hlloo"},
	}
	t.Run("restrict suggestions", func(t *testing.T) {
		suggestions := makeSuggestions(cmds, "hello", 0.3, 2)
		assert.Len(t, suggestions, 2)
		assert.Equal(t, "hello", suggestions[0].name)
		assert.Equal(t, "helo", suggestions[1].name)
	})
	t.Run("all suggestions", func(t *testing.T) {
		suggestions := makeSuggestions(cmds, "hello", 0.3, 999)
		assert.Len(t, suggestions, len(cmds)-1)
		assert.Equal(t, "hello", suggestions[0].name)
		assert.Equal(t, "helo", suggestions[1].name)
	})
}
