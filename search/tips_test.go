package search

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph/mock"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/svc"
)

func TestTips(t *testing.T) {
	tests := []struct {
		tokens     []sourcegraph.Token
		wantCancel bool
		wantTips   []sourcegraph.TokenError
		wantErr    error

		mockBuildsGetRepoBuildInfo func(ctx context.Context, op *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error)
	}{
		{
			tokens:   []sourcegraph.Token{},
			wantTips: nil,
			wantErr:  nil,
		},
		{
			tokens: []sourcegraph.Token{
				sourcegraph.RepoToken{URI: "r"},
				sourcegraph.RepoToken{URI: "r"},
			},
			wantCancel: true,
			wantTips: []sourcegraph.TokenError{
				{
					Index:   2,
					Token:   tp(sourcegraph.RepoToken{URI: "r"}),
					Message: `Searching more than 1 repository at a time`,
				},
			},
		},
		{
			tokens: []sourcegraph.Token{
				sourcegraph.RepoToken{URI: "r"},
				sourcegraph.RevToken{Rev: "v"},
				sourcegraph.RevToken{Rev: "v2"},
			},
			wantCancel: true,
			wantTips: []sourcegraph.TokenError{
				{
					Index:   3,
					Token:   tp(sourcegraph.RevToken{Rev: "v2"}),
					Message: `Searching more than 1 revision at a time`,
				},
			},
		},
		{
			tokens: []sourcegraph.Token{
				sourcegraph.UserToken{Login: "u"},
				sourcegraph.UserToken{Login: "u2"},
			},
			wantCancel: true,
			wantTips: []sourcegraph.TokenError{
				{
					Index:   2,
					Token:   tp(sourcegraph.UserToken{Login: "u2"}),
					Message: `Searching more than 1 user's or org's repositories`,
				},
			},
		},
		{
			tokens: []sourcegraph.Token{
				sourcegraph.RevToken{Rev: "v"},
			},
			wantCancel: true,
			wantTips: []sourcegraph.TokenError{
				{
					Index:   1,
					Token:   tp(sourcegraph.RevToken{Rev: "v"}),
					Message: `You must specify a repository to search (e.g., 'github.com/user/repo')`,
				},
			},
		},
		{
			tokens: []sourcegraph.Token{
				sourcegraph.RevToken{Rev: "v"},
				sourcegraph.RepoToken{URI: "r"},
			},
			wantCancel: true,
			wantTips: []sourcegraph.TokenError{
				{
					Index:   1,
					Token:   tp(sourcegraph.RevToken{Rev: "v"}),
					Message: `You must specify a repository to search (e.g., 'github.com/user/repo')`,
				},
			},
		},
		{
			tokens: []sourcegraph.Token{
				sourcegraph.FileToken{Path: "p"},
			},
			wantCancel: true,
			wantTips: []sourcegraph.TokenError{
				{
					Index:   1,
					Token:   tp(sourcegraph.FileToken{Path: "p"}),
					Message: `You must specify a repository to search (e.g., 'github.com/user/repo')`,
				},
			},
		},
		{
			tokens: []sourcegraph.Token{
				sourcegraph.FileToken{Path: "p"},
				sourcegraph.RepoToken{URI: "r"},
			},
			wantCancel: true,
			wantTips: []sourcegraph.TokenError{
				{
					Index:   1,
					Token:   tp(sourcegraph.FileToken{Path: "p"}),
					Message: `You must specify a repository to search (e.g., 'github.com/user/repo')`,
				},
			},
		},
		{
			tokens: []sourcegraph.Token{
				sourcegraph.RepoToken{URI: "r", Repo: &sourcegraph.Repo{URI: "r", DefaultBranch: "b"}},
				sourcegraph.Term("t"),
			},
			wantTips: []sourcegraph.TokenError{
				{
					Index:   1,
					Token:   tp(sourcegraph.RepoToken{URI: "r", Repo: &sourcegraph.Repo{URI: "r", DefaultBranch: "b"}}),
					Message: `No build found for revision "b"`,
				},
			},
			mockBuildsGetRepoBuildInfo: reposGetBuildNone,
		},
		{
			tokens: []sourcegraph.Token{
				sourcegraph.RepoToken{URI: "r", Repo: &sourcegraph.Repo{URI: "r"}},
				sourcegraph.RevToken{Rev: "v", Commit: &vcs.Commit{ID: "c"}},
				sourcegraph.Term("t"),
			},
			wantTips: []sourcegraph.TokenError{
				{
					// TODO(sqs): probably best to associate this
					// error with the RevToken not RepoToken, but
					// let's keep it like this for now for simplicity.
					Index:   1,
					Token:   tp(sourcegraph.RepoToken{URI: "r", Repo: &sourcegraph.Repo{URI: "r"}}),
					Message: `No build found for revision "v"`,
				},
			},
			mockBuildsGetRepoBuildInfo: reposGetBuildNone,
		},
		{
			tokens: []sourcegraph.Token{
				sourcegraph.RepoToken{URI: "r", Repo: &sourcegraph.Repo{URI: "r"}},
				sourcegraph.RevToken{Rev: "v", Commit: &vcs.Commit{ID: "c"}},
				sourcegraph.Term("t"),
			},
			wantTips: []sourcegraph.TokenError{
				{
					// TODO(sqs): probably best to associate this
					// error with the RevToken not RepoToken, but
					// let's keep it like this for now for simplicity.
					Index:   1,
					Token:   tp(sourcegraph.RepoToken{URI: "r", Repo: &sourcegraph.Repo{URI: "r"}}),
					Message: `Latest build is 1 commits behind (c2).`,
				},
			},
			mockBuildsGetRepoBuildInfo: reposGetBuildOld,
		},
	}
	for _, test := range tests {
		label := "<< " + debugFormatTokens(test.tokens) + " >> "

		ctx := svc.WithServices(context.Background(), svc.Services{
			Builds: &mock.BuildsServer{
				GetRepoBuildInfo_: test.mockBuildsGetRepoBuildInfo,
			},
		})

		cancel, tips, err := Tips(ctx, test.tokens)
		if !reflect.DeepEqual(err, test.wantErr) {
			if test.wantErr == nil {
				t.Errorf("%s: Tips: %s", label, err)
			} else {
				t.Errorf("%s: Tips: got error\n%q\n\nwant\n%q", label, err, test.wantErr)
			}
			continue
		}
		if err != nil {
			continue
		}

		if cancel != test.wantCancel {
			t.Errorf("%s: got tips cancel %v, want %v", label, cancel, test.wantCancel)
		}

		// Allow specifying just a prefix of the tip message.
		for i, tip := range tips {
			if i < len(test.wantTips) {
				wantTip := &test.wantTips[i]
				if wantTip.Message != "" && strings.HasPrefix(tip.Message, wantTip.Message) && tip.Message != wantTip.Message {
					wantTip.Message += "... (prefix match)"
					tips[i].Message = wantTip.Message
				}
			}
		}

		if !reflect.DeepEqual(tips, test.wantTips) {
			t.Errorf("%s: got tips %v, want %v", label, tips, test.wantTips)
		}
	}
}
