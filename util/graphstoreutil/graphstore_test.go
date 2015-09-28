package graphstoreutil

import "testing"

func TestParseS3Bucket(t *testing.T) {
	tests := []struct {
		s       string
		wantURL string
	}{}
	for _, test := range tests {
		u, err := parseS3Bucket(test.s)
		if err != nil {
			t.Errorf("%s: error: %s", test.s, err)
			continue
		}
		if u.String() != test.wantURL {
			t.Errorf("%s: got %q, want %q", test.s, u, test.wantURL)
		}
	}
}
