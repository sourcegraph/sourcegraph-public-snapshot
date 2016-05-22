package sharedsecret

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
)

func TestToken(t *testing.T) {
	idkey.SetTestEnvironment(512)
	k, err := idkey.Generate()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := TokenSource(k, "*").Token(); err != nil {
		t.Fatal(err)
	}
}

func TestShortToken(t *testing.T) {
	idkey.SetTestEnvironment(512)
	k, err := idkey.Generate()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ShortTokenSource(k, "*").Token(); err != nil {
		t.Fatal(err)
	}
}

func TestShortTokenLength(t *testing.T) {
	idkey.SetTestEnvironment(512)
	k, err := idkey.Generate()
	if err != nil {
		t.Fatal(err)
	}

	tok, err := ShortTokenSource(k, "*").Token()
	if err != nil {
		t.Fatal(err)
	}
	if max := 255; len(tok.AccessToken) > max {
		t.Errorf("sharedsecret token too long (%d chars). Because this token is used as a git HTTP Basic Auth password by the worker, it must fit within Git <2.0 (libcurl's) length restriction of %d.", len(tok.AccessToken), max)
	} else {
		t.Logf("sharedsecret access token length is %d", len(tok.AccessToken))
	}
}
