package env

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEnvironMap(t *testing.T) {
	environ := []string{
		"FOO=bar",
		"BAZ=",
	}
	want := map[string]string{
		"FOO": "bar",
		"BAZ": "",
	}
	got := environMap(environ)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("mismatch (-want, +got):\n%s", diff)
	}
}

func TestLock(t *testing.T) {
	// Test that calling lock won't panic.
	Lock()
}

func TestGet(t *testing.T) {
	reset := func(osEnviron map[string]string) {
		env = nil
		environ = osEnviron
		locked = false
		expvarPublish = false // avoid "Reuse of exported var name" panic from package expvar
	}
	t.Cleanup(func() { reset(nil) })

	t.Run("normal", func(t *testing.T) {
		reset(map[string]string{"B": "z"})

		a := Get("A", "x", "foo")
		b := Get("B", "y", "bar")
		b2 := Get("B", "y", "bar")
		Lock()
		if want := "x"; a != want {
			t.Errorf("got A == %q, want %q", a, want)
		}
		if want := "z"; b != want {
			t.Errorf("got B == %q, want %q", b, want)
		}
		if want := "z"; b2 != want {
			t.Errorf("got B2 == %q, want %q", b2, want)
		}
	})

	t.Run("conflicting registrations", func(t *testing.T) {
		reset(nil)

		Get("A", "x", "foo")
		defer func() {
			if e := recover(); e == nil {
				t.Error("want panic")
			}
		}()
		Get("A", "y", "bar")
		t.Error("want panic")
	})
}
