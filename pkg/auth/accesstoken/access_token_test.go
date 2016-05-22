package accesstoken

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
)

func TestAsymmetricToken(t *testing.T) {
	testToken(t, true)
}

func TestSymmetricToken(t *testing.T) {
	testToken(t, false)
}

func testToken(t *testing.T, useAsymmetricEnc bool) {
	idkey.SetTestEnvironment(512)
	k, err := idkey.Generate()
	if err != nil {
		t.Fatal(err)
	}

	tok, err := New(k, nil, nil, 0, useAsymmetricEnc)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ParseAndVerify(k, tok.AccessToken); err != nil {
		t.Fatal(err)
	}

	k2, err := idkey.Generate()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ParseAndVerify(k2, tok.AccessToken); err == nil {
		t.Fatal("error expected")
	}
}
