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
