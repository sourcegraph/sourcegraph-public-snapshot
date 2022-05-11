package secrets

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type mySecrets struct {
	ID     string
	Secret string
}

func TestSecrets(t *testing.T) {
	t.Run("Put and Get", func(t *testing.T) {
		data := mySecrets{ID: "foo", Secret: "bar"}
		store := newStore("")
		err := store.Put("foo", data)
		if err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		want := data
		got := mySecrets{}
		err = store.Get("foo", &got)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("wrong secret data. (-want +got):\n%s", diff)
		}
	})

	t.Run("SaveFile and LoadFile", func(t *testing.T) {
		f, err := os.CreateTemp(os.TempDir(), "secrets*.json")
		if err != nil {
			t.Fatalf("%v", err)
		}
		f.Close()
		filepath := f.Name()
		_ = os.Remove(filepath) // we just want the path, not the file
		t.Cleanup(func() {
			_ = os.Remove(filepath)
		})

		// Assign a secret and save it
		s, err := LoadFromFile(filepath)
		if err != nil {
			t.Fatalf("%v", err)
		}
		data := map[string]any{"key": "val"}
		s.Put("foo", data)
		err = s.SaveFile()
		if err != nil {
			t.Fatalf("%v", err)
		}

		// Fetch it back and compare
		got, err := LoadFromFile(filepath)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if diff := cmp.Diff(s.m, got.m); diff != "" {
			t.Fatalf("(-want +got):\n%s", diff)
		}
	})
}
