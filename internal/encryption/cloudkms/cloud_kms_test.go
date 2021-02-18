//+build test_cloudkms

package cloudkms

import (
	"context"
	"testing"
)

func TestRoundtrip(t *testing.T) {
	ctx := context.Background()
	k, err := NewKey(ctx, "projects/sourcegraph-dev/locations/global/keyRings/arussellsaw-test/cryptoKeys/testing")
	if err != nil {
		t.Fatal(err)
	}
	ct, err := k.Encrypt(ctx, []byte("test1234"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := k.Decrypt(ctx, ct)
	if err != nil {
		t.Fatal(err)
	}
	if res.Secret() != "test1234" {
		t.Fatalf("expected %s, got %s", "test1234", res.Secret())
	}
}
