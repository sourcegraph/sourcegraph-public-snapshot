package types

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/go-diff/diff"

	adobatches "github.com/sourcegraph/sourcegraph/internal/batches/sources/azuredevops"
	bbcs "github.com/sourcegraph/sourcegraph/internal/batches/sources/bitbucketcloud"
	gerritbatches "github.com/sourcegraph/sourcegraph/internal/batches/sources/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func TestChangeset_Clone(t *testing.T) {
	original := &Changeset{
		ID: 1,
		BatchChanges: []BatchChangeAssoc{
			{BatchChangeID: 999, IsArchived: true, Detach: true, Archive: true},
		},
	}

	clone := original.Clone()
	clone.BatchChanges[0].IsArchived = false

	if !original.BatchChanges[0].IsArchived {
		t.Fatalf("BatchChanges association was not cloned but is still reference")
	}
}

func TestChangeset_DiffStat(t *testing.T) {
	var (
		added   int32 = 77
		deleted int32 = 99
	)

	for name, tc := range map[string]struct {
		c    Changeset
		want *diff.Stat
	}{
		"added missing": {
			c: Changeset{
				DiffStatAdded:   nil,
				DiffStatDeleted: &deleted,
			},
			want: nil,
		},
		"deleted missing": {
			c: Changeset{
				DiffStatAdded:   &added,
				DiffStatDeleted: nil,
			},
			want: nil,
		},
		"all present": {
			c: Changeset{
				DiffStatAdded:   &added,
				DiffStatDeleted: &deleted,
			},
			want: &diff.Stat{
				Added:   added,
				Deleted: deleted,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			have := tc.c.DiffStat()
			if (tc.want == nil && have != nil) || (tc.want != nil && have == nil) {
				t.Errorf("mismatched nils in diff stats: have %+v; want %+v", have, tc.want)
			} else if tc.want != nil && have != nil {
				if d := cmp.Diff(*have, *tc.want); d != "" {
					t.Errorf("incorrect diff stat: %s", d)
				}
			}
		})
	}
}

func TestChangeset_SetMetadata(t *testing.T) {
	for name, tc := range map[string]struct {
		meta any
		want *Changeset
	}{
		"bitbucketcloud with fork": {
			meta: &bbcs.AnnotatedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					ID: 12345,
					Source: bitbucketcloud.PullRequestEndpoint{
						Branch: bitbucketcloud.PullRequestBranch{Name: "branch"},
						Repo:   bitbucketcloud.Repo{FullName: "fork/repo", UUID: "fork"},
					},
					UpdatedOn: time.Unix(10, 0),
				},
				Statuses: []*bitbucketcloud.PullRequestStatus{},
			},
			want: &Changeset{
				ExternalID:            "12345",
				ExternalServiceType:   extsvc.TypeBitbucketCloud,
				ExternalBranch:        "refs/heads/branch",
				ExternalForkNamespace: "fork",
				ExternalUpdatedAt:     time.Unix(10, 0),
			},
		},
		"bitbucketcloud without fork": {
			meta: &bbcs.AnnotatedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					ID: 12345,
					Source: bitbucketcloud.PullRequestEndpoint{
						Branch: bitbucketcloud.PullRequestBranch{Name: "branch"},
						Repo:   bitbucketcloud.Repo{UUID: "repo"},
					},
					Destination: bitbucketcloud.PullRequestEndpoint{
						Repo: bitbucketcloud.Repo{UUID: "repo"},
					},
					UpdatedOn: time.Unix(10, 0),
				},
				Statuses: []*bitbucketcloud.PullRequestStatus{},
			},
			want: &Changeset{
				ExternalID:            "12345",
				ExternalServiceType:   extsvc.TypeBitbucketCloud,
				ExternalBranch:        "refs/heads/branch",
				ExternalForkNamespace: "",
				ExternalUpdatedAt:     time.Unix(10, 0),
			},
		},
		"bitbucketserver": {
			meta: &bitbucketserver.PullRequest{
				ID: 12345,
				FromRef: bitbucketserver.Ref{
					ID: "refs/heads/branch",
					Repository: bitbucketserver.RefRepository{
						ID: 23456,
						Project: bitbucketserver.ProjectKey{
							Key: "upstream",
						},
					},
				},
				UpdatedDate: 10 * 1000,
			},
			want: &Changeset{
				ExternalID:            "12345",
				ExternalServiceType:   extsvc.TypeBitbucketServer,
				ExternalBranch:        "refs/heads/branch",
				ExternalForkNamespace: "upstream",
				ExternalUpdatedAt:     time.Unix(10, 0),
			},
		},
		"GitHub": {
			meta: &github.PullRequest{
				Number:      12345,
				HeadRefName: "branch",
				UpdatedAt:   time.Unix(10, 0),
			},
			want: &Changeset{
				ExternalID:          "12345",
				ExternalServiceType: extsvc.TypeGitHub,
				ExternalBranch:      "refs/heads/branch",
				ExternalUpdatedAt:   time.Unix(10, 0),
			},
		},
		"GitLab": {
			meta: &gitlab.MergeRequest{
				IID:          12345,
				SourceBranch: "branch",
				UpdatedAt:    gitlab.Time{Time: time.Unix(10, 0)},
			},
			want: &Changeset{
				ExternalID:          "12345",
				ExternalServiceType: extsvc.TypeGitLab,
				ExternalBranch:      "refs/heads/branch",
				ExternalUpdatedAt:   time.Unix(10, 0),
			},
		},
		"Azure DevOps with fork": {
			meta: &adobatches.AnnotatedPullRequest{
				PullRequest: &azuredevops.PullRequest{
					ID:            12345,
					SourceRefName: "refs/heads/branch",
					ForkSource: &azuredevops.ForkRef{
						Repository: azuredevops.Repository{
							Name: "forked-repo",
							Project: azuredevops.Project{
								Name: "fork",
							},
						},
					},
					CreationDate: time.Unix(10, 0),
				},
				Statuses: []*azuredevops.PullRequestBuildStatus{},
			},
			want: &Changeset{
				ExternalID:            "12345",
				ExternalServiceType:   extsvc.TypeAzureDevOps,
				ExternalBranch:        "refs/heads/branch",
				ExternalForkNamespace: "fork",
				ExternalForkName:      "forked-repo",
				ExternalUpdatedAt:     time.Unix(10, 0),
			},
		},
		"Azure DevOps without fork": {
			meta: &adobatches.AnnotatedPullRequest{
				PullRequest: &azuredevops.PullRequest{
					ID:            12345,
					SourceRefName: "refs/heads/branch",
					CreationDate:  time.Unix(10, 0),
				},
				Statuses: []*azuredevops.PullRequestBuildStatus{},
			},
			want: &Changeset{
				ExternalID:            "12345",
				ExternalServiceType:   extsvc.TypeAzureDevOps,
				ExternalBranch:        "refs/heads/branch",
				ExternalForkNamespace: "",
				ExternalForkName:      "",
				ExternalUpdatedAt:     time.Unix(10, 0),
			},
		},
		"Gerrit": {
			meta: &gerritbatches.AnnotatedChange{
				Change: &gerrit.Change{
					ChangeID: "I5de272baea22ef34dfbd00d6e96c45b25019697f",
					Branch:   "branch",
					Updated:  time.Unix(10, 0),
				},
			},
			want: &Changeset{
				ExternalID:            "I5de272baea22ef34dfbd00d6e96c45b25019697f",
				ExternalServiceType:   extsvc.TypeGerrit,
				ExternalBranch:        "refs/heads/branch",
				ExternalForkNamespace: "",
				ExternalForkName:      "",
				ExternalUpdatedAt:     time.Unix(10, 0),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			have := &Changeset{}
			want := tc.want
			want.Metadata = tc.meta

			if err := have.SetMetadata(tc.meta); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if d := cmp.Diff(have, want); d != "" {
				t.Errorf("metadata not updated as expected: %s", d)
			}
		})
	}
}

func TestChangeset_Title(t *testing.T) {
	want := "foo"
	for name, meta := range map[string]any{
		"azuredevops": &adobatches.AnnotatedPullRequest{
			PullRequest: &azuredevops.PullRequest{Title: want},
		},
		"bitbucketcloud": &bbcs.AnnotatedPullRequest{
			PullRequest: &bitbucketcloud.PullRequest{Title: want},
		},
		"bitbucketserver": &bitbucketserver.PullRequest{
			Title: want,
		},
		"Gerrit": &gerritbatches.AnnotatedChange{
			Change: &gerrit.Change{
				Subject: want,
			},
		},
		"GitHub": &github.PullRequest{
			Title: want,
		},
		"GitLab": &gitlab.MergeRequest{
			Title: want,
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: meta}
			have, err := c.Title()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != want {
				t.Errorf("unexpected title: have %s; want %s", have, want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.Title(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_ExternalCreatedAt(t *testing.T) {
	want := time.Unix(10, 0)
	for name, meta := range map[string]any{
		"azuredevops": &adobatches.AnnotatedPullRequest{
			PullRequest: &azuredevops.PullRequest{CreationDate: want},
		},
		"bitbucketcloud": &bbcs.AnnotatedPullRequest{
			PullRequest: &bitbucketcloud.PullRequest{CreatedOn: want},
		},
		"bitbucketserver": &bitbucketserver.PullRequest{
			CreatedDate: 10 * 1000,
		},
		"Gerrit": &gerritbatches.AnnotatedChange{
			Change: &gerrit.Change{
				Created: want,
			},
		},
		"GitHub": &github.PullRequest{
			CreatedAt: want,
		},
		"GitLab": &gitlab.MergeRequest{
			CreatedAt: gitlab.Time{Time: want},
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: meta}
			if have := c.ExternalCreatedAt(); have != want {
				t.Errorf("unexpected external creation date: have %+v; want %+v", have, want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		want := time.Time{}
		if have := c.ExternalCreatedAt(); have != want {
			t.Errorf("unexpected external creation date: have %+v; want %+v", have, want)
		}
	})
}

func TestChangeset_Body(t *testing.T) {
	want := "foo"
	for name, meta := range map[string]any{
		"azuredevops": &adobatches.AnnotatedPullRequest{
			PullRequest: &azuredevops.PullRequest{
				Description: want,
			},
		},
		"bitbucketcloud": &bbcs.AnnotatedPullRequest{
			PullRequest: &bitbucketcloud.PullRequest{
				Rendered: bitbucketcloud.RenderedPullRequestMarkup{
					Description: bitbucketcloud.RenderedMarkup{Raw: want},
				},
			},
		},
		"bitbucketserver": &bitbucketserver.PullRequest{
			Description: want,
		},
		"Gerrit": &gerritbatches.AnnotatedChange{
			Change: &gerrit.Change{
				Subject: want,
			},
		},
		"GitHub": &github.PullRequest{
			Body: want,
		},
		"GitLab": &gitlab.MergeRequest{
			Description: want,
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: meta}
			have, err := c.Body()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != want {
				t.Errorf("unexpected body: have %s; want %s", have, want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.Body(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_URL(t *testing.T) {
	want := "foo"
	for name, meta := range map[string]struct {
		pr   any
		want string
	}{
		"azuredevops": {
			pr: &adobatches.AnnotatedPullRequest{
				PullRequest: &azuredevops.PullRequest{
					ID: 12,
					Repository: azuredevops.Repository{
						Name: "repoName",
						Project: azuredevops.Project{
							Name: "projectName",
						},
						APIURL: "https://dev.azure.com/sgtestazure/projectName/_git/repositories/repoName",
					},
					URL: "https://dev.azure.com/sgtestazure/projectID/_apis/git/repositories/repoID/pullRequests/12",
				},
			},
			want: "https://dev.azure.com/sgtestazure/projectName/_git/repoName/pullrequest/12",
		},
		"bitbucketcloud": {
			pr: &bbcs.AnnotatedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					Links: bitbucketcloud.Links{
						"html": bitbucketcloud.Link{Href: want},
					},
				},
			},
			want: want,
		},
		"bitbucketserver": {
			pr: &bitbucketserver.PullRequest{
				Links: struct {
					Self []struct {
						Href string `json:"href"`
					} `json:"self"`
				}{
					Self: []struct {
						Href string `json:"href"`
					}{{Href: want}},
				},
			},
			want: want,
		},
		"Gerrit": {
			pr: &gerritbatches.AnnotatedChange{
				Change: &gerrit.Change{
					ChangeNumber: 1,
					Project:      "foo",
				},
				CodeHostURL: url.URL{Scheme: "https", Host: "gerrit.sgdev.org"},
			},
			want: "https://gerrit.sgdev.org/c/foo/+/1",
		},
		"GitHub": {
			pr: &github.PullRequest{
				URL: want,
			},
			want: want,
		},
		"GitLab": {
			pr: &gitlab.MergeRequest{
				WebURL: want,
			},
			want: want,
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: meta.pr}
			have, err := c.URL()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != meta.want {
				t.Errorf("unexpected URL: have %s; want %s", have, meta.want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.URL(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_HeadRefOid(t *testing.T) {
	for name, tc := range map[string]struct {
		meta any
		want string
	}{
		"azuredevops": {
			meta: &adobatches.AnnotatedPullRequest{},
			want: "",
		},
		"bitbucketcloud": {
			meta: &bbcs.AnnotatedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					Source: bitbucketcloud.PullRequestEndpoint{
						Commit: bitbucketcloud.PullRequestCommit{Hash: "foo"},
					},
				},
			},
			want: "foo",
		},
		"bitbucketserver": {
			meta: &bitbucketserver.PullRequest{},
			want: "",
		},
		"Gerrit": {
			meta: &gerritbatches.AnnotatedChange{},
			want: "",
		},
		"GitHub": {
			meta: &github.PullRequest{HeadRefOid: "foo"},
			want: "foo",
		},
		"GitLab": {
			meta: &gitlab.MergeRequest{
				DiffRefs: gitlab.DiffRefs{HeadSHA: "foo"},
			},
			want: "foo",
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: tc.meta}
			have, err := c.HeadRefOid()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != tc.want {
				t.Errorf("unexpected head ref OID: have %s; want %s", have, tc.want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.HeadRefOid(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_HeadRef(t *testing.T) {
	for name, tc := range map[string]struct {
		meta any
		want string
	}{
		"azuredevops": {
			meta: &adobatches.AnnotatedPullRequest{
				PullRequest: &azuredevops.PullRequest{
					SourceRefName: "refs/heads/foo",
				},
			},
			want: "refs/heads/foo",
		},
		"bitbucketcloud": {
			meta: &bbcs.AnnotatedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					Source: bitbucketcloud.PullRequestEndpoint{
						Branch: bitbucketcloud.PullRequestBranch{Name: "foo"},
					},
				},
			},
			want: "refs/heads/foo",
		},
		"bitbucketserver": {
			meta: &bitbucketserver.PullRequest{
				FromRef: bitbucketserver.Ref{ID: "foo"},
			},
			want: "foo",
		},
		"Gerrit": {
			// Gerrit does not return the head ref
			meta: &gerritbatches.AnnotatedChange{},
			want: "",
		},
		"GitHub": {
			meta: &github.PullRequest{HeadRefName: "foo"},
			want: "refs/heads/foo",
		},
		"GitLab": {
			meta: &gitlab.MergeRequest{
				SourceBranch: "foo",
			},
			want: "refs/heads/foo",
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: tc.meta}
			have, err := c.HeadRef()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != tc.want {
				t.Errorf("unexpected head ref: have %s; want %s", have, tc.want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.HeadRef(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_BaseRefOid(t *testing.T) {
	for name, tc := range map[string]struct {
		meta any
		want string
	}{
		"azuredevops": {
			meta: &adobatches.AnnotatedPullRequest{
				PullRequest: &azuredevops.PullRequest{},
			},
			want: "",
		},
		"bitbucketcloud": {
			meta: &bbcs.AnnotatedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					Destination: bitbucketcloud.PullRequestEndpoint{
						Commit: bitbucketcloud.PullRequestCommit{Hash: "foo"},
					},
				},
			},
			want: "foo",
		},
		"bitbucketserver": {
			meta: &bitbucketserver.PullRequest{},
			want: "",
		},
		"Gerrit": {
			meta: &gerritbatches.AnnotatedChange{},
			want: "",
		},
		"GitHub": {
			meta: &github.PullRequest{BaseRefOid: "foo"},
			want: "foo",
		},
		"GitLab": {
			meta: &gitlab.MergeRequest{
				DiffRefs: gitlab.DiffRefs{BaseSHA: "foo"},
			},
			want: "foo",
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: tc.meta}
			have, err := c.BaseRefOid()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != tc.want {
				t.Errorf("unexpected base ref OID: have %s; want %s", have, tc.want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.BaseRefOid(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_BaseRef(t *testing.T) {
	for name, tc := range map[string]struct {
		meta any
		want string
	}{
		"azuredevops": {
			meta: &adobatches.AnnotatedPullRequest{
				PullRequest: &azuredevops.PullRequest{TargetRefName: "refs/heads/foo"},
			},
			want: "refs/heads/foo",
		},
		"bitbucketcloud": {
			meta: &bbcs.AnnotatedPullRequest{
				PullRequest: &bitbucketcloud.PullRequest{
					Destination: bitbucketcloud.PullRequestEndpoint{
						Branch: bitbucketcloud.PullRequestBranch{Name: "foo"},
					},
				},
			},
			want: "refs/heads/foo",
		},
		"bitbucketserver": {
			meta: &bitbucketserver.PullRequest{
				ToRef: bitbucketserver.Ref{ID: "foo"},
			},
			want: "foo",
		},
		"Gerrit": {
			meta: &gerritbatches.AnnotatedChange{
				Change: &gerrit.Change{
					Branch: "foo",
				},
			},
			want: "refs/heads/foo",
		},
		"GitHub": {
			meta: &github.PullRequest{BaseRefName: "foo"},
			want: "refs/heads/foo",
		},
		"GitLab": {
			meta: &gitlab.MergeRequest{
				TargetBranch: "foo",
			},
			want: "refs/heads/foo",
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: tc.meta}
			have, err := c.BaseRef()
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if have != tc.want {
				t.Errorf("unexpected base ref: have %s; want %s", have, tc.want)
			}
		})
	}

	t.Run("unknown changeset type", func(t *testing.T) {
		c := &Changeset{}
		if _, err := c.BaseRef(); err == nil {
			t.Error("unexpected nil error")
		}
	})
}

func TestChangeset_Labels(t *testing.T) {
	for name, tc := range map[string]struct {
		meta any
		want []ChangesetLabel
	}{
		"azuredevops": {
			meta: &adobatches.AnnotatedPullRequest{},
			want: []ChangesetLabel{},
		},
		"bitbucketcloud": {
			meta: &bbcs.AnnotatedPullRequest{},
			want: []ChangesetLabel{},
		},
		"bitbucketserver": {
			meta: &bitbucketserver.PullRequest{},
			want: []ChangesetLabel{},
		},
		"Gerrit": {
			meta: &gerritbatches.AnnotatedChange{
				Change: &gerrit.Change{
					Hashtags: []string{"black", "green"},
				},
			},
			want: []ChangesetLabel{
				{Name: "black", Color: "000000"},
				{Name: "green", Color: "000000"},
			},
		},
		"GitHub": {
			meta: &github.PullRequest{
				Labels: struct{ Nodes []github.Label }{
					Nodes: []github.Label{
						{
							Name:        "red door",
							Color:       "black",
							Description: "paint it black",
						},
						{
							Name:        "grün",
							Color:       "green",
							Description: "groan",
						},
					},
				},
			},
			want: []ChangesetLabel{
				{
					Name:        "red door",
					Color:       "black",
					Description: "paint it black",
				},
				{
					Name:        "grün",
					Color:       "green",
					Description: "groan",
				},
			},
		},
		"GitLab": {
			meta: &gitlab.MergeRequest{
				Labels: []string{"black", "green"},
			},
			want: []ChangesetLabel{
				{Name: "black", Color: "000000"},
				{Name: "green", Color: "000000"},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := &Changeset{Metadata: tc.meta}
			if d := cmp.Diff(c.Labels(), tc.want); d != "" {
				t.Errorf("unexpected labels: %s", d)
			}
		})
	}
}

func TestChangesetMetadata(t *testing.T) {
	now := timeutil.Now()

	githubActor := github.Actor{
		AvatarURL: "https://avatars2.githubusercontent.com/u/1185253",
		Login:     "mrnugget",
		URL:       "https://github.com/mrnugget",
	}

	githubPR := &github.PullRequest{
		ID:           "FOOBARID",
		Title:        "Fix a bunch of bugs",
		Body:         "This fixes a bunch of bugs",
		URL:          "https://github.com/sourcegraph/sourcegraph/pull/12345",
		Number:       12345,
		State:        "MERGED",
		Author:       githubActor,
		Participants: []github.Actor{githubActor},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	changeset := &Changeset{
		RepoID:              42,
		CreatedAt:           now,
		UpdatedAt:           now,
		Metadata:            githubPR,
		BatchChanges:        []BatchChangeAssoc{},
		ExternalID:          "12345",
		ExternalServiceType: extsvc.TypeGitHub,
	}

	title, err := changeset.Title()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := githubPR.Title, title; want != have {
		t.Errorf("changeset title wrong. want=%q, have=%q", want, have)
	}

	body, err := changeset.Body()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := githubPR.Body, body; want != have {
		t.Errorf("changeset body wrong. want=%q, have=%q", want, have)
	}

	url, err := changeset.URL()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := githubPR.URL, url; want != have {
		t.Errorf("changeset url wrong. want=%q, have=%q", want, have)
	}
}

func TestChangeset_ResetReconcilerState(t *testing.T) {
	for name, tc := range map[string]struct {
		changeset *Changeset
		state     ReconcilerState
	}{
		"created changeset; has rollout windows": {
			changeset: &Changeset{CurrentSpecID: 1},
			state:     ReconcilerStateScheduled,
		},
		"created changeset; no rollout windows": {
			changeset: &Changeset{CurrentSpecID: 1},
			state:     ReconcilerStateQueued,
		},
		"tracking changeset; has rollout windows": {
			changeset: &Changeset{CurrentSpecID: 0},
			state:     ReconcilerStateQueued,
		},
		"tracking changeset; no rollout windows": {
			changeset: &Changeset{CurrentSpecID: 0},
			state:     ReconcilerStateQueued,
		},
	} {
		t.Run(name, func(t *testing.T) {
			// Set up a funky changeset state so we verify that the fields that
			// should be overwritten are.
			msg := "an appropriate error"
			tc.changeset.NumResets = 42
			tc.changeset.NumFailures = 43
			tc.changeset.FailureMessage = &msg
			tc.changeset.SyncErrorMessage = &msg

			tc.changeset.ResetReconcilerState(tc.state)
			if have := tc.changeset.ReconcilerState; have != tc.state {
				t.Errorf("unexpected reconciler state: have=%v want=%v", have, tc.state)
			}
			if have := tc.changeset.NumResets; have != 0 {
				t.Errorf("unexpected number of resets: have=%d want=0", have)
			}
			if have := tc.changeset.NumFailures; have != 0 {
				t.Errorf("unexpected number of failures: have=%d want=0", have)
			}
			if have := tc.changeset.FailureMessage; have != nil {
				t.Errorf("unexpected non-nil failure message: %s", *have)
			}
			if have := tc.changeset.SyncErrorMessage; have != nil {
				t.Errorf("unexpected non-nil sync error message: %s", *have)
			}
		})
	}
}
