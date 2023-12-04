package images

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/ci/images"
)

func mustTime() time.Time {
	t, err := time.Parse("2006-01-02", "2006-01-02")
	if err != nil {
		panic(err)
	}
	return t
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		want    *ParsedMainBranchImageTag
		wantErr bool
	}{
		{
			"base",
			"12345_2021-01-02_abcdefghijkl",
			&ParsedMainBranchImageTag{
				Build:       12345,
				Date:        "2021-01-02",
				ShortCommit: "abcdefghijkl",
			},
			false,
		},
		{
			"err",
			"3.25.5",
			nil,
			true,
		},
		{
			"from constructor",
			images.BranchImageTag(mustTime(), "abcde", 1234, "main", ""),
			&ParsedMainBranchImageTag{
				Build:       1234,
				Date:        "2006-01-02",
				ShortCommit: "abcde",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMainBranchImageTag(tt.tag)
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

func TestFindLatestTag(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{
			"base",
			[]string{"v3.25.2", "12345_2022-01-01_abcdefghijkl"},
			"12345_2022-01-01_abcdefghijkl",
		},
		{
			"higher_build_first",
			[]string{"99981_2022-01-15_999999a", "99982_2022-01-29_abcdefghijkl"},
			"99982_2022-01-29_abcdefghijkl",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := FindLatestMainTag(tt.tags); got != tt.want {
				t.Errorf("findLatestTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRawImgString(t *testing.T) {
	tests := []struct {
		name   string
		rawImg string
		want   *Repository
	}{
		{
			"base",
			"index.docker.io/sourcegraph/server:3.36.2@sha256:07d7407fdc656d7513aa54cdffeeecb33aa4e284eea2fd82e27342411430e5f2",
			&Repository{
				registry: "docker.io",
				org:      "sourcegraph",
				name:     "server",
				tag:      "3.36.2",
				digest:   "sha256:07d7407fdc656d7513aa54cdffeeecb33aa4e284eea2fd82e27342411430e5f2",
			},
		},
		{
			"base",
			"index.docker.io/sourcegraph/server:3.36.2",
			&Repository{
				registry: "docker.io",
				org:      "sourcegraph",
				name:     "server",
				tag:      "3.36.2",
				digest:   "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := ParseRepository(tt.rawImg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseImgString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
