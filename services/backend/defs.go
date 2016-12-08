package backend

import (
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/inconshreveable/log15.v2"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/htmlutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
)

var Defs = &defs{}

type defs struct{}

func (s *defs) Get(ctx context.Context, op *sourcegraph.DefsGetOp) (res *sourcegraph.Def, err error) {
	if Mocks.Defs.Get != nil {
		return Mocks.Defs.Get(ctx, op)
	}

	ctx, done := trace(ctx, "Defs", "Get", op, &err)
	defer done()

	defSpec := op.Def

	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Defs.Get", defSpec.Repo); err != nil {
		return nil, err
	}

	if !isAbsCommitID(defSpec.CommitID) {
		return nil, legacyerr.Errorf(legacyerr.InvalidArgument, "absolute commit ID required (got %q)", defSpec.CommitID)
	}

	rawDef, err := s.get(ctx, defSpec)
	if err != nil {
		return nil, err
	}
	def := &sourcegraph.Def{Def: *rawDef}
	if op.Opt == nil {
		op.Opt = &sourcegraph.DefGetOptions{}
	}
	if op.Opt.Doc {
		def.DocHTML = htmlutil.EmptyHTML()
		if len(def.Docs) > 0 {
			def.DocHTML = htmlutil.Sanitize(def.Docs[0].Data)
		}
	}
	if op.Opt.ComputeLineRange {
		startLine, endLine, err := computeLineRange(ctx, sourcegraph.TreeEntrySpec{
			RepoRev: sourcegraph.RepoRevSpec{Repo: defSpec.Repo, CommitID: defSpec.CommitID},
			Path:    def.File,
		}, def.DefStart, def.DefEnd)
		if err != nil {
			log15.Warn("Defs.Get: failed to compute line range.", "err", err, "repo", defSpec.Repo, "commitID", defSpec.CommitID, "file", def.File)
		}
		def.StartLine = startLine
		def.EndLine = endLine
	}
	populateDefFormatStrings(def)
	return def, nil
}

func computeLineRange(ctx context.Context, entrySpec sourcegraph.TreeEntrySpec, startByte, endByte uint32) (startLine, endLine uint32, err error) {
	entry, err := (&repoTree{}).Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: entrySpec,
	})
	if err != nil {
		return
	}

	const max = 1024 * 1024 // 1 MB max size
	if len(entry.Contents) > max {
		err = fmt.Errorf("file exceeds max size (%d bytes)", max)
		return
	}

	line := uint32(1)
	for i, c := range entry.Contents {
		if uint32(i) == startByte {
			startLine = line
		}
		if uint32(i) == endByte {
			endLine = line
			break
		}
		if c == '\n' {
			line++
		}
	}
	return
}

// get returns the def with the given def key (and no additional
// information, such as docs).
func (s *defs) get(ctx context.Context, def sourcegraph.DefSpec) (*graph.Def, error) {
	repo, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: def.Repo})
	if err != nil {
		return nil, err
	}

	d, err := localstore.Graph.Defs(srcstore.ByDefKey(def.DefKey(repo.URI)))
	if err != nil {
		return nil, err
	}
	if len(d) == 0 {
		return nil, legacyerr.Errorf(legacyerr.NotFound, "def %v not found", def)
	}
	return d[0], nil
}

func populateDefFormatStrings(def *sourcegraph.Def) {
	if _, present := graph.MakeDefFormatters[def.UnitType]; !present {
		return
	}
	f := def.Fmt()
	quals := func(fn func(graph.Qualification) string) graph.QualFormatStrings {
		return graph.QualFormatStrings{
			Unqualified:             fn(graph.Unqualified),
			ScopeQualified:          fn(graph.ScopeQualified),
			DepQualified:            fn(graph.DepQualified),
			RepositoryWideQualified: fn(graph.RepositoryWideQualified),
			LanguageWideQualified:   fn(graph.LanguageWideQualified),
		}
	}
	def.FmtStrings = &graph.DefFormatStrings{
		Name:                 quals(f.Name),
		Type:                 quals(f.Type),
		NameAndTypeSeparator: f.NameAndTypeSeparator(),
		Language:             f.Language(),
		DefKeyword:           f.DefKeyword(),
		Kind:                 f.Kind(),
	}
}

type MockDefs struct {
	Get              func(v0 context.Context, v1 *sourcegraph.DefsGetOp) (*sourcegraph.Def, error)
	ListRefs         func(v0 context.Context, v1 *sourcegraph.DeprecatedDefsListRefsOp) (*sourcegraph.RefList, error)
	ListRefLocations func(v0 context.Context, v1 *sourcegraph.DeprecatedDefsListRefLocationsOp) (*sourcegraph.DeprecatedRefLocationsList, error)
	RefreshIndex     func(v0 context.Context, v1 *sourcegraph.DefsRefreshIndexOp) error
}

func (s *MockDefs) MockGet(t *testing.T, wantDef sourcegraph.DefSpec) (called *bool) {
	called = new(bool)
	s.Get = func(ctx context.Context, op *sourcegraph.DefsGetOp) (*sourcegraph.Def, error) {
		*called = true
		def := op.Def
		if def != wantDef {
			t.Errorf("got def %+v, want %+v", def, wantDef)
			return nil, legacyerr.Errorf(legacyerr.NotFound, "def %v not found", wantDef)
		}
		return &sourcegraph.Def{Def: graph.Def{DefKey: def.DefKey("r")}}, nil
	}
	return
}

func (s *MockDefs) MockGet_Return(t *testing.T, wantDef *sourcegraph.Def) (called *bool) {
	called = new(bool)
	s.Get = func(ctx context.Context, op *sourcegraph.DefsGetOp) (*sourcegraph.Def, error) {
		*called = true
		def := op.Def
		if def != wantDef.DefSpec(def.Repo) {
			t.Errorf("got def %+v, want %+v", def, wantDef.DefSpec(def.Repo))
			return nil, legacyerr.Errorf(legacyerr.NotFound, "def %v not found", wantDef.DefKey)
		}
		return wantDef, nil
	}
	return
}

func (s *MockDefs) MockRefreshIndex(t *testing.T, wantOp *sourcegraph.DefsRefreshIndexOp) (called *bool) {
	called = new(bool)
	s.RefreshIndex = func(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) error {
		*called = true
		if !reflect.DeepEqual(op, wantOp) {
			t.Fatalf("unexpected DefsRefreshIndexOp, got %+v != %+v", op, wantOp)
		}
		return nil
	}
	return
}
