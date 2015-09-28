package accesstoken

import (
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/auth/idkey"
)

func TestParseSelfSignedToken(t *testing.T) {
	idkey.Bits = 512 // small for testing
	k, err := idkey.Generate()
	if err != nil {
		t.Fatal(err)
	}

	tok, err := NewSelfSigned(k, nil, nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	ctx := idkey.NewContext(context.Background(), k)

	if _, _, err := ParseAndVerify(ctx, tok.AccessToken); err != nil {
		t.Fatal(err)
	}
}
