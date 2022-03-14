package run

import "testing"

func TestRenderfield(t *testing.T) {
	fakeEnv := map[string]string{
		"FOOBAR_VERSION": "1.2.3.4",
	}
	lookup := func(key string) string { return fakeEnv[key] }

	for _, tt := range []struct {
		input string
		want  string
	}{
		{
			input: `https://example.com/releases/v{{ getEnv "FOOBAR_VERSION" }}/foobar_{{ getEnv "FOOBAR_VERSION" }}_amd64.tar.gz`,
			want:  `https://example.com/releases/v1.2.3.4/foobar_1.2.3.4_amd64.tar.gz`,
		},
	} {
		have, err := renderField("url", tt.input, lookup)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if have != tt.want {
			t.Fatalf("wrong output.\n\twant=%q\n\thave=%q", tt.want, have)
		}
	}

}
