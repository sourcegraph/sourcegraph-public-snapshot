package auth

import "testing"

func TestSafeRedirectURL(t *testing.T) {
	tests := map[string]string{
		"":               "/",
		"/":              "/",
		"a@b.com:c":      "/",
		"a@b.com/c":      "/",
		"//a":            "/",
		"http://a.com/b": "/b",
		"//a.com/b":      "/b",
		"//a@b.com/c":    "/c",
		"/a?b":           "/a?b",
	}
	for input, want := range tests {
		got := SafeRedirectURL(input)
		if got != want {
			t.Errorf("%q: got %q, want %q", input, got, want)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_6(size int) error {
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
