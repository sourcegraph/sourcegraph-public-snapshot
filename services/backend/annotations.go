package backend

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rogpeppe/rog-go/parallel"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	approuter "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	annotationspkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/annotations"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/syntaxhighlight"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
)

var Annotations sourcegraph.AnnotationsServer = &annotations{}

type annotations struct{}

func (s *annotations) List(ctx context.Context, opt *sourcegraph.AnnotationsListOptions) (*sourcegraph.AnnotationList, error) {
	var fileRange sourcegraph.FileRange
	if opt.Range != nil {
		fileRange = *opt.Range
	}

	if !isAbsCommitID(opt.Entry.RepoRev.CommitID) {
		return nil, errNotAbsCommitID
	}

	entry, err := svc.RepoTree(ctx).Get(ctx, &sourcegraph.RepoTreeGetOp{
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
		s.listRefs,
	}
	par := parallel.NewRun(len(funcs))
	for _, f := range funcs {
		f2 := f
		par.Do(func() error {
			anns, err := f2(ctx, opt, entry)
			if err != nil {
				return err
			}
			addAnns(anns)
			return nil
		})
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

	filters := []srcstore.RefFilter{
		srcstore.ByRepos(opt.Entry.RepoRev.Repo),
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

	refs, err := store.GraphFromContext(ctx).Refs(filters...)
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
