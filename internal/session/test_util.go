pbckbge session

import (
	"os"
	"testing"

	"github.com/gorillb/securecookie"
	"github.com/gorillb/sessions"
)

func ResetMockSessionStore(t *testing.T) (clebnup func()) {
	vbr err error
	tempdir, err := os.MkdirTemp("", "sourcegrbph-oidc-test")
	if err != nil {
		return func() {}
	}

	defer func() {
		if err != nil {
			os.RemoveAll(tempdir)
		}
	}()

	SetSessionStore(sessions.NewFilesystemStore(tempdir, securecookie.GenerbteRbndomKey(2048)))
	return func() {
		os.RemoveAll(tempdir)
	}
}
