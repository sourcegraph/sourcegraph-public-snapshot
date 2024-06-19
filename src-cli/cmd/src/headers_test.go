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
		{environ: []string{"SRC_HEADERS=foo:bar\nbar:baz\nAUTHORIZATION:Bearer somerandomstring"}, headers: map[string]string{"foo": "bar", "bar": "baz", "authorization": "Bearer somerandomstring"}},
		{environ: []string{"SRC_HEADERS=foo:bar\nbar:baz\nfoo-bar:baz-bar"}, headers: map[string]string{"foo": "bar", "bar": "baz", "foo-bar": "baz-bar"}},
		{environ: []string{"SRC_HEADERS=\"foo:bar\nbar:baz\nfoo-bar:baz-bar\""}, headers: map[string]string{"foo": "bar", "bar": "baz", "foo-bar": "baz-bar"}},
		{environ: []string{"SRC_HEADERS=foo:bar\nbar:baz\n foo-bar    :   baz-bar\nb: bar", "SRC_HEADER_A=foo"}, headers: map[string]string{"foo": "bar", "bar": "baz", "foo-bar": "baz-bar", "b": "bar", "a": "foo"}},
		{environ: []string{"SRC_HEADERS", "SRC_HEADER_A=foo"}, headers: map[string]string{"a": "foo"}},
	}

	for _, testCase := range testCases {
		t.Run(strings.Join(testCase.environ, " "), func(t *testing.T) {
			if diff := cmp.Diff(testCase.headers, parseAdditionalHeadersFromEnviron(testCase.environ)); diff != "" {
				t.Errorf("unexpected headers: %s", diff)
			}
		})
	}
}
