package graphqlbackend

import "testing"

func TestGitCommitBody(t *testing.T) {
	tests := map[string]string{
		"hello":               "",
		"hello\n":             "",
		"hello\n\n":           "",
		"hello\nworld":        "world",
		"hello\n\nworld":      "world",
		"hello\n\nworld\nfoo": "world\nfoo",
	}
	for input, want := range tests {
		got := gitCommitBody(input)
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_142(size int) error {
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
