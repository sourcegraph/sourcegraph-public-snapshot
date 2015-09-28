package sourcegraph

import "testing"

const commitID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestRepo_GitHubHTMLURL(t *testing.T) {
	tests := []struct {
		repo *Repo
		want string
	}{
		{
			repo: &Repo{URI: "github.com/o/r"},
			want: "https://github.com/o/r",
		},
		{
			repo: &Repo{URI: "foo.com/x/y"},
			want: "",
		},
	}
	for _, test := range tests {
		htmlURL := test.repo.GitHubHTMLURL()
		if htmlURL != test.want {
			t.Errorf("got %q, want %q", htmlURL, test.want)
		}
	}
}
