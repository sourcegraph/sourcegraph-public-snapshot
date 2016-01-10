package hello

import "testing"

func TestWorld(t *testing.T) {
	want := "Hello, world!"
	if got := World(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Alice", "Hello, Alice!"},
		{"", "Hello!"},
	}
	for _, test := range tests {
		got := Name(test.name)
		if got != test.want {
			t.Errorf("got %q, want %q", got, test.want)
		}
	}
}
