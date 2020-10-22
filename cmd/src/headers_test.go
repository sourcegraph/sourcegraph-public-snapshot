package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseAdditionalHeaders(t *testing.T) {
	testCases := []struct {
		environ []string
		headers map[string]string
	}{
		{environ: []string{}, headers: map[string]string{}},
		{environ: []string{"AUTHORIZATION=foo,bar,baz"}, headers: map[string]string{}},
		{environ: []string{"SRC_HEADER_AUTHORIZATION=foo,bar,baz"}, headers: map[string]string{"authorization": "foo,bar,baz"}},
		{environ: []string{"SRC_HEADER_A=foo", "SRC_HEADER_B=bar", "SRC_HEADER_C=baz"}, headers: map[string]string{"a": "foo", "b": "bar", "c": "baz"}},
		{environ: []string{"SRC_HEADER_A", "SRC_HEADER_B=", "SRC_HEADER_=baz"}, headers: map[string]string{}},
		{environ: []string{"SRC_HEADER_X-Dbx-Auth-Token=foo"}, headers: map[string]string{"x-dbx-auth-token": "foo"}},
	}

	for _, testCase := range testCases {
		t.Run(strings.Join(testCase.environ, " "), func(t *testing.T) {
			if diff := cmp.Diff(testCase.headers, parseAdditionalHeadersFromMap(testCase.environ)); diff != "" {
				t.Errorf("unexpected headers: %s", diff)
			}
		})
	}
}
