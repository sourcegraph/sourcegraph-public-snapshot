package main

import (
	"fmt"
	"testing"
)

func Test_trimImageTarballTarget(t *testing.T) {
	tests := []struct {
		target string
		want   string
	}{
		{
			target: "//cmd/worker:image_tarball",
			want:   "cmd/worker",
		},
		{
			target: "//docker-images/caddy:image_tarball",
			want:   "docker-images/caddy",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q to %q", tt.target, tt.want), func(t *testing.T) {
			got := trimImageTarballTarget(tt.target)
			if got != tt.want {
				t.Logf("got %q but wanted %q", got, tt.want)
				t.Fail()
			}
		})
	}
}
