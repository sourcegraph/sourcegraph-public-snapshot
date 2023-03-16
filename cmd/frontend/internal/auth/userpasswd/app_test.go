package userpasswd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppNonce(t *testing.T) {
	// We directly test against AppNonce to ensure that works. This also
	// exercises the Nonce paths.
	assert := assert.New(t)

	// If we forget to generate a nonce, ensure we don't allow in random
	// nonces.
	assert.False(AppNonce.Verify(""))
	assert.False(AppNonce.Verify("horsegraph"))

	nonce, err := AppNonce.Value()
	assert.NoError(err)
	assert.NotEmpty(nonce)

	// Still check random nonces don't work after generating
	assert.False(AppNonce.Verify(""))
	assert.False(AppNonce.Verify("horsegraph"))

	// We should get back the same value since we haven't used it yet
	{
		nonceAgain, err := AppNonce.Value()
		assert.NoError(err)
		assert.Equal(nonce, nonceAgain)
	}

	// success! Now every Verify after this should fail, even with the same
	// nonce.
	assert.True(AppNonce.Verify(nonce))

	assert.False(AppNonce.Verify(nonce))
	assert.False(AppNonce.Verify(""))
	assert.False(AppNonce.Verify("horsegraph"))

	// Now if we ask for the current nonce value we should get back a new one
	nonce2, err := AppNonce.Value()
	assert.NoError(err)
	assert.NotEmpty(nonce2)
	assert.NotEqual(nonce, nonce2)
}
