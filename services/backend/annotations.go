package backend

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/neelance/parallel"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	approuter "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	annotationspkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/annotations"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/syntaxhighlight"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
)

var Annotations = &annotations{}

type annotations struct{}

func (s *annotations) List(ctx context.Context, opt *sourcegraph.AnnotationsListOptions) (*sourcegraph.AnnotationList, error) {
	if Mocks.Annotations.List != nil {
		return Mocks.Annotations.List(ctx, opt)
	}

	var fileRange sourcegraph.FileRange
	if opt.Range != nil {
		fileRange = *opt.Range
	}

	if !isAbsCommitID(opt.Entry.RepoRev.CommitID) {
		return nil, errNotAbsCommitID
	}

	entry, err := RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: opt.Entry,
		Opt: &sourcegraph.RepoTreeGetOptions{
			GetFileOptions: sourcegraph.GetFileOptions{
				FileRange: fileRange,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var (
		mu      sync.Mutex
		allAnns []*sourcegraph.Annotation
	)
	addAnns := func(anns []*sourcegraph.Annotation) {
		mu.Lock()
		allAnns = append(allAnns, anns...)
		mu.Unlock()
	}

	funcs := []func(context.Context, *sourcegraph.AnnotationsListOptions, *sourcegraph.TreeEntry) ([]*sourcegraph.Annotation, error){
		s.listSyntaxHighlights,
	}

	if !opt.NoSrclibAnns {
		funcs = append(funcs, s.listRefs)
	}

	par := parallel.NewRun(len(funcs))
	for _, f := range funcs {
		f2 := f
		par.Acquire()
		go func() {
			defer par.Release()
			anns, err := f2(ctx, opt, entry)
			if err != nil {
				par.Error(err)
				return
			}
			addAnns(anns)
		}()
	}
	if err := par.Wait(); err != nil {
		return nil, err
	}

	allAnns = annotationspkg.Prepare(allAnns)

	return &sourcegraph.AnnotationList{
		Annotations:    allAnns,
		LineStartBytes: computeLineStartBytes(entry.Contents),
	}, nil
}

func (s *annotations) listSyntaxHighlights(ctx context.Context, opt *sourcegraph.AnnotationsListOptions, entry *sourcegraph.TreeEntry) ([]*sourcegraph.Annotation, error) {
	if opt.Range == nil {
		opt.Range = &sourcegraph.FileRange{}
	}

	lexer := selectAppropriateLexer(opt.Entry.Path)
	if lexer == nil {
		return nil, nil
	}

	var c syntaxhighlight.TokenCollectorAnnotator
	if _, err := syntaxhighlight.Annotate(entry.Contents, lexer, &c); err != nil {
		return nil, err
	}

	anns := make([]*sourcegraph.Annotation, 0, len(c.Tokens))
	for _, tok := range c.Tokens {
		if class := syntaxhighlight.DefaultHTMLConfig.GetTokenClass(tok); class != "" {
			anns = append(anns, &sourcegraph.Annotation{
				StartByte: uint32(opt.Range.StartByte) + uint32(tok.Offset),
				EndByte:   uint32(opt.Range.StartByte) + uint32(tok.Offset+len(tok.Text)),
				Class:     class,
			})
		}
	}
	return anns, nil
}

// selectAppropriateLexer selects an appropriate lexer to use given the file path.
// It returns nil if the file should not be annotated.
func selectAppropriateLexer(path string) syntaxhighlight.Lexer {
	ext := filepath.Ext(path)
	if ext == "" {
		// Files with no extensions (e.g., AUTHORS, README, LICENSE, etc.) don't get annotated.
		return nil
	}
	lexer := syntaxhighlight.NewLexerByExtension(ext)
	if lexer == nil {
		// Use a fallback lexer for other types.
		lexer = &syntaxhighlight.FallbackLexer{}
	}
	return lexer
}

func (s *annotations) listRefs(ctx context.Context, opt *sourcegraph.AnnotationsListOptions, entry *sourcegraph.TreeEntry) ([]*sourcegraph.Annotation, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Annotations.listRefs", opt.Entry.RepoRev.Repo); err != nil {
		return nil, err
	}

	if opt.Entry.Path == "" {
		return nil, fmt.Errorf("listRefs: no file path specified for file in %v", opt.Entry.RepoRev)
	}

	repo, err := (&repos{}).Get(ctx, &sourcegraph.RepoSpec{ID: opt.Entry.RepoRev.Repo})
	if err != nil {
		return nil, err
	}

	filters := []srcstore.RefFilter{
		srcstore.ByRepos(repo.URI),
		srcstore.ByCommitIDs(opt.Entry.RepoRev.CommitID),
		srcstore.ByFiles(true, opt.Entry.Path),
	}
	if opt.Range != nil {
		start := uint32(opt.Range.StartByte)
		end := uint32(opt.Range.EndByte)
		if start != 0 {
			filters = append(filters, srcstore.RefFilterFunc(func(ref *graph.Ref) bool {
				return ref.Start >= start
			}))
		}
		if end != 0 {
			filters = append(filters, srcstore.RefFilterFunc(func(ref *graph.Ref) bool {
				return ref.End <= end
			}))
		}
	}

	refs, err := localstore.Graph.Refs(filters...)
	if err != nil {
		return nil, err
	}

	anns := make([]*sourcegraph.Annotation, len(refs))
	for i, ref := range refs {
		def := ref.DefKey()
		var u string
		if strings.HasPrefix(def.Path, "https://") || strings.HasPrefix(def.Path, "http://") {
			// If the ref's def path is an absolute URL, set the annotation's URL to be that.
			// This is used in some languages (e.g., JavaScript, CSS) for references to builtin
			// or standard library definitions.
			u = def.Path
		} else {
			u = approuter.Rel.URLToDefKey(def).String()
		}

		anns[i] = &sourcegraph.Annotation{
			URL:       u,
			Def:       ref.Def,
			StartByte: ref.Start,
			EndByte:   ref.End,
		}
	}
	return anns, nil
}

func computeLineStartBytes(data []byte) []uint32 {
	if len(data) == 0 {
		return []uint32{}
	}
	pos := []uint32{0}
	for i, b := range data {
		if b == '\n' {
			pos = append(pos, uint32(i+1))
		}
	}
	return pos
}

func (s *annotations) GetDefAtPos(ctx context.Context, opt *sourcegraph.AnnotationsGetDefAtPosOptions) (*sourcegraph.DefSpec, error) {
	if Mocks.Annotations.GetDefAtPos != nil {
		return Mocks.Annotations.GetDefAtPos(ctx, opt)
	}

	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Annotations.GetDefAtPos", opt.Entry.RepoRev.Repo); err != nil {
		return nil, err
	}

	if opt.Entry.Path == "" {
		return nil, fmt.Errorf("GetDefAtPos: no file path specified for file in %v", opt.Entry.RepoRev)
	}

	repo, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: opt.Entry.RepoRev.Repo})
	if err != nil {
		return nil, err
	}

	entry, err := RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: opt.Entry,
	})
	if err != nil {
		return nil, err
	}

	offset := uint32(0)
	contents := entry.Contents
	for line := opt.Line; line > 0; line-- {
		eol := bytes.IndexByte(contents, '\n')
		if eol == -1 {
			break
		}
		offset += uint32(eol + 1)
		contents = contents[eol+1:]
	}
	offset += opt.Character

	filters := []srcstore.RefFilter{
		srcstore.ByRepos(repo.URI),
		srcstore.ByCommitIDs(opt.Entry.RepoRev.CommitID),
		srcstore.ByFiles(true, opt.Entry.Path),
		srcstore.RefFilterFunc(func(ref *graph.Ref) bool {
			return ref.Start <= offset && ref.End > offset
		}),
	}

	refs, err := localstore.Graph.Refs(filters...)
	if err != nil {
		return nil, err
	}

	if len(refs) == 0 {
		return nil, grpc.Errorf(codes.NotFound, "no ref found")
	}

	r := refs[0]
	defRepo, err := localstore.Repos.GetByURI(ctx, r.DefRepo)
	if err != nil {
		return nil, err
	}

	var defCommitID string
	switch defRepo.ID {
	case opt.Entry.RepoRev.Repo:
		defCommitID = opt.Entry.RepoRev.CommitID
	default:
		defaultRev, err := Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
			Repo: defRepo.ID,
		})
		if err != nil {
			return nil, err
		}
		defCommitID = defaultRev.CommitID
	}

	return &sourcegraph.DefSpec{
		Repo:     defRepo.ID,
		CommitID: defCommitID,
		UnitType: r.DefUnitType,
		Unit:     r.DefUnit,
		Path:     r.DefPath,
	}, nil
}

type MockAnnotations struct {
	List        func(v0 context.Context, v1 *sourcegraph.AnnotationsListOptions) (*sourcegraph.AnnotationList, error)
	GetDefAtPos func(v0 context.Context, v1 *sourcegraph.AnnotationsGetDefAtPosOptions) (*sourcegraph.DefSpec, error)
}

func (s *MockAnnotations) MockList(t *testing.T, wantAnns ...*sourcegraph.Annotation) (called *bool) {
	called = new(bool)
	s.List = func(ctx context.Context, opt *sourcegraph.AnnotationsListOptions) (*sourcegraph.AnnotationList, error) {
		*called = true
		return &sourcegraph.AnnotationList{Annotations: wantAnns}, nil
	}
	return
}
