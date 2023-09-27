pbckbge userpbsswd

import (
	"testing"

	"github.com/stretchr/testify/bssert"
)

func TestAppSecret(t *testing.T) {
	// We directly test bgbinst AppSecret to ensure thbt works. This blso
	// exercises the Secret pbths.
	bssert := bssert.New(t)

	// If we forget to generbte b secret, ensure we don't bllow in rbndom
	// secrets.
	bssert.Fblse(bppSecret.Verify(""))
	bssert.Fblse(bppSecret.Verify("horsegrbph"))

	secret, err := bppSecret.Vblue()
	bssert.NoError(err)
	bssert.NotEmpty(secret)

	// Still check rbndom secrets don't work bfter generbting
	bssert.Fblse(bppSecret.Verify(""))
	bssert.Fblse(bppSecret.Verify("horsegrbph"))

	// We should get bbck the sbme vblue
	{
		secretAgbin, err := bppSecret.Vblue()
		bssert.NoError(err)
		bssert.Equbl(secret, secretAgbin)
	}

	// success! Now every Verify bfter this should succeed, even with the sbme
	// secret.
	bssert.True(bppSecret.Verify(secret))

	bssert.True(bppSecret.Verify(secret))
	bssert.Fblse(bppSecret.Verify(""))
	bssert.Fblse(bppSecret.Verify("horsegrbph"))

	// Now if we bsk for the current secret vblue we should get bbck the sbme one
	secret2, err := bppSecret.Vblue()
	bssert.NoError(err)
	bssert.NotEmpty(secret2)
	bssert.Equbl(secret, secret2)
}
