package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
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

func TestParsePreviousStepResult(t *testing.T) {
	tests := []struct {
		name            string
		step            int
		skippedSteps    map[int]struct{}
		newStepFileFunc func(t *testing.T) string
		expected        execution.AfterStepResult
		expectedErr     error
	}{
		{
			name:     "No previous step",
			step:     0,
			expected: execution.AfterStepResult{},
		},
		{
			name:         "Previous step is skipped",
			step:         1,
			skippedSteps: map[int]struct{}{0: {}},
			expected:     execution.AfterStepResult{},
		},
		{
			name:         "All previous step is skipped",
			step:         3,
			skippedSteps: map[int]struct{}{0: {}, 1: {}, 2: {}},
			expected:     execution.AfterStepResult{},
		},
		{
			name: "Middle step skipped",
			step: 2,
			newStepFileFunc: func(t *testing.T) string {
				path := t.TempDir()
				err := os.WriteFile(filepath.Join(path, "step0.json"), []byte(`{"version": 2}`), os.ModePerm)
				require.NoError(t, err)
				return path
			},
			skippedSteps: map[int]struct{}{1: {}},
			expected:     execution.AfterStepResult{Version: 2},
		},
		{
			name: "Previous step is not skipped",
			step: 1,
			newStepFileFunc: func(t *testing.T) string {
				path := t.TempDir()
				err := os.WriteFile(filepath.Join(path, "step0.json"), []byte(`{"version": 2}`), os.ModePerm)
				require.NoError(t, err)
				return path
			},
			expected: execution.AfterStepResult{Version: 2},
		},
		{
			name: "Previous step is not skipped, but file is invalid",
			step: 1,
			newStepFileFunc: func(t *testing.T) string {
				path := t.TempDir()
				err := os.WriteFile(filepath.Join(path, "step0.json"), []byte(`{"version": 2`), os.ModePerm)
				require.NoError(t, err)
				return path
			},
			expectedErr: errors.New("failed to unmarshal step result file: unexpected end of JSON input"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var path string
			if test.newStepFileFunc != nil {
				path = test.newStepFileFunc(t)
			}
			result, err := parsePreviousStepResult(path, test.step)

			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}
