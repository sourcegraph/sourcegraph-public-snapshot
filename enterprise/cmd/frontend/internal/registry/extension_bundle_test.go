package registry

import "testing"

func TestParseExtensionBundleFilename(t *testing.T) {
	tests := map[string]int64{
		"123.js":        123,
		"123-a-b.js":    123,
		"123-a-b-c.map": 123,
	}
	for input, want := range tests {
		got, err := parseExtensionBundleFilename(input)
		if err != nil {
			t.Errorf("%q: %s", input, err)
		}
		if got != want {
			t.Errorf("%q: got %d, want %d", input, got, want)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_677(size int) error {
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
