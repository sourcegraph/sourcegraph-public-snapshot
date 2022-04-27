package api

import (
	"testing"
)

func TestUndeleteRepoName(t *testing.T) {
	tests := []struct {
		name string
		have RepoName
		want RepoName
	}{
		{
			name: "Blank",
			have: RepoName("github.com/owner/repo"),
			want: RepoName("github.com/owner/repo"),
		},
		{
			name: "Non deleted",
			have: RepoName("github.com/owner/repo"),
			want: RepoName("github.com/owner/repo"),
		},
		{
			name: "Deleted 1",
			have: RepoName("DELETED-1650360042.603863-github.com/owner/repo"),
			want: RepoName("github.com/owner/repo"),
		},
		{
			name: "Deleted 2",
			have: RepoName("DELETED-1650977466.716686-github.com/owner/repo"),
			want: RepoName("github.com/owner/repo"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UndeleteRepoName(tt.have); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
