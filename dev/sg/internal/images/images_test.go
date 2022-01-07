package images

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
)

func TestParseTag(t *testing.T) {

	tests := []struct {
		name    string
		tag     string
		want    *SgImageTag
		wantErr bool
	}{
		// TODO: Add test cases.
		{"base",
			"12345_2021-01-02_abcdefg",
			&SgImageTag{
				buildNum:  12345,
				date:      "2021-01-02",
				shortSHA1: "abcdefg",
			},
			false,
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
	stdout.Out.SetVerbose()

	tests := []struct {
		name string
		tags []string
		want string
	}{
		// TODO: Add test cases.
		{
			"base",
			[]string{"v3.25.2", "12345_2022-01-01_asbcefg"},
			"12345_2022-01-01_asbcefg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findLatestTag(tt.tags); got != tt.want {
				t.Errorf("findLatestTag() = %v, want %v", got, tt.want)
			}
		})
	}
}
