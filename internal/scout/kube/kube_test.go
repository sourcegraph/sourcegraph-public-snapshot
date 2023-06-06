package kube

import "testing"

func TestAcceptedFileSystem(t *testing.T) {
	cases := []struct {
		name       string
		filesystem string
		want       bool
	}{
		{
			name:       "should return true if filesystem matches 'matched' regular expression",
			filesystem: "/dev/sda",
			want:       true,
		},
		{
			name:       "should return false if filesystem doesn't match 'matched' regular expression",
			filesystem: "/dev/sda1",
			want:       false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := acceptedFileSystem(tc.filesystem)
			if got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}
