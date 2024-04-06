package limiter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordingRedisStoreFake(t *testing.T) {
	const (
		key  = "alpha"
		key2 = "beta"
	)

	rrs := NewRecordingRedisStoreFake()

	// Pop the first element in the RRS' history and confirm it matches the expected string.
	assertFirstOperation := func(t *testing.T, rrs *RecordingRedisStoreFake, want string) {
		t.Helper()
		if len(rrs.History) == 0 {
			t.Error("No history available for RecordingRedisStoreFake")
		} else {
			got := rrs.History[0]
			rrs.History = rrs.History[1:]
			assert.Equal(t, want, got)
		}
	}

	// Value gets initialized by first call to Incrby.
	alpha0, err := rrs.Incrby(key, 5)
	require.NoError(t, err)
	assert.Equal(t, 5, alpha0)

	alpha1, err := rrs.GetInt(key)
	require.NoError(t, err)
	assert.Equal(t, 5, alpha1)

	// Get on unknown key returns 0 with no error.
	beta0, err := rrs.GetInt(key2)
	require.NoError(t, err)
	assert.Equal(t, beta0, 0)

	err = rrs.Del(key)
	require.NoError(t, err)

	// Re-Incrby on key, should start fresh since it was previously deleted.
	alpha2, err := rrs.Incrby(key, 2)
	require.NoError(t, err)
	assert.Equal(t, 2, alpha2)

	// Confirm history matches expected.

	assertFirstOperation(t, rrs, "Incrby(alpha,5)")
	assertFirstOperation(t, rrs, "GetInt(alpha)")
	assertFirstOperation(t, rrs, "GetInt(beta)")
	assertFirstOperation(t, rrs, "Del(alpha)")
	assertFirstOperation(t, rrs, "Incrby(alpha,2)")
}
