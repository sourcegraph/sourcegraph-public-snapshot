package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"testing"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"golang.org/x/net/context"
)

func ExamplePublicKeyToHash() {
	fmt.Println(publicKeyToHash([]byte("sample-public-key-like-data")))
	fmt.Println(publicKeyToHash([]byte("another-public-key-like-data")))

	// Output:
	// W01LbPOaqmRk4cRUZyUfV7E2E6s
	// e769jyY2s86wd_tVaYNidSNlU04
}

func TestUserKeys(t *testing.T) {
	var keys = map[int32][]byte{
		1: []byte("sample-public-key-for-user-1"),
		2: []byte("sample-public-key-for-user-2"),
		3: []byte("sample-public-key-for-user-3"),
	}

	tempDir, err := ioutil.TempDir("", "sourcegraph_userkeys_test_")
	if err != nil {
		t.Fatal(err)
	}

	s := &userKeys{
		dir: tempDir,
	}

	ctx := (context.Context)(nil)

	// Add keys.
	for uid, key := range keys {
		err := s.AddKey(ctx, uid, sourcegraph.SSHPublicKey{key})
		if err != nil {
			t.Error(err)
		}
	}

	// Check that users can be found via their keys.
	for uid, key := range keys {
		userSpec, err := s.LookupUser(ctx, sourcegraph.SSHPublicKey{key})
		if err != nil {
			t.Error(err)
		}
		if got, want := userSpec.UID, uid; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	// Delete a user's key.
	const deletedUID = 2
	err = s.DeleteKey(ctx, deletedUID)
	if err != nil {
		t.Error(err)
	}

	// Check that non-deleted user's key is still there.
	{
		const keptUID int32 = 1
		userSpec, err := s.LookupUser(ctx, sourcegraph.SSHPublicKey{keys[keptUID]})
		if err != nil {
			t.Error(err)
		}
		if got, want := userSpec.UID, keptUID; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	// Check that deleted user's key is no longer present.
	_, err = s.LookupUser(ctx, sourcegraph.SSHPublicKey{keys[deletedUID]})
	if err != nil {
		// Ok.
	} else {
		t.Errorf("expected error, this user's key should be deleted by now")
	}

	// Delete the rest.
	err = s.DeleteKey(ctx, 1)
	if err != nil {
		t.Error(err)
	}
	err = s.DeleteKey(ctx, 3)
	if err != nil {
		t.Error(err)
	}

	// Deleting the key of a user that doesn't have a key should error.
	err = s.DeleteKey(ctx, 1)
	if err != nil {
		// Ok.
	} else {
		t.Error("expected an error because key doesn't exist")
	}

	// Make sure empty folders are cleaned up.
	if names := readDirNames(tempDir); len(names) != 0 {
		t.Errorf("expected store to be empty because all keys deleted, but it has files: %v", names)
	}
}

func readDirNames(dir string) []string {
	f, err := os.Open(dir)
	if err != nil {
		panic(err)
	}
	names, err := f.Readdirnames(0)
	f.Close()
	if err != nil {
		panic(err)
	}
	sort.Strings(names)
	return names
}
