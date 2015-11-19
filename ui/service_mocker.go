package ui

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"sourcegraph.com/sqs/pbtypes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httptestutil"
)

// serviceMocker holds the Sourcegraph API client and enriches payloads with
// the possibility to generate mocks.
type serviceMocker struct {
	mockClient *sourcegraph.Client
}

// mockPayload is a structure against which the JSON body of the POST request
// the asked for mock data will be decoded. Wherever matched, those values will
// be used as mock data for any function that returns them.
type mockPayload struct {
	Repo          *sourcegraph.Repo          `json:",omitempty"`
	Commit        *vcs.Commit                `json:",omitempty"`
	RepoBuildInfo *sourcegraph.RepoBuildInfo `json:",omitempty"`
	RepoConfig    *sourcegraph.RepoConfig    `json:",omitempty"`
	TreeEntry     *sourcegraph.TreeEntry     `json:",omitempty"`
	Def           *sourcegraph.Def           `json:",omitempty"`
	DefList       *sourcegraph.DefList       `json:",omitempty"`
	Examples      *sourcegraph.ExampleList   `json:",omitempty"`
}

// Mock sets the services and the API client to their mock implementation, as well
// as mocks out used functions and decodes the body of the request.
// Whenever new payloads are added that use services which aren't mocked, the
// mockPayload structure and the corresponding mocks should be added here as this
// will constantly grow as new functionalities are added.
//
// Any newly added mocks should never assume that the request body will contain a mock
// for them and should be able to fall back to valid default values. Please refer to
// current mock implementations for reference.
//
// WARNING: This method alters global state by overwriting handlerutil.APIClient and
// handlerutil.Service which will cause all callers to use mock data. Due to this,
// all tests run while using this state will have to be made sequentially and not
// concurrently or in parallel.
func (sm *serviceMocker) Mock(r *http.Request) error {
	mock := new(mockPayload)
	if err := json.NewDecoder(r.Body).Decode(mock); err != nil {
		log.Println("warning: request body could not be decoded. Using default 'mock' data.")
	}
	defer r.Body.Close()

	log.Printf("MOCKING %s\n", r.URL)

	var mocks httptestutil.MockClients
	handlerutil.APIClient = func(r *http.Request) *sourcegraph.Client { return mocks.Client() }

	// Users, etc.
	mocks.Users.Get_ = func(context.Context, *sourcegraph.UserSpec) (*sourcegraph.User, error) {
		return nil, nil
	}
	mocks.Defs.List_ = func(ctx context.Context, opt *sourcegraph.DefListOptions) (*sourcegraph.DefList, error) {
		defs := []*sourcegraph.Def{
			&sourcegraph.Def{
				Def: graph.Def{
					DefStart: 0,
					DefEnd:   5,
					Name:     "ABC",
					DefKey: graph.DefKey{
						Unit:     "unit",
						UnitType: "GoPackage",
						Repo:     "github.com/gbbr/gomez",
						CommitID: "abcdefghijklmnopqrstuv",
					},
					Data: []byte("{}"),
				},
			},
			&sourcegraph.Def{
				Def: graph.Def{
					DefStart: 8,
					DefEnd:   14,
					Name:     "QWE",
					DefKey: graph.DefKey{
						Unit:     "unit",
						UnitType: "GoPackage",
						Repo:     "github.com/gbbr/gomez",
						CommitID: "abcdefghijklmnopqrstuv",
					},
					Data: []byte("{}"),
				},
			},
		}
		if mock.DefList != nil {
			return mock.DefList, nil
		}
		return &sourcegraph.DefList{Defs: defs}, nil
	}

	// Repos
	mocks.MirrorRepos.RefreshVCS_ = func(context.Context, *sourcegraph.MirrorReposRefreshVCSOp) (*pbtypes.Void, error) { return nil, nil }
	mocks.Repos.Get_ = func(ctx context.Context, _ *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		repo := new(sourcegraph.Repo)
		if mock.Repo != nil {
			repo = mock.Repo
		}
		return repo, nil
	}
	mocks.Repos.GetConfig_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.RepoConfig, error) {
		config := new(sourcegraph.RepoConfig)
		if mock.RepoConfig != nil {
			config = mock.RepoConfig
		}
		return config, nil
	}
	mocks.Repos.GetCommit_ = func(ctx context.Context, rev *sourcegraph.RepoRevSpec) (*vcs.Commit, error) {
		commit := new(vcs.Commit)
		if mock.Commit != nil {
			commit = mock.Commit
		}
		return commit, nil
	}

	// BuildsService
	mocks.Builds.GetRepoBuildInfo_ = func(ctx context.Context, op *sourcegraph.BuildsGetRepoBuildInfoOp) (*sourcegraph.RepoBuildInfo, error) {
		defaultLast := &sourcegraph.Build{
			CommitID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		}
		buildInfo := &sourcegraph.RepoBuildInfo{LastSuccessful: defaultLast}
		if mock.RepoBuildInfo != nil {
			buildInfo = mock.RepoBuildInfo
			if buildInfo.LastSuccessful == nil {
				buildInfo.LastSuccessful = defaultLast
			} else if buildInfo.LastSuccessful.CommitID == "" {
				buildInfo.LastSuccessful.CommitID = defaultLast.CommitID
			}
		}
		return buildInfo, nil
	}

	// RepoTreeService
	mocks.RepoTree.Get_ = func(ctx context.Context, op *sourcegraph.RepoTreeGetOp) (*sourcegraph.TreeEntry, error) {
		entry := &sourcegraph.TreeEntry{
			TreeEntry: &vcsclient.TreeEntry{
				Type: vcsclient.FileEntry,
			},
		}
		if mock.TreeEntry != nil {
			entry = mock.TreeEntry
		}
		return entry, nil
	}

	// DefsService
	mocks.Defs.Get_ = func(ctx context.Context, op *sourcegraph.DefsGetOp) (*sourcegraph.Def, error) {
		d := &sourcegraph.Def{
			Def: graph.Def{
				DefStart: 0,
				DefEnd:   5,
				Name:     "ABC",
				DefKey: graph.DefKey{
					Unit:     "unit",
					UnitType: "GoPackage",
					Repo:     "repo",
					CommitID: "abcdefghijklmnopqrstuv",
				},
				Data: []byte("{}"),
			},
		}
		if mock.Def != nil {
			d = mock.Def
		}
		return d, nil
	}
	mocks.Defs.ListExamples_ = func(ctx context.Context, op *sourcegraph.DefsListExamplesOp) (*sourcegraph.ExampleList, error) {
		ex := &sourcegraph.Example{
			Ref: graph.Ref{
				DefRepo:     "github.com/gbbr/gomez",
				DefUnitType: "GoPackage",
				DefUnit:     "github.com/gbbr/gomez/smtp",
				DefPath:     "transaction",
				Repo:        "github.com/gbbr/gomez",
				CommitID:    "ab725e29f93c8450a8565327e06bb810e247d053",
				UnitType:    "GoPackage",
				Unit:        "github.com/gbbr/gomez/smtp",
				Def:         false,
				File:        "smtp/commands.go",
			},
			SourceCode: &sourcegraph.SourceCode{
				Lines: []*sourcegraph.SourceCodeLine{
					&sourcegraph.SourceCodeLine{
						StartByte: 0,
						EndByte:   150,
						Tokens: []*sourcegraph.SourceCodeToken{
							{
								Label: "space",
							},
							{
								Class: "pln",
								Label: "zxc",
							},
							{
								URL:   []string{"a"},
								IsDef: true,
								Class: "iii",
								Label: "asd",
							},
						},
					},
				},
			},
			StartLine: 0,
			EndLine:   5,
			Error:     false,
		}
		if mock.Examples != nil {
			return mock.Examples, nil
		}
		return &sourcegraph.ExampleList{Examples: []*sourcegraph.Example{ex}}, nil
	}
	return nil
}
