package accesstoken

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
)

func TestParseSelfSignedToken(t *testing.T) {
	idkey.SetTestEnvironment(512)
	k, err := idkey.Generate()
	if err != nil {
		t.Fatal(err)
	}

	tok, err := NewSelfSigned(k, nil, nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ParseAndVerify(k, tok.AccessToken); err != nil {
		t.Fatal(err)
	}
}
