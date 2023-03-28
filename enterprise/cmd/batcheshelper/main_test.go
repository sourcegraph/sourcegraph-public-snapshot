package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestParseArguments(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedArgs args
		expectedErr  error
	}{
		{
			name: "Pre arguments are valid",
			args: []string{"pre", "1"},
			expectedArgs: args{
				mode: "pre",
				step: 1,
			},
		},
		{
			name: "Post arguments are valid",
			args: []string{"post", "1"},
			expectedArgs: args{
				mode: "post",
				step: 1,
			},
		},
		{
			name:        "Unknown mode",
			args:        []string{"foo", "1"},
			expectedErr: errors.New("invalid mode \"foo\""),
		},
		{
			name:        "No arguments",
			expectedErr: errors.New("missing arguments"),
		},
		{
			name:        "Too many arguments",
			args:        []string{"pre", "1", "foo"},
			expectedErr: errors.New("too many arguments"),
		},
		{
			name:        "Invalid step",
			args:        []string{"pre", "foo"},
			expectedErr: errors.New("failed to parse step: strconv.Atoi: parsing \"foo\": invalid syntax"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a, err := parseArgs(test.args)

			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedArgs, a)
			}
		})
	}
}
