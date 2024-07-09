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

	t.Run("when CI=true identity is ci@sourcegraph.com", func(t *testing.T) {
		os.Setenv("CI", "true")
		t.Cleanup(func() {
			os.Unsetenv("CI")
		})
		if email := emailfunc(); email != "ci@sourcegraph.com" {
			t.Fatal("expected 'ci@sourcegraph.com' identity when CI=true")
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
