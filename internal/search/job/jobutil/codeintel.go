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

type CodeIntelSearchJob struct {
	SymbolSearch job.Job
	Relationship query.SymbolRelationship
}

type streamingSenderFunc func(streaming.SearchEvent)

func (s streamingSenderFunc) Send(e streaming.SearchEvent) { s(e) }

func (s *CodeIntelSearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

	seenRanges := make(map[string]map[result.Range]struct{})
	var symbolSearchErrors error
	alert, err = s.SymbolSearch.Run(ctx, clients, streaming.StreamFunc(func(se streaming.SearchEvent) {
		if se.Results.Len() == 0 {
			return
		}
		for _, m := range se.Results {
			switch fm := m.(type) {
			case *result.FileMatch:
				codeintel := graph.Store()
				if codeintel == nil {
					return
				}

				if len(fm.Symbols) > 0 {
					var locations []types.CodeIntelLocation
					var err error
					for _, symbol := range fm.Symbols {
						req := types.CodeIntelRequestArgs{
							RepositoryID: int(fm.Repo.ID),
							Commit:       string(fm.CommitID),
							Path:         fm.Path,
							Line:         symbol.Symbol.Line - 1,
							Character:    symbol.Symbol.Character,
							Limit:        100,
							RawCursor:    "",
						}

						switch s.Relationship {
						case query.SymbolRelationshipReferences:
							locations, err = codeintel.GetReferences(ctx, fm.Repo, req)

						case query.SymbolRelationshipImplements:
							locations, err = codeintel.GetImplementations(ctx, fm.Repo, req)

						// TODO: case "callers"

						default:
							err = errors.Newf("unknown relationship query %q", s.Relationship)
						}
					}
					if err != nil {
						symbolSearchErrors = errors.Append(symbolSearchErrors, err)
						continue
					}
					for _, l := range locations {
						r := result.Range{
							Start: result.Location{
								Offset: l.TargetRange.Start.Character,
								Line:   l.TargetRange.Start.Line,
							},
							End: result.Location{
								Offset: l.TargetRange.End.Character,
								Line:   l.TargetRange.End.Line,
							},
						}
						if seenRanges[l.Path] == nil {
							seenRanges[l.Path] = map[result.Range]struct{}{r: {}}
						} else if _, seen := seenRanges[l.Path][r]; seen {
							continue
						}

						f, err := clients.Gitserver.ReadFile(ctx, authz.DefaultSubRepoPermsChecker,
							fm.Repo.Name, api.CommitID(l.TargetCommit), l.Path)
						if err != nil {
							symbolSearchErrors = errors.Append(symbolSearchErrors, err)
							continue
						}
						stream.Send(streaming.SearchEvent{
							Results: result.Matches{
								&result.FileMatch{
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
								},
							},
						})
					}
				}
				return
			default:
				// ignore
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

func (s *CodeIntelSearchJob) Children() []job.Describer { return nil }

func (s *CodeIntelSearchJob) MapChildren(fn job.MapFunc) job.Job { return s }

func (s *CodeIntelSearchJob) Fields(v job.Verbosity) (res []log.Field) {
	return []log.Field{
		log.String("relationship", string(s.Relationship)),
	}
}
func (s *CodeIntelSearchJob) Name() string { return "CodeIntelSearchJob" }
