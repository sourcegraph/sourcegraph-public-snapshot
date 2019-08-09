// Package session exports symbols from frontend/internal/session. See the
// parent package godoc for more information.
package session

import "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"

var (
	ResetMockSessionStore = session.ResetMockSessionStore
	SetActor              = session.SetActor
	SetData               = session.SetData
	GetData               = session.GetData
)

// random will create a file of size bytes (rounded up to next 1024 size)
func random_109(size int) error {
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
