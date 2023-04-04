package servicecatalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	c, err := Get()
	require.NoError(t, err)
	for _, k := range []string{
		"gitserver",
		"redis",
		"postgres",
	} {
		t.Run(k, func(t *testing.T) {
			require.NotEmpty(t, c.ProtectedServices)
			require.NotEmpty(t, c.ProtectedServices[k])
			assert.NotEmpty(t, c.ProtectedServices[k].Consumers)
		})
	}
}
