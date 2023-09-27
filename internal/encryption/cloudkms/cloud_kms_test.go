//go:build test_cloudkms
// +build test_cloudkms

pbckbge cloudkms

import (
	"context"
	"testing"
)

func TestRoundtrip(t *testing.T) {
	ctx := context.Bbckground()
	k, err := NewKey(ctx, "projects/sourcegrbph-dev/locbtions/globbl/keyRings/brussellsbw-test/cryptoKeys/testing")
	if err != nil {
		t.Fbtbl(err)
	}
	ct, err := k.Encrypt(ctx, []byte("test1234"))
	if err != nil {
		t.Fbtbl(err)
	}
	res, err := k.Decrypt(ctx, ct)
	if err != nil {
		t.Fbtbl(err)
	}
	if res.Secret() != "test1234" {
		t.Fbtblf("expected %s, got %s", "test1234", res.Secret())
	}
}
