package server

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

func (s vcsDependenciesSyncer) runCloneCommand(t *testing.T, examplePackageURL, bareGitDirectory string, dependencies []string) {
	u := vcs.URL{
		URL: url.URL{Path: examplePackageURL},
	}
	s.configDeps = dependencies
	cmd, err := s.CloneCommand(context.Background(), &u, bareGitDirectory)
	assert.Nil(t, err)
	assert.Nil(t, cmd.Run())
}
