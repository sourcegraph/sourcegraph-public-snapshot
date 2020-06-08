package github

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// NOTE: To update VCR for this test, please use the token of "sourcegraph-vcr"
// for GITHUB_TOKEN, which can be found in 1Password.
func TestClient_ListRepositoryCollaborators(t *testing.T) {
	tests := []struct {
		name      string
		owner     string
		repo      string
		wantUsers []*Collaborator
	}{
		{
			name:  "public repo",
			owner: "sourcegraph-vcr-repos",
			repo:  "public-org-repo-1",
			wantUsers: []*Collaborator{
				{
					ID:         "MDQ6VXNlcjYzMjkwODUx", // sourcegraph-vcr as owner
					DatabaseID: 63290851,
				}, {
					ID:         "MDQ6VXNlcjY2NDY0Nzcz", // sourcegraph-vcr-amy as organization member
					DatabaseID: 66464773,
				},
			},
		},
		{
			name:  "private repo",
			owner: "sourcegraph-vcr-repos",
			repo:  "private-org-repo-1",
			wantUsers: []*Collaborator{
				{
					ID:         "MDQ6VXNlcjM5MDk4MDA4", // sourcegraph-dogfood-user as outside collaborator
					DatabaseID: 39098008,
				}, {
					ID:         "MDQ6VXNlcjYzMjkwODUx", // sourcegraph-vcr as owner
					DatabaseID: 63290851,
				}, {
					ID:         "MDQ6VXNlcjY2NDY0Nzcz", // sourcegraph-vcr-amy as organization member
					DatabaseID: 66464773,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, save := newClient(t, "ListRepositoryCollaborators_"+test.name)
			defer save()

			users, _, err := client.ListRepositoryCollaborators(context.Background(), test.owner, test.repo, 1)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.wantUsers, users); diff != "" {
				t.Fatalf("Users mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
