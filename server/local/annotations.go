package local

import (
	"sort"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/rogpeppe/rog-go/parallel"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
	approuter "src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/syntaxhighlight"
)

var Annotations sourcegraph.AnnotationsServer = &annotations{}

type annotations struct{}

func (s *annotations) List(ctx context.Context, opt *sourcegraph.AnnotationsListOptions) (*sourcegraph.AnnotationList, error) {
	entry, err := svc.RepoTree(ctx).Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: opt.Entry,
		Opt: &sourcegraph.RepoTreeGetOptions{
			GetFileOptions: sourcegraph.GetFileOptions{
				FileRange: *opt.Range,
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

	sort.Sort(sortableAnnotations(allAnns))

	return &sourcegraph.AnnotationList{
		Annotations:    allAnns,
		LineStartBytes: computeLineStartBytes(entry.Contents),
	}, nil
}

func (s *annotations) listSyntaxHighlights(ctx context.Context, opt *sourcegraph.AnnotationsListOptions, entry *sourcegraph.TreeEntry) ([]*sourcegraph.Annotation, error) {
	if opt.Range == nil {
		opt.Range = &sourcegraph.FileRange{}
	}

	var c syntaxhighlight.TokenCollectorAnnotator
	if _, err := syntaxhighlight.Annotate(entry.Contents, opt.Entry.Path, "", &c); err != nil {
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

func (s *annotations) listRefs(ctx context.Context, opt *sourcegraph.AnnotationsListOptions, entry *sourcegraph.TreeEntry) ([]*sourcegraph.Annotation, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Annotations.listRefs", opt.Entry.RepoRev.URI); err != nil {
		return nil, err
	}

	dataVer, err := svc.Repos(ctx).GetSrclibDataVersionForPath(ctx, &opt.Entry)
	if grpc.Code(err) == codes.NotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	filters := []srcstore.RefFilter{
		srcstore.ByRepos(opt.Entry.RepoRev.URI),
		srcstore.ByCommitIDs(dataVer.CommitID),
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
		var rev string
		if ref.Repo == opt.Entry.RepoRev.URI {
			rev = opt.Entry.RepoRev.Rev
		}

		anns[i] = &sourcegraph.Annotation{
			URL:       approuter.Rel.URLToDefAtRev(ref.DefKey(), rev).String(),
			StartByte: ref.Start,
			EndByte:   ref.End,
		}
	}
	return anns, nil
}

type sortableAnnotations []*sourcegraph.Annotation

func (a sortableAnnotations) Len() int      { return len(a) }
func (a sortableAnnotations) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a sortableAnnotations) Less(i, j int) bool {
	return a[i].StartByte < a[j].StartByte ||
		(a[i].StartByte == a[j].StartByte && (a[i].EndByte < a[j].EndByte ||
			(a[i].EndByte == a[j].EndByte && a[i].URL != "")))
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
