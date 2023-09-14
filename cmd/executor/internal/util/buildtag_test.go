//go:build !shell
// +build !shell

package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
)

func TestHasShellBuildTag(t *testing.T) {
	assert.False(t, util.HasShellBuildTag())
}
