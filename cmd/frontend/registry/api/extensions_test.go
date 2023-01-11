package api

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSplitExtensionID(t *testing.T) {
	tests := map[string]struct {
		wantPrefix    string
		wantPublisher string
		wantName      string
		wantErr       bool
	}{
		"":        {wantErr: true},
		"/":       {wantErr: true},
		"a/":      {wantErr: true},
		"/a":      {wantErr: true},
		"//":      {wantErr: true},
		"b/c":     {wantPublisher: "b", wantName: "c"},
		"a/b/c":   {wantPrefix: "a", wantPublisher: "b", wantName: "c"},
		"a/b/c/d": {wantPrefix: "a/b", wantPublisher: "c", wantName: "d"},
	}
	for extensionID, test := range tests {
		t.Run(extensionID, func(t *testing.T) {
			prefix, publisher, name, err := SplitExtensionID(extensionID)
			if (err != nil) != test.wantErr {
				t.Errorf("got error %v, want error? %v", err, test.wantErr)
			}
			if err != nil {
				return
			}
			if prefix != test.wantPrefix {
				t.Errorf("got prefix %q, want %q", prefix, test.wantPrefix)
			}
			if publisher != test.wantPublisher {
				t.Errorf("got publisher %q, want %q", publisher, test.wantPublisher)
			}
			if name != test.wantName {
				t.Errorf("got name %q, want %q", name, test.wantName)
			}
		})
	}
}

func TestParseExtensionID(t *testing.T) {
	tests := map[string]struct {
		mockConfiguredPrefix         string
		wantPrefix                   string
		wantExtensionIDWithoutPrefix string
		wantIsLocal                  bool
		wantErr                      bool
	}{
		"":      {wantErr: true},
		"b/c":   {wantExtensionIDWithoutPrefix: "b/c", wantIsLocal: true},
		"a/b/c": {wantErr: true},
		"x/y/z": {mockConfiguredPrefix: "x", wantPrefix: "x", wantExtensionIDWithoutPrefix: "y/z", wantIsLocal: true},
		"y/z":   {mockConfiguredPrefix: "x", wantExtensionIDWithoutPrefix: "y/z", wantIsLocal: false},
		"w/y/z": {mockConfiguredPrefix: "x", wantErr: true},
	}
	for extensionID, test := range tests {
		t.Run(extensionID, func(t *testing.T) {
			var tmp *string
			if test.mockConfiguredPrefix != "" {
				tmp = &test.mockConfiguredPrefix
			}
			mockLocalRegistryExtensionIDPrefix = &tmp
			defer func() { mockLocalRegistryExtensionIDPrefix = nil }()

			prefix, extensionIDWithoutPrefix, isLocal, err := ParseExtensionID(extensionID)
			if (err != nil) != test.wantErr {
				t.Errorf("got error %v, want error? %v", err, test.wantErr)
			}
			if err != nil {
				return
			}
			if prefix != test.wantPrefix {
				t.Errorf("got prefix %q, want %q", prefix, test.wantPrefix)
			}
			if extensionIDWithoutPrefix != test.wantExtensionIDWithoutPrefix {
				t.Errorf("got extensionIDWithoutPrefix %q, want %q", extensionIDWithoutPrefix, test.wantExtensionIDWithoutPrefix)
			}
			if isLocal != test.wantIsLocal {
				t.Errorf("got isLocal %v, want %v", isLocal, test.wantIsLocal)
			}
		})
	}
}

func enableLegacyExtensions() {
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{
			EnableLegacyExtensions: true,
		},
	}})
}
