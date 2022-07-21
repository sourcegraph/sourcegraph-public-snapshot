package codeownership

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func Test_applyCodeOwnershipFiltering(t *testing.T) {
	type args struct {
		fileOwnersMustInclude []string
		fileOwnersMustExclude []string
		matches               []result.Match
	}
	tests := []struct {
		name        string
		args        args
		wantMatches []result.Match
		wantErr     string
	}{
		{
			name: "filters all matches if we have to include an owner",
			args: args{
				fileOwnersMustInclude: []string{"@sqs"},
				fileOwnersMustExclude: []string{},
				matches: []result.Match{
					&result.FileMatch{
						File: result.File{
							Path: "README.md",
						},
					},
				},
			},
			wantMatches: []result.Match{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			matches, err := applyCodeOwnershipFiltering(ctx, tt.args.fileOwnersMustInclude, tt.args.fileOwnersMustExclude, tt.args.matches)
			if diff := cmp.Diff(matches, tt.wantMatches, cmpopts.IgnoreUnexported(search.RepoStatusMap{})); diff != "" {
				t.Fatal(diff)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected err, got none")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected err %q, got %q", tt.wantErr, err.Error())
				}
			}
		})
	}
}
