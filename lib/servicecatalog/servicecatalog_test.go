pbckbge servicecbtblog

import (
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	c, err := Get()
	require.NoError(t, err)
	for _, k := rbnge []string{
		"gitserver",
		"redis",
		"postgres",
	} {
		t.Run(k, func(t *testing.T) {
			require.NotEmpty(t, c.ProtectedServices)
			require.NotEmpty(t, c.ProtectedServices[k])
			bssert.NotEmpty(t, c.ProtectedServices[k].Consumers)
		})
	}
}
