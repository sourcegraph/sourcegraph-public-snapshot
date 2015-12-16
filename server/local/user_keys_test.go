// +build exectest

package local_test

import (
	"bytes"
	"testing"

	"golang.org/x/crypto/ssh"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/testserver"
)

func TestAddKey(t *testing.T) {
	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{DisableAccessControl: true},
	)

	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	_, err := a.Client.UserKeys.ClearKeys(ctx, &pbtypes.Void{})
	if err != nil {
		t.Fatal(err)
	}

	keyBytes := []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCmrKBS1TCw8RVW4WKeBg9tabk0QjxqW5YB5xQJzLEhBRrIQ7nrFX2D9LfBwbJ0m+0Lc5u9Fpbf8J8QPMlulQB0E573euMP1S/NLCuzao1PenPlUH/Jv5pIIALsMKcgz7jrJ3PLTC+IjD9pXerEN90m4hKVDOwg+GcznzRH4WWLEBa8nJzY2rP78EGE937xLapEp5mPGlHGNzloQcsYJ7fCZf0M0ncc6IrubSTIVwzacDXbUJKvs9T8Vfu3D7WYjj6ed11vwcDvjYIP7sgPfdwhHbTBJzf1walDb8zy0RJX8BLbhFm55zXyI2xDETsxAXPIjOAFN9GzaKi7UB0O/95B m@rtin.so")
	key, _, _, _, err := ssh.ParseAuthorizedKey(keyBytes)

	// Add a key
	_, err = a.Client.UserKeys.AddKey(ctx, &sourcegraph.SSHPublicKey{Key: key.Marshal(), Name: "test key"})
	if err != nil {
		t.Fatal(err)
	}

	// Get and validate keys
	keyList, err := a.Client.UserKeys.ListKeys(ctx, &pbtypes.Void{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v len: %d", keyList.SSHKeys, len(keyList.SSHKeys))
	keys := keyList.SSHKeys
	if len(keys) != 1 {
		t.Fatal("Invalid key count")
	}

	for _, key := range keys {
		if key.Id != 0 {
			t.Fatal("invalid key id")
		}

		if key.Name != "test key" {
			t.Fatalf("invalid key name: %s", key.Name)
		}

		// Take out the e-mail from our original key
		addedKey := bytes.TrimSuffix(keyBytes, []byte(" m@rtin.so"))
		trimmedKey := bytes.TrimSpace(key.Key)
		if !bytes.Equal(trimmedKey, addedKey) {
			t.Logf("\n%s\n%s", addedKey, trimmedKey)
			t.Fatal("invalid key bytes")
		}
	}
}
