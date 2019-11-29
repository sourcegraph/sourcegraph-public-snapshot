package a8n

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestNewCampaignType_ArgsValidation(t *testing.T) {
	tests := []struct {
		name               string
		campaignType, args string

		wantArgs interface{}
		err      string
	}{
		{
			name:         "valid comby",
			campaignType: "comby",
			args:         `{"scopeQuery":"repo:github","matchTemplate":"foobar","rewriteTemplate":"barfoo"}`,
			wantArgs: combyArgs{
				ScopeQuery:      "repo:github",
				MatchTemplate:   "foobar",
				RewriteTemplate: "barfoo",
			},
		},
		{
			name:         "invalid comby",
			campaignType: "comby",
			args:         `{"scopeQuery":""}`,
			err:          "3 errors occurred:\n\t* matchTemplate is required\n\t* rewriteTemplate is required\n\t* scopeQuery: String length must be greater than or equal to 1\n\n",
		},
		{
			name:         "valid credentials",
			campaignType: "credentials",
			args:         `{"scopeQuery":"repo:github","matchers":[{"type":"npm"}]}`,
			wantArgs: credentialsArgs{
				ScopeQuery: "repo:github",
				Matchers:   []credentialsMatcher{{MatcherType: "npm"}},
			},
		},
		{
			name:         "invalid credentials",
			campaignType: "credentials",
			args:         `{"scopeQuery":"","matchers":[]}`,
			err:          "2 errors occurred:\n\t* matchers: Array must have at least 1 items\n\t* scopeQuery: String length must be greater than or equal to 1\n\n",
		},
	}

	for _, tc := range tests {
		ct, err := NewCampaignType(tc.campaignType, tc.args, nil)

		var have string
		if err != nil {
			have = err.Error()
		}
		if have != tc.err {
			t.Errorf("got error %q, want %q", have, tc.err)
		}

		if tc.err != "" {
			continue
		}

		switch ct := ct.(type) {
		case *comby:
			wantArgs, _ := tc.wantArgs.(combyArgs)
			if !reflect.DeepEqual(ct.args, wantArgs) {
				t.Errorf("wrong args:\n%s", cmp.Diff(ct.args, wantArgs))
			}
		case *credentials:
			wantArgs, _ := tc.wantArgs.(credentialsArgs)
			if !reflect.DeepEqual(ct.args, wantArgs) {
				t.Errorf("wrong args:\n%s", cmp.Diff(ct.args, wantArgs))
			}
		default:
			t.Fatal("unknown campaign type")
		}
	}
}

func TestCampaignType_Comby(t *testing.T) {
	ctx := context.Background()

	combyJsonLineDiffs := []string{
		`{"uri":"file1.txt","diff":"--- file1.txt\n+++ file1.txt\n@@ -1,3 +1,3 @@\n file1-line1\n-file1-line2\n+file1-lineFOO\n file1-line3"}`,
		`{"uri":"file2.txt","diff":"--- file2.txt\n+++ file2.txt\n@@ -1,3 +1,3 @@\n file2-line1\n-file2-line2\n+file2-lineFOO\n file2-line3"}`,
	}
	diffs := []string{
		`diff file1.txt file1.txt
--- file1.txt
+++ file1.txt
@@ -1,3 +1,3 @@
 file1-line1
-file1-line2
+file1-lineFOO
 file1-line3`,
		`diff file2.txt file2.txt
--- file2.txt
+++ file2.txt
@@ -1,3 +1,3 @@
 file2-line1
-file2-line2
+file2-lineFOO
 file2-line3`,
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
			wantDiff: diffs[0],
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
			wantDiff: diffs[0] + "\n" + diffs[1],
		},
		{
			name:     "success multiple file diffs unsorted",
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

				fmt.Fprintln(w, combyJsonLineDiffs[1])
				fmt.Fprintln(w, combyJsonLineDiffs[0])
			},
			wantDiff: diffs[0] + "\n" + diffs[1],
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
