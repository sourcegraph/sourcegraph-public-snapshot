package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadConfig(t *testing.T) {
	makeTempConfig := func(t *testing.T, c config) (string, func()) {
		data, err := json.Marshal(c)
		if err != nil {
			t.Fatal(err)
		}
		tmpDir, err := ioutil.TempDir("", "")
		if err != nil {
			t.Fatal(err)
		}
		filePath := filepath.Join(tmpDir, "config.json")
		err = ioutil.WriteFile(filePath, data, 0600)
		if err != nil {
			t.Fatal(err)
		}
		return filePath, func() { os.RemoveAll(tmpDir) }
	}

	tests := []struct {
		name         string
		fileContents *config
		envToken     string
		envEndpoint  string
		flagEndpoint string
		want         *config
		wantErr      string
	}{
		{
			name: "defaults",
			want: &config{
				Endpoint: "https://sourcegraph.com",
			},
		},
		{
			name: "config file, no overrides, trim slash",
			fileContents: &config{
				Endpoint:    "https://example.com/",
				AccessToken: "deadbeef",
			},
			want: &config{
				Endpoint:    "https://example.com",
				AccessToken: "deadbeef",
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
				Endpoint:    "https://override.com",
				AccessToken: "abc",
			},
		},
		{
			name:     "no config file, token from environment",
			envToken: "abc",
			want: &config{
				Endpoint:    "https://sourcegraph.com",
				AccessToken: "abc",
			},
		},
		{
			name:        "no config file, endpoint from environment",
			envEndpoint: "https://example.com",
			want: &config{
				Endpoint:    "https://example.com",
				AccessToken: "",
			},
		},
		{
			name:        "no config file, both variables",
			envEndpoint: "https://example.com",
			envToken:    "abc",
			want: &config{
				Endpoint:    "https://example.com",
				AccessToken: "abc",
			},
		},
		{
			name:         "endpoint flag should override config",
			flagEndpoint: "https://override.com/",
			fileContents: &config{
				Endpoint:    "https://example.com/",
				AccessToken: "deadbeef",
			},
			want: &config{
				Endpoint:    "https://override.com",
				AccessToken: "deadbeef",
			},
		},
		{
			name:         "endpoint flag should override environment",
			flagEndpoint: "https://override.com/",
			envEndpoint:  "https://example.com",
			envToken:     "abc",
			want: &config{
				Endpoint:    "https://override.com",
				AccessToken: "abc",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oldConfigPath := *configPath
			defer func() { *configPath = oldConfigPath }()

			if test.flagEndpoint != "" {
				val := test.flagEndpoint
				endpoint = &val
				defer func() { endpoint = nil }()
			}

			if test.fileContents != nil {
				p, cleanup := makeTempConfig(t, *test.fileContents)
				defer cleanup()
				*configPath = p
			}
			oldToken := os.Getenv("SRC_ACCESS_TOKEN")
			defer func() { os.Setenv("SRC_ACCESS_TOKEN", oldToken) }()
			oldEndpoint := os.Getenv("SRC_ENDPOINT")
			defer func() { os.Setenv("SRC_ENDPOINT", oldEndpoint) }()

			if err := os.Setenv("SRC_ACCESS_TOKEN", test.envToken); err != nil {
				t.Fatal(err)
			}
			if err := os.Setenv("SRC_ENDPOINT", test.envEndpoint); err != nil {
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
