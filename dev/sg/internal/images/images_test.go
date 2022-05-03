package images

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestParseTag(t *testing.T) {
	if testing.Verbose() {
		stdout.Out.SetVerbose()
	}
	tests := []struct {
		name    string
		tag     string
		want    *SgImageTag
		wantErr bool
	}{
		{
			"base",
			"12345_2021-01-02_abcdefghijkl",
			&SgImageTag{
				buildNum:  12345,
				date:      "2021-01-02",
				shortSHA1: "abcdefghijkl",
			},
			false,
		},
		{
			"err",
			"3.25.5",
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTag(tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTag() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findLatestTag(t *testing.T) {
	if testing.Verbose() {
		stdout.Out.SetVerbose()
	}

	tests := []struct {
		name    string
		tags    []string
		want    string
		wantErr *error
	}{
		{
			"base",
			[]string{"v3.25.2", "12345_2022-01-01_abcdefghijkl"},
			"12345_2022-01-01_abcdefghijkl",
			nil,
		},
		{
			"higher_build_first",
			[]string{"99981_2022-01-15_999999a", "99982_2022-01-29_abcdefghijkl"},
			"99982_2022-01-29_abcdefghijkl",
			nil,
		},
		{
			"zoekt tag unsupported",
			[]string{"0.0.0-20200504095446-118acdf7aa8f", "0.0.0-20200505130024-763a9ca9b37c", "latest", "insiders"},
			"",
			&ErrUnsupportedTag,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findLatestTag(tt.tags)
			if err != nil {
				if tt.wantErr == nil {
					t.Errorf("got findLatestTag() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if !errors.Is(err, *tt.wantErr) {
					t.Errorf("findLatestTag() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				return
			}
			if got != tt.want {
				t.Errorf("findLatestTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRawImgString(t *testing.T) {
	if testing.Verbose() {
		stdout.Out.SetVerbose()
	}

	tests := []struct {
		name string
		tag  string
		want *ImageReference
	}{
		{
			"base",
			"index.docker.io/sourcegraph/server:3.36.2@sha256:07d7407fdc656d7513aa54cdffeeecb33aa4e284eea2fd82e27342411430e5f2",
			&ImageReference{
				Registry: "docker.io",
				Name:     "sourcegraph/server",
				Tag:      "3.36.2",
				Digest:   "sha256:07d7407fdc656d7513aa54cdffeeecb33aa4e284eea2fd82e27342411430e5f2",
			},
		},
		{
			"only tag",
			"index.docker.io/prom/prometheus:v2.35.0",
			&ImageReference{
				Registry: "docker.io",
				Name:     "prom/prometheus",
				Tag:      "v2.35.0",
			},
		},
		{
			"only digest",
			"index.docker.io/sourcegraph/server@sha256:07d7407fdc656d7513aa54cdffeeecb33aa4e284eea2fd82e27342411430e5f2",
			&ImageReference{
				Registry: "docker.io",
				Name:     "sourcegraph/server",
				Digest:   "sha256:07d7407fdc656d7513aa54cdffeeecb33aa4e284eea2fd82e27342411430e5f2",
				Tag:      "",
			},
		},
		{
			"non-sg image",
			"prom/blackbox-exporter:v0.17.0@sha256:1d8a5c9ff17e2493a39e4aea706b4ea0c8302ae0dc2aa8b0e9188c5919c9bd9c",
			&ImageReference{
				Registry: "docker.io",
				Name:     "prom/blackbox-exporter",
				Tag:      "v0.17.0",
				Digest:   "sha256:1d8a5c9ff17e2493a39e4aea706b4ea0c8302ae0dc2aa8b0e9188c5919c9bd9c",
			},
		},
		{
			"org and repo only",
			"sourcegraph/gitserver",
			&ImageReference{
				Registry: "docker.io",
				Name:     "sourcegraph/gitserver",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := parseImgString(tt.tag); !reflect.DeepEqual(got, tt.want) {
				if err != nil {
					t.Errorf("parseImgString() error = %v", err)
				} else {
					t.Errorf("parseImgString() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
