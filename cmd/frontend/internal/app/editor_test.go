package app

import (
	"context"
	"strconv"
	"testing"
)

func TestEditor_remoteURLToRepoURI(t *testing.T) {
	tests := []struct {
		remoteURL, want, wantErr string
	}{
		{
			remoteURL: "ssh://git@github.com:org/repo",
			want:      "github.com/org/repo",
		},
		{
			remoteURL: "git@github.com:org/repo",
			want:      "github.com/org/repo",
		},
		{
			remoteURL: "git://git@github.com:org/repo",
			want:      "github.com/org/repo",
		},
		{
			remoteURL: "http://github.com/org/repo",
			want:      "github.com/org/repo",
		},
		{
			remoteURL: "https://github.com/org/repo",
			want:      "github.com/org/repo",
		},
		{
			remoteURL: "https://user:pass@github.com/org/repo",
			want:      "github.com/org/repo",
		},
		{
			remoteURL: "git@companydomain.com",
			wantErr:   "Git remote URL \"git@companydomain.com\" not supported.",
		},
		{
			remoteURL: "git@github.companydomain.com",
			wantErr:   "Git remote URL \"git@github.companydomain.com\" not supported.",
		},
	}
	for i, tst := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := remoteURLToRepoURI(context.Background(), tst.remoteURL)
			if err != nil {
				if err.Error() != tst.wantErr {
					t.Fatalf("got err %q want err %q; remoteURL %q\n", err.Error(), tst.wantErr, tst.remoteURL)
				}
			} else if got != tst.want {
				t.Fatalf("got %q want %q; remoteURL %q\n", got, tst.want, tst.remoteURL)
			}
		})
	}
}
