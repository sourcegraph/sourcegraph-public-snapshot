package privategoogleaccess

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

func TestGetPrivateGoogleIPs(t *testing.T) {
	ips, err := getPrivateGoogleIPs()
	require.NoError(t, err)
	autogold.Expect([]string{
		"199.36.153.8", "199.36.153.9", "199.36.153.10",
		"199.36.153.11",
	}).Equal(t, ips)
}
