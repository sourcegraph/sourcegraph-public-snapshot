package campaigns

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/src-cli/internal/api"
)

func TestSetDefaultQueryCount(t *testing.T) {
	for in, want := range map[string]string{
		"":                     hardCodedCount,
		"count:10":             "count:10",
		"r:foo":                "r:foo" + hardCodedCount,
		"r:foo count:10":       "r:foo count:10",
		"r:foo count:10 f:bar": "r:foo count:10 f:bar",
		"r:foo count:":         "r:foo count:" + hardCodedCount,
		"r:foo count:xyz":      "r:foo count:xyz" + hardCodedCount,
	} {
		t.Run(in, func(t *testing.T) {
			have := setDefaultQueryCount(in)
			if have != want {
				t.Errorf("unexpected query: have %q; want %q", have, want)
			}
		})
	}
}

func TestResolveRepositorySearch(t *testing.T) {
	client, done := mockGraphQLClient(testResolveRepositorySearchResult)
	defer done()

	svc := &Service{client: client}

	repos, err := svc.resolveRepositorySearch(context.Background(), "sourcegraph reconciler or src-cli")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// We want multiple FileMatches in the same repository to result in one repository
	// and Repository and FileMatches should be folded into the same repository.
	if have, want := len(repos), 3; have != want {
		t.Fatalf("wrong number of repos. want=%d, have=%d", want, have)
	}
}

const testResolveRepositorySearchResult = `{
  "data": {
    "search": {
      "results": {
        "results": [
          {
            "__typename": "FileMatch",
            "file": { "path": "website/src/components/PricingTable.tsx" },
            "repository": {
              "id": "UmVwb3NpdG9yeTo0MDc=",
              "name": "github.com/sd9/about",
              "url": "/github.com/sd9/about",
              "externalRepository": { "serviceType": "github" },
              "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "1576f235209fbbfc918129db35f3d108347b74cb" } }
            }
          },
          {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeToxOTM=",
            "name": "github.com/sd9/src-cli",
            "url": "/github.com/sd9/src-cli",
            "externalRepository": { "serviceType": "github" },
            "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "21dd58b08d64620942401b5543f5b0d33498bacb" } }
          },
          {
            "__typename": "FileMatch",
            "file": { "path": "cmd/src/api.go" },
            "repository": {
              "id": "UmVwb3NpdG9yeToxOTM=",
              "name": "github.com/sd9/src-cli",
              "url": "/github.com/sd9/src-cli",
              "externalRepository": { "serviceType": "github" },
              "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "21dd58b08d64620942401b5543f5b0d33498bacb" } }
            }
          },
          {
            "__typename": "FileMatch",
            "file": { "path": "client/web/src/components/externalServices/externalServices.tsx" },
            "repository": {
              "id": "UmVwb3NpdG9yeTo2",
              "name": "github.com/sourcegraph/sourcegraph",
              "url": "/github.com/sourcegraph/sourcegraph",
              "externalRepository": { "serviceType": "github" },
              "defaultBranch": { "name": "refs/heads/main", "target": { "oid": "e612c34f2e27005928ff3dfdd8e5ead5a0cb1526" } }
            }
          },
          {
            "__typename": "FileMatch",
            "file": { "path": "cmd/frontend/internal/httpapi/src_cli.go" },
            "repository": {
              "id": "UmVwb3NpdG9yeTo2",
              "name": "github.com/sourcegraph/sourcegraph",
              "url": "/github.com/sourcegraph/sourcegraph",
              "externalRepository": { "serviceType": "github" },
              "defaultBranch": { "name": "refs/heads/main", "target": { "oid": "e612c34f2e27005928ff3dfdd8e5ead5a0cb1526" } }
            }
          }
        ]
      }
    }
  }
}
`

func mockGraphQLClient(response string) (client api.Client, done func()) {
	mux := http.NewServeMux()
	mux.HandleFunc("/.api/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	})

	ts := httptest.NewServer(mux)

	var clientBuffer bytes.Buffer
	client = api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

	return client, ts.Close
}

func TestResolveRepositories_Unsupported(t *testing.T) {
	client, done := mockGraphQLClient(testResolveRepositoriesUnsupported)
	defer done()

	spec := &CampaignSpec{
		On: []OnQueryOrRepository{
			{RepositoriesMatchingQuery: "testquery"},
		},
	}

	t.Run("allowUnsupported:true", func(t *testing.T) {
		svc := &Service{client: client, allowUnsupported: true}

		repos, err := svc.ResolveRepositories(context.Background(), spec)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(repos) != 4 {
			t.Fatalf("wrong number of repos. want=%d, have=%d", 4, len(repos))
		}
	})

	t.Run("allowUnsupported:false", func(t *testing.T) {
		svc := &Service{client: client, allowUnsupported: false}

		repos, err := svc.ResolveRepositories(context.Background(), spec)
		repoSet, ok := err.(UnsupportedRepoSet)
		if !ok {
			t.Fatalf("err is not UnsupportedRepoSet")
		}
		if len(repoSet) != 1 {
			t.Fatalf("wrong number of repos. want=%d, have=%d", 1, len(repoSet))
		}
		if len(repos) != 3 {
			t.Fatalf("wrong number of repos. want=%d, have=%d", 4, len(repos))
		}
	})
}

const testResolveRepositoriesUnsupported = `{
  "data": {
    "search": {
      "results": {
        "results": [
          {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeToxMw==",
            "name": "bitbucket.sgdev.org/SOUR/automation-testing",
            "url": "/bitbucket.sgdev.org/SOUR/automation-testing",
            "externalRepository": { "serviceType": "bitbucketserver" },
            "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "b978d56de5578a935ca0bf07b56528acc99d5a61" } }
          },
          {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeTo0",
            "name": "github.com/sourcegraph/automation-testing",
            "url": "/github.com/sourcegraph/automation-testing",
            "externalRepository": { "serviceType": "github" },
            "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "6ac8a32ecaf6c4dc5ce050b9af51bce3db8efd63" } }
          },
          {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeTo2MQ==",
            "name": "gitlab.sgdev.org/sourcegraph/automation-testing",
            "url": "/gitlab.sgdev.org/sourcegraph/automation-testing",
            "externalRepository": { "serviceType": "gitlab" },
            "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "3b79a5d479d2af9cfe91e0aad4e9dddca7278150" } }
          },
		  {
            "__typename": "Repository",
            "id": "UmVwb3NpdG9yeToxNDg=",
            "name": "git-codecommit.us-est-1.amazonaws.com/stripe-go",
            "url": "/git-codecommit.us-est-1.amazonaws.com/stripe-go",
            "externalRepository": { "serviceType": "awscodecommit" },
            "defaultBranch": { "name": "refs/heads/master", "target": { "oid": "3b79a5d479d2af9cfe91e0aad4e9dddca7278150" } }
          }
        ]
      }
    }
  }
}
`
