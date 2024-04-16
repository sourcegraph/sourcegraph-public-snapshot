package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadConfig(t *testing.T) {
	tests := []struct {
		name         string
		fileContents *config
		envToken     string
		envFooHeader string
		envHeaders   string
		envEndpoint  string
		flagEndpoint string
		want         *config
		wantErr      string
	}{
		{
			name: "defaults",
			want: &config{
				Endpoint:          "https://sourcegraph.com",
				AdditionalHeaders: map[string]string{},
			},
		},
		{
			name: "config file, no overrides, trim slash",
			fileContents: &config{
				Endpoint:    "https://example.com/",
				AccessToken: "deadbeef",
			},
			want: &config{
				Endpoint:          "https://example.com",
				AccessToken:       "deadbeef",
				AdditionalHeaders: map[string]string{},
			},
		},
		{
			name: "config file, token override only",
			fileContents: &config{
				Endpoint:    "https://example.com/",
				AccessToken: "deadbeef",
			},
			envToken: "abc",
			want:     nil,
			wantErr:  errConfigMerge.Error(),
		},
		{
			name: "config file, endpoint override only",
			fileContents: &config{
				Endpoint:    "https://example.com/",
				AccessToken: "deadbeef",
			},
			envEndpoint: "https://exmaple2.com",
			want:        nil,
			wantErr:     errConfigMerge.Error(),
		},
		{
			name: "config file, both override",
			fileContents: &config{
				Endpoint:    "https://example.com/",
				AccessToken: "deadbeef",
			},
			envToken:    "abc",
			envEndpoint: "https://override.com",
			want: &config{
				Endpoint:          "https://override.com",
				AccessToken:       "abc",
				AdditionalHeaders: map[string]string{},
			},
		},
		{
			name:     "no config file, token from environment",
			envToken: "abc",
			want: &config{
				Endpoint:          "https://sourcegraph.com",
				AccessToken:       "abc",
				AdditionalHeaders: map[string]string{},
			},
		},
		{
			name:        "no config file, endpoint from environment",
			envEndpoint: "https://example.com",
			want: &config{
				Endpoint:          "https://example.com",
				AccessToken:       "",
				AdditionalHeaders: map[string]string{},
			},
		},
		{
			name:        "no config file, both variables",
			envEndpoint: "https://example.com",
			envToken:    "abc",
			want: &config{
				Endpoint:          "https://example.com",
				AccessToken:       "abc",
				AdditionalHeaders: map[string]string{},
			},
		},
		{
			name:         "endpoint flag should override config",
			flagEndpoint: "https://override.com/",
			fileContents: &config{
				Endpoint:          "https://example.com/",
				AccessToken:       "deadbeef",
				AdditionalHeaders: map[string]string{},
			},
			want: &config{
				Endpoint:          "https://override.com",
				AccessToken:       "deadbeef",
				AdditionalHeaders: map[string]string{},
			},
		},
		{
			name:         "endpoint flag should override environment",
			flagEndpoint: "https://override.com/",
			envEndpoint:  "https://example.com",
			envToken:     "abc",
			want: &config{
				Endpoint:          "https://override.com",
				AccessToken:       "abc",
				AdditionalHeaders: map[string]string{},
			},
		},
		{
			name:         "additional header (with SRC_HEADER_ prefix)",
			flagEndpoint: "https://override.com/",
			envEndpoint:  "https://example.com",
			envToken:     "abc",
			envFooHeader: "bar",
			want: &config{
				Endpoint:          "https://override.com",
				AccessToken:       "abc",
				AdditionalHeaders: map[string]string{"foo": "bar"},
			},
		},
		{
			name:         "additional headers (with SRC_HEADERS key)",
			flagEndpoint: "https://override.com/",
			envEndpoint:  "https://example.com",
			envToken:     "abc",
			envHeaders:   "foo:bar\nfoo-bar:bar-baz",
			want: &config{
				Endpoint:          "https://override.com",
				AccessToken:       "abc",
				AdditionalHeaders: map[string]string{"foo-bar": "bar-baz", "foo": "bar"},
			},
		},
		{
			name:        "additional headers SRC_HEADERS_AUTHORIZATION and SRC_ACCESS_TOKEN",
			envToken:    "abc",
			envEndpoint: "https://override.com",
			envHeaders:  "Authorization:Bearer",
			wantErr:     errConfigAuthorizationConflict.Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setEnv := func(name, val string) {
				old := os.Getenv(name)
				if err := os.Setenv(name, val); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { os.Setenv(name, old) })
			}
			setEnv("SRC_ACCESS_TOKEN", test.envToken)
			setEnv("SRC_ENDPOINT", test.envEndpoint)

			tmpDir := t.TempDir()
			testHomeDir = tmpDir

			if test.flagEndpoint != "" {
				val := test.flagEndpoint
				endpoint = &val
				t.Cleanup(func() { endpoint = nil })
			}

			if test.fileContents != nil {
				oldConfigPath := *configPath
				t.Cleanup(func() { *configPath = oldConfigPath })

				data, err := json.Marshal(*test.fileContents)
				if err != nil {
					t.Fatal(err)
				}
				filePath := filepath.Join(tmpDir, "config.json")
				err = os.WriteFile(filePath, data, 0600)
				if err != nil {
					t.Fatal(err)
				}
				*configPath = filePath
			}

			if err := os.Setenv("SRC_HEADER_FOO", test.envFooHeader); err != nil {
				t.Fatal(err)
			}

			if err := os.Setenv("SRC_HEADERS", test.envHeaders); err != nil {
				t.Fatal(err)
			}

			config, err := readConfig()
			if diff := cmp.Diff(test.want, config); diff != "" {
				t.Errorf("config: %v", diff)
			}
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if diff := cmp.Diff(test.wantErr, errMsg); diff != "" {
				t.Errorf("err: %v", diff)
			}
		})
	}
}
