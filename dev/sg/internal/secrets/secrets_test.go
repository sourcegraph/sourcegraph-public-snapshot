package secrets

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var input = `{"foo": {"bar": "baz"}}`

func TestSecrets(t *testing.T) {
	t.Run("Load", func(t *testing.T) {
		s, err := Load(strings.NewReader(input))
		if err != nil {
			t.Fatalf("%v", err)
		}
		if s == nil {
			t.Fatalf("want secrets, got nil")
		}
		if got, ok := s.m["foo"]; !ok || got == nil {
			t.Fatal("want secret to include foo, found nothing")
		}
	})

	t.Run("Put and Get", func(t *testing.T) {
		data := map[string]string{"key": "val"}
		store := New()
		err := store.Put("foo", data)
		if err != nil {
			t.Fatalf("want no error, got %v", err)
		}

		want := data
		got, err := store.Get("foo")
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("wrong secret data. (-want +got):\n%s", diff)
		}
	})

	t.Run("Save and Load", func(t *testing.T) {
		f, err := os.CreateTemp(os.TempDir(), "secrets*.json")
		if err != nil {
			t.Fatalf("%v", err)
		}
		filepath := f.Name()
		t.Cleanup(func() {
			_ = os.Remove(filepath)
		})

		// Assign a secret and save it
		s := New()
		data := map[string]interface{}{"key": "val"}
		s.Put("foo", data)
		err = s.Save(f)
		if err != nil {
			t.Fatalf("%v", err)
		}
		f.Close()

		// Fetch it back and compare
		got, err := LoadFile(filepath)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if diff := cmp.Diff(s.m, got.m); diff != "" {
			t.Fatalf("(-want +got):\n%s", diff)
		}
	})
}
