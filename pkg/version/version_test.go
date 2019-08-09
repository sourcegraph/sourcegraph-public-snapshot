package version

import "testing"

func TestVersion(t *testing.T) {
	t.Run("dev", func(t *testing.T) {
		Mock(devVersion)
		if got, want := Version(), devVersion; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("non-dev", func(t *testing.T) {
		Mock("1.2.3")
		if got, want := Version(), "1.2.3"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestIsDev(t *testing.T) {
	tests := map[string]bool{
		devVersion: true,
		"1.2.3":    false,
	}
	for version, want := range tests {
		if got := IsDev(version); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_963(size int) error {
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
