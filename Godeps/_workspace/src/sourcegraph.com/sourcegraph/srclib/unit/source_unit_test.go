package unit

import "testing"

func TestParseID(t *testing.T) {
	tests := []struct {
		input              string
		wantName, wantType string
	}{
		{"a@b", "a", "b"},
		{"a%2Fb@b", "a/b", "b"},
	}
	for _, test := range tests {
		name, typ, err := ParseID(test.input)
		if err != nil {
			t.Fatal(err)
		}
		if name != test.wantName {
			t.Errorf("got name %q, want %q", name, test.wantName)
		}
		if typ != test.wantType {
			t.Errorf("got type %q, want %q", typ, test.wantType)
		}
	}
}
