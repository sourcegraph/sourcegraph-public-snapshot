package feature

import "testing"

func TestSetFeatures(t *testing.T) {
	type mock struct {
		FeatA bool
		FeatB bool
	}

	cases := []struct {
		environ  []string
		expected mock
		hasError bool
	}{
		// No environment should not set anything in the mock
		{
			[]string{},
			mock{},
			false,
		},

		// Having an unexpected feature should not error out
		{
			[]string{"SG_FEATURE_UNEXPECTED=t"},
			mock{},
			false,
		},

		// Having an unexpected feature value should error out
		{
			[]string{"SG_FEATURE_FEATB=xyz"},
			mock{},
			true,
		},

		// Ignores other environment variables and sets a feature
		{
			[]string{"HOME=/root", "SG_FEATURE_FEATB=t"},
			mock{FeatB: true},
			false,
		},
	}

	for _, c := range cases {
		in := mock{}
		err := setFeatures(&in, c.environ)
		if (err != nil) != (c.hasError) {
			t.Errorf("Expected error for %#v", c)
		} else if in != c.expected {
			t.Errorf("%#v != %#v for %v", in, c.expected, c.environ)
		}
	}
}
