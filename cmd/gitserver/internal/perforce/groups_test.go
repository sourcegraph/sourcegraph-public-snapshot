package perforce

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseP4GroupMembers(t *testing.T) {
	groupOut := []byte(`{"Group":"all","MaxLockTime":"unset","MaxOpenFiles":"unset","MaxResults":"unset","MaxScanRows":"unset","Owners0":"admin","PasswordTimeout":"unset","Timeout":"43200","Users0":"admin","Users1":"alice","Users2":"bob","Users3":"buildkite","Users4":"test-perforce"}`)

	users, err := parseP4GroupMembers(groupOut)
	require.NoError(t, err)

	require.Equal(t, []string{"admin", "alice", "bob", "buildkite", "test-perforce"}, users)
}
