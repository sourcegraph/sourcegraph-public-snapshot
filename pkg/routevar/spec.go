package routevar

import "fmt"

// InvalidError occurs when a spec string is invalid.
type InvalidError struct {
	Type  string // Repo, etc.
	Input string // the original string input
	Err   error  // underlying error (nil for routine regexp match failures)
}

func (e InvalidError) Error() string {
	str := fmt.Sprintf("invalid input for %s: %q", e.Type, e.Input)
	if e.Err != nil {
		str += " " + e.Err.Error()
	}
	return str
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_888(size int) error {
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
