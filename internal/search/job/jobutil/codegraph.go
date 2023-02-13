package jobutil

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/graph"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CodeGraphSearchJob is an experimental search job for querying on code intel data and
// relationships.
type CodeGraphSearchJob struct {
	// CodeIntel should be provided by internal/search/graph.Store(). It may be nil if
	// unimplemented.
	CodeIntel graph.CodeIntelStore
	// SymbolSearch is a search job that should provide symbol results.
	SymbolSearch job.Job
	// Relationship is the code intel graph relationship to query on.
	Relationship query.SymbolRelationship
}

func (s *CodeGraphSearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

	if s.CodeIntel == nil {
		err = errors.New("CodeIntel graph search unimplemented")
		return nil, err
	}

	seenRanges := make(map[string]map[result.Range]struct{})
	var symbolSearchErrors error
	alert, err = s.SymbolSearch.Run(ctx, clients, streaming.StreamFunc(func(se streaming.SearchEvent) {
		if se.Results.Len() == 0 {
			return
		}
		for _, m := range se.Results {
			// Symbol results are always FileMatch
			fm, ok := m.(*result.FileMatch)
			if !ok {
				continue
			}
			if len(fm.Symbols) == 0 {
				continue
			}
			// For each symbol, get results with precise relationships to the
			// symbol.
			var locations []types.CodeIntelLocation
			var err error
			for _, symbol := range fm.Symbols {
				// TODO: We should paginate code graph searches
				req := types.CodeIntelRequestArgs{
					RepositoryID: int(fm.Repo.ID),
					Commit:       string(fm.CommitID),
					Path:         fm.Path,
					// symbols are 1-indexed but codeintel is 0-indexed
					Line:      symbol.Symbol.Line - 1,
					Character: symbol.Symbol.Character,
					Limit:     100,
					RawCursor: "",
				}

				switch s.Relationship {
				case query.SymbolRelationshipReferences:
					locations, err = s.CodeIntel.GetReferences(ctx, fm.Repo, req)

				case query.SymbolRelationshipImplements:
					locations, err = s.CodeIntel.GetImplementations(ctx, fm.Repo, req)

				default:
					err = errors.Newf("unknown relationship query %q", s.Relationship)
				}
			}
			if err != nil {
				symbolSearchErrors = errors.Append(symbolSearchErrors, err)
				continue
			}

			for _, l := range locations {
				// Range identifies this result
				r := result.Range{
					Start: result.Location{
						Column: l.TargetRange.Start.Character,
						Line:   l.TargetRange.Start.Line,
					},
					End: result.Location{
						Column: l.TargetRange.End.Character,
						Line:   l.TargetRange.End.Line,
					},
				}
				// Deduplicate results
				if seenRanges[l.Path] == nil {
					seenRanges[l.Path] = map[result.Range]struct{}{r: {}}
				} else if _, seen := seenRanges[l.Path][r]; seen {
					continue
				}

				// TODO: Right now, we just return the result as a chunk match
				// because we do not get the actual symbol at the location. We
				// probably want to be able to adapt these back to symbol results.
				f, err := clients.Gitserver.ReadFile(ctx, authz.DefaultSubRepoPermsChecker,
					fm.Repo.Name, api.CommitID(l.TargetCommit), l.Path)
				if err != nil {
					symbolSearchErrors = errors.Append(symbolSearchErrors, err)
					continue
				}
				match := &result.FileMatch{
					File: result.File{
						Repo:     fm.Repo,
						CommitID: api.CommitID(l.TargetCommit),
						InputRev: fm.InputRev,
						Path:     l.Path,
					},
					ChunkMatches: result.ChunkMatches{
						result.ChunkMatch{
							Content: string(f),
							ContentStart: result.Location{
								Line: l.TargetRange.Start.Line,
							},
							Ranges: result.Ranges{r},
						},
					},
				}
				stream.Send(streaming.SearchEvent{Results: result.Matches{match}})
			}
		}
	}))
	if err != nil {
		return alert, err
	}
	if symbolSearchErrors != nil {
		return alert, err
	}

	return nil, nil
}

func (s *CodeGraphSearchJob) Children() []job.Describer {
	return []job.Describer{s.SymbolSearch}
}

func (s *CodeGraphSearchJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *s
	cp.SymbolSearch = job.Map(cp.SymbolSearch, fn)
	return &cp
}

func (s *CodeGraphSearchJob) Fields(v job.Verbosity) (res []log.Field) {
	return []log.Field{
		log.Bool("implemented", s.CodeIntel != nil),
		log.String("relationship", string(s.Relationship)),
	}
}

func (s *CodeGraphSearchJob) Name() string { return "CodeGraphSearchJob" }
