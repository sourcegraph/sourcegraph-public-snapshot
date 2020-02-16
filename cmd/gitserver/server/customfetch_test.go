package server

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestEmptyCustomGitFetch(t *testing.T) {
	customGitFetch = func() interface{} {
		return buildCustomFetchMappings(nil)
	}

	customCmd := customFetchCmd(context.Background(), "git@github.com:sourcegraph/sourcegraph.git")
	if customCmd != nil {
		t.Errorf("expected nil custom cmd for empty configuration, got %+v", customCmd)
	}
}

func TestCustomGitFetch(t *testing.T) {
	mappings := []*schema.CustomGitFetchMapping{
		{
			DomainPath: "github.com/foo/normal/one",
			Fetch:      "echo normal one",
		},
		{
			DomainPath: "github.com/foo/normal/two",
			Fetch:      "echo normal two",
		},
		{
			DomainPath: "git@github.com:foo/faulty",
			Fetch:      "",
		},
	}

	tests := []struct {
		Url            string
		ExpectedCustom bool
		ExpectedArgs   []string
	}{
		{
			Url:            "https://8cd1419f4d5c1e0527f2893c9422f1a2a435116d@github.com/foo/normal/one",
			ExpectedCustom: true,
			ExpectedArgs:   []string{"echo", "normal", "one"},
		},
		{
			Url:            "https://8cd1419f4d5c1e0527f2893c9422f1a2a435116d@github.com/foo/normal/two",
			ExpectedCustom: true,
			ExpectedArgs:   []string{"echo", "normal", "two"},
		},
		{
			Url:            "https://8cd1419f4d5c1e0527f2893c9422f1a2a435116d@github.com/foo/faulty",
			ExpectedCustom: false,
		},
		{
			Url:            "https://8cd1419f4d5c1e0527f2893c9422f1a2a435116dgit@github.com/bar/notthere",
			ExpectedCustom: false,
		},
	}

	customGitFetch = func() interface{} {
		return buildCustomFetchMappings(mappings)
	}

	for _, test := range tests {
		customCmd := customFetchCmd(context.Background(), test.Url)

		if test.ExpectedCustom {
			if customCmd == nil {
				t.Errorf("expected custom command for url %s", test.Url)
			} else {
				if !reflect.DeepEqual(customCmd.Args, test.ExpectedArgs) {
					t.Errorf("expected custom command args %v for url %s, got %v", test.ExpectedArgs, test.Url,
						customCmd.Args)
				}
			}
		} else {
			if customCmd != nil {
				t.Errorf("expected no custom command for url %s, got %s", test.Url, customCmd.Path)
			}
		}
	}
}
