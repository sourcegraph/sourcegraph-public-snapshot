package a8n

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestCampaignType_Comby(t *testing.T) {
	ctx := context.Background()

	combyJsonLineDiffs := []string{
		`{"uri":"file1.txt","diff":"--- file1.txt\n+++ file1.txt\n@@ -1,3 +1,3 @@\n file1-line1\n-file1-line2\n+file1-lineFOO\n file1-line3"}`,
		`{"uri":"file2.txt","diff":"--- file2.txt\n+++ file2.txt\n@@ -1,3 +1,3 @@\n file2-line1\n-file2-line2\n+file2-lineFOO\n file2-line3"}`,
	}

	tests := []struct {
		name string

		repoName string
		commitID string
		args     combyArgs

		handler func(w http.ResponseWriter, r *http.Request)

		wantDiff string
		wantErr  string
	}{
		{
			name:     "success single file diff",
			repoName: "github.com/sourcegraph/sourcegraph",
			commitID: "deadbeef",
			args: combyArgs{
				ScopeQuery:      "repo:gorilla",
				MatchTemplate:   "example.com",
				RewriteTemplate: "sourcegraph.com",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Transfer-Encoding", "chunked")
				w.WriteHeader(http.StatusOK)

				fmt.Fprintln(w, combyJsonLineDiffs[0])
			},
			wantDiff: `diff file1.txt file1.txt
--- file1.txt
+++ file1.txt
@@ -1,3 +1,3 @@
 file1-line1
-file1-line2
+file1-lineFOO
 file1-line3`,
		},
		{
			name:     "success multiple file diffs",
			repoName: "github.com/sourcegraph/sourcegraph",
			commitID: "deadbeef",
			args: combyArgs{
				ScopeQuery:      "repo:gorilla",
				MatchTemplate:   "example.com",
				RewriteTemplate: "sourcegraph.com",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Transfer-Encoding", "chunked")
				w.WriteHeader(http.StatusOK)

				fmt.Fprintln(w, combyJsonLineDiffs[0])
				fmt.Fprintln(w, combyJsonLineDiffs[1])
			},
			wantDiff: `diff file1.txt file1.txt
--- file1.txt
+++ file1.txt
@@ -1,3 +1,3 @@
 file1-line1
-file1-line2
+file1-lineFOO
 file1-line3
diff file2.txt file2.txt
--- file2.txt
+++ file2.txt
@@ -1,3 +1,3 @@
 file2-line1
-file2-line2
+file2-lineFOO
 file2-line3`,
		},
		{
			name:     "error",
			repoName: "github.com/sourcegraph/sourcegraph",
			commitID: "deadbeef",
			args: combyArgs{
				ScopeQuery:      "repo:gorilla",
				MatchTemplate:   "example.com",
				RewriteTemplate: "sourcegraph.com",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: `unexpected response status from replacer service: "500 Internal Server Error"`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				validateReplacerQuery(t, r.URL.Query(), tc.repoName, tc.commitID, tc.args)
				tc.handler(w, r)
			}))
			defer ts.Close()

			ct := &comby{
				replacerURL: ts.URL,
				httpClient:  &http.Client{},
				args:        tc.args,
			}

			if tc.wantErr == "" {
				tc.wantErr = "<nil>"
			}

			haveDiff, err := ct.generateDiff(ctx, api.RepoName(tc.repoName), api.CommitID(tc.commitID))
			if have, want := fmt.Sprint(err), tc.wantErr; have != want {
				t.Fatalf("have error: %q\nwant error: %q", have, want)
			}

			if haveDiff != tc.wantDiff {
				t.Fatalf("wrong diff.\nhave=%q\nwant=%q", haveDiff, tc.wantDiff)
			}
		})
	}
}

func validateReplacerQuery(t *testing.T, vals url.Values, repo, commit string, args combyArgs) {
	t.Helper()

	tests := []struct {
		name, want string
	}{
		{"repo", repo},
		{"commit", commit},
		{"matchtemplate", args.MatchTemplate},
		{"rewritetemplate", args.RewriteTemplate},
	}

	for _, tc := range tests {
		have, ok := vals[tc.name]
		if !ok || len(have) < 1 {
			t.Errorf("url param %q missing", tc.name)
			continue
		}
		if have[0] != tc.want {
			t.Errorf("wrong %q param: %s (want=%s)", tc.name, have[0], tc.want)
		}
	}
}
