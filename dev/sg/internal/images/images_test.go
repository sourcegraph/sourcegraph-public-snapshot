package images

import (
	"reflect"
	"testing"
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
		{
			"b2",
			"123412",
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
