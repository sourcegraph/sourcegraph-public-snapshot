package analytics

import (
	"os"
	"path"
	"testing"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

func TestGetEmail(t *testing.T) {
	os.Setenv("HOME", t.TempDir())
	sgHome, err := root.GetSGHomePath()
	if err != nil {
		t.Fatal(err)
	}

	whoamiPath := path.Join(sgHome, "whoami.json")

	t.Run("whoami doesnt exist", func(t *testing.T) {
		if email := emailfunc(); email != "anonymous" {
			t.Fatal("expected anonymous")
		}
	})

	t.Run("misformed whoami", func(t *testing.T) {
		err := os.WriteFile(whoamiPath, []byte("{"), 0o700)
		if err != nil {
			t.Fatal(err)
		}
		if email := emailfunc(); email != "anonymous" {
			t.Fatal("expected anonymous")
		}
	})

	t.Run("well formed", func(t *testing.T) {
		err := os.WriteFile(whoamiPath, []byte(`{"email":"bananaphone@gmail.com"}`), 0o700)
		if err != nil {
			t.Fatal(err)
		}
		if email := emailfunc(); email != "bananaphone@gmail.com" {
			t.Fatal("expected bananaphone@gmail.com")
		}
	})
}
