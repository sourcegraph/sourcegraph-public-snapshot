// +build pgsqltest

package pgsql

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestRepos_CreateAndGetRemoteKey(t *testing.T) {
	t.Parallel()

	var s MirroredRepoSSHKeys
	ctx, done := testContext()
	defer done()

	wantPEM := []byte(testRSAKeyPEM)
	block, _ := pem.Decode(wantPEM)
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	repo := &sourcegraph.Repo{URI: "x.com/r"}

	// Create key.
	if err := s.Create(ctx, repo.URI, privKey); err != nil {
		t.Fatal(err)
	}

	// Get key.
	gotPEM, err := s.GetPEM(ctx, repo.URI)
	if err != nil {
		t.Fatal(err)
	}

	// Compare.
	if !bytes.Equal(gotPEM, wantPEM) {
		t.Errorf("got PEM %q, want PEM %q", gotPEM, wantPEM)
	}
}

func TestRepos_GetRemoteKey_none(t *testing.T) {
	t.Parallel()

	var s MirroredRepoSSHKeys
	ctx, done := testContext()
	defer done()

	repo := &sourcegraph.Repo{URI: "x.com/r"}

	// Get a key that doesn't exist.
	got, err := s.GetPEM(ctx, repo.URI)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Errorf("got err %q, want nil", err)
	}
	if got != nil {
		t.Errorf("got PEM %v, want nil", got)
	}
}

const testRSAKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIBOwIBAAJBALdGZxkXDAjsYk10ihwU6Id2KeILz1TAJuoq4tOgDWxEEGeTrcld
r/ZwVaFzjWzxaf6zQIJbfaSEAhqD5yo72+sCAwEAAQJBAK8PEVU23Wj8mV0QjwcJ
tZ4GcTUYQL7cF4+ezTCE9a1NrGnCP2RuQkHEKxuTVrxXt+6OF15/1/fuXnxKjmJC
nxkCIQDaXvPPBi0c7vAxGwNY9726x01/dNbHCE0CBtcotobxpwIhANbbQbh3JHVW
2haQh4fAG5mhesZKAGcxTyv4mQ7uMSQdAiAj+4dzMpJWdSzQ+qGHlHMIBvVHLkqB
y2VdEyF7DPCZewIhAI7GOI/6LDIFOvtPo6Bj2nNmyQ1HU6k/LRtNIXi4c9NJAiAr
rrxx26itVhJmcvoUhOjwuzSlP2bE5VHAvkGB352YBg==
-----END RSA PRIVATE KEY-----
`
