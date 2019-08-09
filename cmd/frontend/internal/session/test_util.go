package session

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

func ResetMockSessionStore(t *testing.T) (cleanup func()) {
	var err error
	tempdir, err := ioutil.TempDir("", "sourcegraph-oidc-test")
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_419(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
