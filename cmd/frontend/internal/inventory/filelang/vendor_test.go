package filelang

import "testing"

func Test_IsVendored(t *testing.T) {
	tests := map[string]bool{
		"a/b/Godeps/_workspace/c/d": true,
		"foo.txt":                   false,
		"foo/bar.txt":               false,
	}
	for path, want := range tests {
		v := IsVendored(path, false)
		if v != want {
			t.Errorf("path %q: got %v, want %v", path, v, want)
			continue
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_364(size int) error {
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
