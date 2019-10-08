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
	}
}
