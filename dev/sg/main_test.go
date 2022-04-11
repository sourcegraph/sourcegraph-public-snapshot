package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestAppRun(t *testing.T) {
	assert.NoError(t, sg.Run([]string{"help"}))
}

func TestCommandFormatting(t *testing.T) {
	for _, cmd := range sg.Commands {
		testCommandFormatting(t, cmd)
		// for top-level commands, also require a category
		assert.NotEmptyf(t, cmd.Category, "top-level command %s Category should be set", cmd.Name)
	}
}

func testCommandFormatting(t *testing.T, cmd *cli.Command) {
	t.Run(cmd.Name, func(t *testing.T) {
		assert.NotEmpty(t, cmd.Usage, "Usage should be set")
		assert.False(t, strings.HasSuffix(cmd.Usage, "."), "Usage should not end with period")

		for _, subCmd := range cmd.Subcommands {
			testCommandFormatting(t, subCmd)
		}
	})
}
