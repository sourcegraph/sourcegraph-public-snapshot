package returnto

import "testing"

func TestCheckSafe(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"foo", false},
		{"foo/bar", false},
		{"/foo/bar", false},
		{"http://foo", true},
		{"https://foo", true},
		{"//foo", true},
	}

	for _, test := range tests {
		err := CheckSafe(test.url)
		if gotErr := (err != nil); gotErr != test.wantErr {
			t.Errorf("got error %q, want error is (%v)", err, test.wantErr)
		}
	}
}
