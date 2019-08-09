package pathmatch

import "testing"

func TestCompilePattern(t *testing.T) {
	tests := []struct {
		pattern string
		options CompileOptions
		want    map[string]bool
		wantErr bool
	}{
		{
			pattern: `README.md`,
			options: CompileOptions{},
			want: map[string]bool{
				"README.md": true,
				"main.go":   false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			match, err := CompilePattern(test.pattern, test.options)
			if (err != nil) != test.wantErr {
				t.Errorf("got error %v, want %v", err, test.wantErr)
			}
			if err != nil {
				return
			}
			for path, want := range test.want {
				got := match.MatchPath(path)
				if got != want {
					t.Errorf("path %q: got %v, want %v", path, got, want)
				}

				if got2 := match.Copy().MatchPath(path); got != got2 {
					t.Errorf("path %q: after copy, got %v, want %v", path, got2, got)
				}
			}
		})
	}
}

func TestCompilePathPatterns(t *testing.T) {
	match, err := CompilePathPatterns([]string{`main\.go`, `m`}, `README\.md`, CompileOptions{RegExp: true})
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]bool{
		"README.md": false,
		"main.go":   true,
	}
	for path, want := range want {
		got := match.MatchPath(path)
		if got != want {
			t.Errorf("path %q: got %v, want %v", path, got, want)
			continue
		}

		if got2 := match.Copy().MatchPath(path); got != got2 {
			t.Errorf("path %q: after copy, got %v, want %v", path, got2, got)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_856(size int) error {
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
