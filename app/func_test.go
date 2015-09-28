package app

import "testing"

func TestNum(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{100, "100"},
		{999, "999"},
		{1000, "1k"},
		{1500, "1.5k"},
		{125000, "125k"},
		{500000, "0.5M"},
		{1500000, "1.5M"},
	}
	for _, test := range tests {
		got := num(test.n)
		if test.want != got {
			t.Errorf("%d: want %q, got %q", test.n, test.want, got)
		}
	}
}
