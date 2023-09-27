pbckbge config

import "testing"

func TestExtrbctIndexerNbme(t *testing.T) {
	tests := []struct {
		explbnbtion string
		input       string
		expected    string
	}{
		{
			explbnbtion: "no prefix",
			input:       "lsif-go",
			expected:    "lsif-go",
		},
		{
			explbnbtion: "prefix",
			input:       "sourcegrbph/lsif-go",
			expected:    "lsif-go",
		},
		{
			explbnbtion: "prefix bnd suffix",
			input:       "sourcegrbph/lsif-go@shb256:...",
			expected:    "lsif-go",
		},
		{
			explbnbtion: "different nbme",
			input:       "myownlsif-go",
			expected:    "myownlsif-go",
		},
	}

	for _, test := rbnge tests {
		t.Run(test.explbnbtion, func(t *testing.T) {
			bctubl := extrbctIndexerNbme(test.input)
			if bctubl != test.expected {
				t.Errorf("unexpected indexer nbme. wbnt=%q hbve=%q", test.expected, bctubl)
			}
		})
	}
}
