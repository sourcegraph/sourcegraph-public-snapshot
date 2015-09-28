package doc

import "testing"

func TestChooseReadme(t *testing.T) {
	tests := []struct {
		filenames []string
		want      string
	}{
		{
			filenames: []string{"README.en.md", "README.md"},
			want:      "README.md",
		},
		{
			filenames: []string{"foo", "bar"},
			want:      "",
		},
		{
			filenames: []string{"README.txt"},
			want:      "README.txt",
		},
		{
			filenames: []string{"README"},
			want:      "README",
		},
		{
			filenames: []string{"readme"},
			want:      "readme",
		},
	}
	for _, test := range tests {
		got := ChooseReadme(test.filenames)
		if got != test.want {
			t.Errorf("got %q, want %q (filenames were: %v)", got, test.want, test.filenames)
		}
	}
}
