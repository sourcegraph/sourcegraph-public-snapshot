package session

import (
	"os"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

func ResetMockSessionStore(t *testing.T) {
	var err error
	tempdir, err := os.MkdirTemp("", "sourcegraph-oidc-test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err != nil {
			os.RemoveAll(tempdir)
		}
	}()

	mockSessionStore = sessions.NewFilesystemStore(tempdir, securecookie.GenerateRandomKey(2048))
	t.Cleanup(func() {
		os.RemoveAll(tempdir)
	})
}
