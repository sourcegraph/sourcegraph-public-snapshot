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
	mux := http.NewServeMux()
	mux.HandleFunc("/.api/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(testResolveRepositorySearchResult))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	var clientBuffer bytes.Buffer
	client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

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
