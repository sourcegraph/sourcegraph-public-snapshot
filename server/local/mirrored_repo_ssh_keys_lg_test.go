// +build exectest,pgsqltest

package local_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/server/testserver"
)

func TestMirroredRepoSSHKeys_lg(t *testing.T) {
	if testserver.Store != "pgsql" {
		t.Skip()
	}

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{DisableAccessControl: true},
	)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	repo, err := a.Client.Repos.Create(ctx, &sourcegraph.ReposCreateOp{
		URI:      "r/r",
		VCS:      "git",
		CloneURL: "http://example.com/dummy.git",
		Mirror:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("created repo %q", repo.URI)

	// Generate new keypair.
	const privKeyBits = 2048
	wantPrivKey, err := rsa.GenerateKey(rand.Reader, privKeyBits)
	if err != nil {
		t.Fatal(err)
	}
	wantKeyPEMBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(wantPrivKey),
	})

	_, err = a.Client.MirroredRepoSSHKeys.Create(ctx, &sourcegraph.MirroredRepoSSHKeysCreateOp{
		Repo: sourcegraph.RepoSpec{URI: "r/r"},
		Key:  sourcegraph.SSHPrivateKey{PEM: wantKeyPEMBytes},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("created key %s...", wantKeyPEMBytes[:100])

	privKey, err := a.Client.MirroredRepoSSHKeys.Get(ctx, &sourcegraph.RepoSpec{URI: "r/r"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(privKey.PEM, wantKeyPEMBytes) {
		t.Errorf("got %s, want %s", privKey.PEM, wantKeyPEMBytes)
	}

	if _, err := a.Client.MirroredRepoSSHKeys.Delete(ctx, &sourcegraph.RepoSpec{URI: "r/r"}); err != nil {
		t.Fatal(err)
	}
	t.Log("deleted key")

	if _, err := a.Client.MirroredRepoSSHKeys.Get(ctx, &sourcegraph.RepoSpec{URI: "r/r"}); grpc.Code(err) != codes.NotFound {
		t.Errorf("after deleting, Get returned err == %v, want NotFound", err)
	}
}
