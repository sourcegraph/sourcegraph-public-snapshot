package userpasswd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppSecret(t *testing.T) {
	// We directly test against AppSecret to ensure that works. This also
	// exercises the Secret paths.
	assert := assert.New(t)

	// If we forget to generate a secret, ensure we don't allow in random
	// secrets.
	assert.False(appSecret.Verify(""))
	assert.False(appSecret.Verify("horsegraph"))

	secret, err := appSecret.Value()
	assert.NoError(err)
	assert.NotEmpty(secret)

	// Still check random secrets don't work after generating
	assert.False(appSecret.Verify(""))
	assert.False(appSecret.Verify("horsegraph"))

	// We should get back the same value
	{
		secretAgain, err := appSecret.Value()
		assert.NoError(err)
		assert.Equal(secret, secretAgain)
	}

	// success! Now every Verify after this should succeed, even with the same
	// secret.
	assert.True(appSecret.Verify(secret))

	assert.True(appSecret.Verify(secret))
	assert.False(appSecret.Verify(""))
	assert.False(appSecret.Verify("horsegraph"))

	// Now if we ask for the current secret value we should get back the same one
	secret2, err := appSecret.Value()
	assert.NoError(err)
	assert.NotEmpty(secret2)
	assert.Equal(secret, secret2)
}
