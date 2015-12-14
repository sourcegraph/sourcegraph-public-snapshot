package local

import (
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/srclib/graph"
	srcstore "sourcegraph.com/sourcegraph/srclib/store"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/util/htmlutil"
)

var Defs sourcegraph.DefsServer = &defs{}

type defs struct{}

var _ sourcegraph.DefsServer = (*defs)(nil)

func (s *defs) Get(ctx context.Context, op *sourcegraph.DefsGetOp) (*sourcegraph.Def, error) {
	defSpec := op.Def

	cacheOnCommitID(ctx, defSpec.CommitID)

	if !isAbsCommitID(defSpec.CommitID) {
		return nil, grpc.Errorf(codes.InvalidArgument, "absolute commit ID required (got %q)", defSpec.CommitID)
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
		def.DocHTML = htmlutil.EmptyForPB()
		if len(def.Docs) > 0 {
			def.DocHTML = htmlutil.SanitizeForPB(def.Docs[0].Data)
		}
	}
	populateDefFormatStrings(def)
	return def, nil
}

// get returns the def with the given def key (and no additional
// information, such as docs). If no such def exists, ErrDefNotExist
// is returned.
func (s *defs) get(ctx context.Context, def sourcegraph.DefSpec) (*graph.Def, error) {
	d, err := store.GraphFromContext(ctx).Defs(srcstore.ByDefKey(def.DefKey()))
	if err != nil {
		return nil, err
	}
	if len(d) == 0 {
		return nil, graph.ErrDefNotExist
	}
	return d[0], nil
}

func (s *defs) List(ctx context.Context, opt *sourcegraph.DefListOptions) (*sourcegraph.DefList, error) {
	if opt == nil {
		opt = &sourcegraph.DefListOptions{}
	}

	shortCache(ctx)

	// Eliminate repos that don't exist.
	origRepoRevs := opt.RepoRevs
	opt.RepoRevs = nil
	for _, repoRev := range origRepoRevs {
		repoURI, commitID := sourcegraph.ParseRepoAndCommitID(repoRev)

		// Dealias.
		rA, err := svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{URI: repoURI})
		if err != nil {
			log.Printf("Warning: dropping repo rev %q from defs list because repo or repo alias was not found: %s.", repoRev, err)
			continue
		}
		repoURI = rA.URI

		// Determine the commit ID to use, if it wasn't specified or
		// if it's a non-commit-ID revspec.
		if !isAbsCommitID(commitID) {
			return nil, grpc.Errorf(codes.InvalidArgument, "absolute commit ID required for repo %q to list defs (got %q)", repoURI, commitID)
		}

		// The repo exists and the commit ID is valid, so include it
		// in the query.
		opt.RepoRevs = append(opt.RepoRevs, repoRev)
	}
	if len(origRepoRevs) > 0 && len(opt.RepoRevs) == 0 {
		log.Printf("Warning: DefsService.List got a RepoRevs param %v but none of the specified repos exist. Returning empty defs list.", origRepoRevs)
		return &sourcegraph.DefList{}, nil
	}

	// TODO(merge-to-master): don't try to search ALL repos until we
	// have a global index. Add a CLI flag to switch this behavior.
	//
	// if len(opt.RepoRevs) == 0 && len(opt.DefKeys) == 0 {
	// 	log.Println("WARNING: Defs.List cancelled - def queries that are not scoped to specific repos are rejected temporarily until global index exists!")
	// 	return &sourcegraph.DefList{}, nil
	// }

	fs := opt.DefFilters()
	fs = append(fs, srcstore.DefsSortByKey{})
	defs0, err := store.GraphFromContext(ctx).Defs(fs...)
	if err != nil {
		return nil, err
	}

	var defs []*sourcegraph.Def
	for i, def0 := range defs0 {
		if i >= opt.Offset() && i < (opt.Offset()+opt.Limit()) {
			defs = append(defs, &sourcegraph.Def{Def: *def0})
		}
	}
	// End kludge
	total := len(defs0)

	if opt.Doc {
		for _, def := range defs {
			def.DocHTML = htmlutil.EmptyForPB()
			if len(def.Docs) > 0 {
				def.DocHTML = htmlutil.SanitizeForPB(def.Docs[0].Data)
			}
		}
	}

	for _, def := range defs {
		populateDefFormatStrings(def)
	}

	return &sourcegraph.DefList{
		Defs: defs,
		ListResponse: sourcegraph.ListResponse{
			Total: int32(total),
		},
	}, nil
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
