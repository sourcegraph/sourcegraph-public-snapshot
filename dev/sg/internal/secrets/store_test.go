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
			t.Fatalf("want no error, got %q", err)
		}

		want := data
		got := mySecrets{}
		err = store.Get("foo", &got)
		if err != nil {
			t.Fatalf("want no error when getting secret, but got: %q", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("wrong secret data. (-want +got):\n%s", diff)
		}
	})

	t.Run("LoadFile returns an error on invalid JSON", func(t *testing.T) {
		f, err := os.CreateTemp(os.TempDir(), "secrets*.json")
		if err != nil {
			t.Fatalf("couldn't create temp secret file: %q", err)
		}
		if _, err := f.WriteString(`{"foo":1`); err != nil {
			t.Fatalf("couldn't write in temp secret file: %q", err)
		}
		f.Close()
		filepath := f.Name()
		t.Cleanup(func() {
			_ = os.Remove(filepath)
		})

		_, err = LoadFromFile(filepath)
		if err == nil {
			t.Fatal("want an error but got none")
		}
	})

	t.Run("LoadFile doesn't fail when file is empty", func(t *testing.T) {
		f, err := os.CreateTemp(os.TempDir(), "secrets*.json")
		if err != nil {
			t.Fatalf("couldn't create temp secret file: %q", err)
		}
		f.Close()
		filepath := f.Name()
		t.Cleanup(func() {
			_ = os.Remove(filepath)
		})

		got, err := LoadFromFile(filepath)
		if err != nil {
			t.Fatalf("want no error when loading an empty file, but got %q instead", err)
		}
		if got == nil {
			t.Fatal("want store to not be nil")
		}
	})

	t.Run("SaveFile and LoadFile", func(t *testing.T) {
		f, err := os.CreateTemp(os.TempDir(), "secrets*.json")
		if err != nil {
			t.Fatalf("couldn't create temp secret file: %q", err)
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
			t.Fatalf("couldn't load temp secret file: %q", err)
		}
		data := map[string]any{"key": "val"}
		if err := s.Put("foo", data); err != nil {
			t.Fatalf("want no error when putting secret, got %q", err)
		}
		err = s.SaveFile()
		if err != nil {
			t.Fatalf("failed to save secrets: %q", err)
		}

		// Fetch it back and compare
		got, err := LoadFromFile(filepath)
		if err != nil {
			t.Fatalf("couldn't load temp secret file: %q", err)
		}
		if diff := cmp.Diff(s.m, got.m); diff != "" {
			t.Fatalf("(-want +got):\n%s", diff)
		}
	})
}
