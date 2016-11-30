package backend

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/neelance/parallel"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	annotationspkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/annotations"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/syntaxhighlight"
)

var Annotations = &annotations{}

type annotations struct{}

func (s *annotations) List(ctx context.Context, opt *sourcegraph.AnnotationsListOptions) (res *sourcegraph.AnnotationList, err error) {
	if Mocks.Annotations.List != nil {
		return Mocks.Annotations.List(ctx, opt)
	}

	ctx, done := trace(ctx, "Annotations", "List", opt, &err)
	defer done()

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

type MockAnnotations struct {
	List func(v0 context.Context, v1 *sourcegraph.AnnotationsListOptions) (*sourcegraph.AnnotationList, error)
}

func (s *MockAnnotations) MockList(t *testing.T, wantAnns ...*sourcegraph.Annotation) (called *bool) {
	called = new(bool)
	s.List = func(ctx context.Context, opt *sourcegraph.AnnotationsListOptions) (*sourcegraph.AnnotationList, error) {
		*called = true
		return &sourcegraph.AnnotationList{Annotations: wantAnns}, nil
	}
	return
}
