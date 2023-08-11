package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/config"
)

func TestIsKubernetes(t *testing.T) {
	os.Setenv("KUBERNETES_SERVICE_HOST", "localhost")
	os.Setenv("KUBERNETES_SERVICE_PORT", "8000")
	t.Cleanup(func() {
		os.Unsetenv("KUBERNETES_SERVICE_HOST")
		os.Unsetenv("KUBERNETES_SERVICE_PORT")
	})

	assert.True(t, config.IsKubernetes())

	os.Unsetenv("KUBERNETES_SERVICE_HOST")
	assert.False(t, config.IsKubernetes())
}
