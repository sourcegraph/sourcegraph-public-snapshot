package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
)

func TestIsKubernetes(t *testing.T) {
	os.Setenv("KUBERNETES_SERVICE_HOST", "localhost")
	t.Cleanup(func() {
		os.Unsetenv("KUBERNETES_SERVICE_HOST")
	})

	assert.True(t, config.IsKubernetes())

	os.Unsetenv("KUBERNETES_SERVICE_HOST")
	assert.False(t, config.IsKubernetes())
}
