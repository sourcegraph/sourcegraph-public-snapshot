package shared

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		contents string
		err      string
	}{
		{
			contents: `{`,
			err:      "unmarshal JSON: failed to parse JSON: [CloseBraceExpected]",
		},

		{
			contents: `{}`,
			err:      "<nil>",
		},
	}
	for _, test := range tests {
		err := validateConfig(
			test.contents,
		)
		if !strings.Contains(fmt.Sprintf("%v", err), test.err) {
			t.Errorf("%s: got %q, want %q", test.contents, err, test.err)
		}
	}
}

func TestValidateExternalURL(t *testing.T) {
	tests := []struct {
		contents string
		err      string
	}{
		{
			contents: `{}`,
			err:      `"externalURL": value cannot be empty`,
		},
		{
			contents: `{"externalURL": "sourcegraph.example.com"}`,
			err:      `"externalURL": must start with http:// or https://"`,
		},
		{
			contents: `{"externalURL": "http://sourcegraph.example.com:abc"}`,
			err:      `invalid port`,
		},

		{
			contents: `{"externalURL": "https://sourcegraph.example.com/"}`,
			err:      "<nil>",
		},
		{
			contents: `{"externalURL": "http://localhost:3080"}`,
			err:      "<nil>",
		},
	}
	for _, test := range tests {
		err := validateConfig(
			test.contents,
			validateExternalURL,
		)
		if !strings.Contains(fmt.Sprintf("%v", err), test.err) {
			t.Errorf("%s: got %q, want %q", test.contents, err, test.err)
		}
	}
}
