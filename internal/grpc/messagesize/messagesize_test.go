package messagesize

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetMessageSizeBytesFromEnv(t *testing.T) {
	tests := []struct {
		name string

		envVar   string
		envValue string

		checkEnvVar string

		expectedSize     int
		expectedErrRegex *regexp.Regexp
	}{
		{
			name: "8 MB",

			envVar:   "TEST_SIZE",
			envValue: "8MB",

			checkEnvVar: "TEST_SIZE",

			expectedSize: 8 * 1000 * 1000,
		},
		{
			name: "invalid size",

			envVar:   "TEST_SIZE",
			envValue: "this-is-not-a-size",

			checkEnvVar: "TEST_SIZE",

			expectedErrRegex: regexp.MustCompile("parsing.* as bytes"),
		},
		{
			name: "too small",

			envVar:   "TEST_SIZE",
			envValue: "2B", // Outside allowed range

			checkEnvVar: "TEST_SIZE",

			expectedErrRegex: regexp.MustCompile("outside of allowed range"),
		},
		{
			name: "unset",

			envVar: "", // Env var not set

			checkEnvVar: "TEST_SIZE",

			expectedErrRegex: regexp.MustCompile("environment variable .* not set"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.envVar != "" {
				t.Setenv(test.envVar, test.envValue)
			}

			size, err := getMessageSizeBytesFromEnv(test.envVar)
			if test.expectedErrRegex != nil {
				if err == nil {
					t.Fatalf("expected error %q, got no error", test.expectedErrRegex)
				}

				if !test.expectedErrRegex.MatchString(err.Error()) {
					t.Fatalf("expected error matching regex %q, got error %q", test.expectedErrRegex, err)
				}
			}

			if err != nil && test.expectedErrRegex == nil {
				t.Fatalf("expected no error, got error %q", err)
			}

			if diff := cmp.Diff(test.expectedSize, size); diff != "" {
				t.Fatalf("unexpected size (-want +got):\n%s", diff)
			}
		})

	}
}
