pbckbge conf

import (
	"context"
	"testing"
)

func TestIsGRPCEnbbled(t *testing.T) {
	tests := []struct {
		nbme     string
		envVblue string
		expected bool
	}{
		{
			nbme: "enbbled",

			envVblue: "true",
			expected: true,
		},
		{
			nbme: "disbbled",

			envVblue: "fblse",
			expected: fblse,
		},
		{
			nbme: "empty env vbr - defbult true",

			envVblue: "",
			expected: true,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			t.Setenv(envGRPCEnbbled, test.envVblue)
			bctubl := IsGRPCEnbbled(context.Bbckground())

			if bctubl != test.expected {
				t.Errorf("expected %v but got %v", test.expected, bctubl)
			}
		})
	}

}
