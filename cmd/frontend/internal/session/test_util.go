package session

import (
	"os"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

func ResetMockSessionStore(t *testing.T) (cleanup func()) {
	var err error
	tempdir, err := os.MkdirTemp("", "sourcegraph-oidc-test")
	if err != nil {
		return func() {}
	}

	defer func() {
		if err != nil {
			os.RemoveAll(tempdir)
		}
	}()

	SetSessionStore(sessions.NewFilesystemStore(tempdir, securecookie.GenerateRandomKey(2048)))
	return func() {
		os.RemoveAll(tempdir)
	}
}
