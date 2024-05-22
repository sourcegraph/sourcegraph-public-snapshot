package resolvers

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/markdown"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type CodeNavServiceResolver interface {
	GitBlobLSIFData(ctx context.Context, args *GitBlobLSIFDataArgs) (GitBlobLSIFDataResolver, error)
	// CodeGraphData is a newer API that is more SCIP-oriented.
	// The second parameter is called 'opts' and not 'args' to reflect
	// that it is not what is exactly provided as input from the GraphQL
	// client.
	CodeGraphData(ctx context.Context, opts *CodeGraphDataOpts) (*[]CodeGraphDataResolver, error)
}

type GitBlobLSIFDataArgs struct {
	Repo      *types.Repo
	Commit    api.CommitID
	Path      string
	ExactPath bool
	ToolName  string
}

func (a *GitBlobLSIFDataArgs) Options() shared.UploadMatchingOptions {
	matching := shared.RootMustEnclosePath
	if !a.ExactPath {
		matching = shared.RootEnclosesPathOrPathEnclosesRoot
	}
	return shared.UploadMatchingOptions{
		RepositoryID:       int(a.Repo.ID),
		Commit:             string(a.Commit),
		Path:               a.Path,
		RootToPathMatching: matching,
		Indexer:            a.ToolName,
	}
}

type GitBlobLSIFDataResolver interface {
	GitTreeLSIFDataResolver
	ToGitTreeLSIFData() (GitTreeLSIFDataResolver, bool)
	ToGitBlobLSIFData() (GitBlobLSIFDataResolver, bool)

	Stencil(ctx context.Context) ([]RangeResolver, error)
	Ranges(ctx context.Context, args *LSIFRangesArgs) (CodeIntelligenceRangeConnectionResolver, error)
	Definitions(ctx context.Context, args *LSIFQueryPositionArgs) (LocationConnectionResolver, error)
	References(ctx context.Context, args *LSIFPagedQueryPositionArgs) (LocationConnectionResolver, error)
	Implementations(ctx context.Context, args *LSIFPagedQueryPositionArgs) (LocationConnectionResolver, error)
	Prototypes(ctx context.Context, args *LSIFPagedQueryPositionArgs) (LocationConnectionResolver, error)
	Hover(ctx context.Context, args *LSIFQueryPositionArgs) (HoverResolver, error)
	VisibleIndexes(ctx context.Context) (_ *[]PreciseIndexResolver, err error)
	Snapshot(ctx context.Context, args *struct{ IndexID graphql.ID }) (_ *[]SnapshotDataResolver, err error)
}

type SnapshotDataResolver interface {
	Offset() int32
	Data() string
	Additional() *[]string
}

type LSIFRangesArgs struct {
	StartLine int32
	EndLine   int32
}

type LSIFQueryPositionArgs struct {
	Line      int32
	Character int32
	Filter    *string
}

type LSIFPagedQueryPositionArgs struct {
	LSIFQueryPositionArgs
	PagedConnectionArgs
	Filter *string
}

type (
	CodeIntelligenceRangeConnectionResolver = ConnectionResolver[CodeIntelligenceRangeResolver]
)

type CodeIntelligenceRangeResolver interface {
	Range(ctx context.Context) (RangeResolver, error)
	Definitions(ctx context.Context) (LocationConnectionResolver, error)
	References(ctx context.Context) (LocationConnectionResolver, error)
	Implementations(ctx context.Context) (LocationConnectionResolver, error)
	Hover(ctx context.Context) (HoverResolver, error)
}

type RangeResolver interface {
	Start() PositionResolver
	End() PositionResolver
}

type PositionResolver interface {
	Line() int32
	Character() int32
}

type (
	LocationConnectionResolver = PagedConnectionResolver[LocationResolver]
)

type LocationResolver interface {
	Resource() GitTreeEntryResolver
	Range() RangeResolver
	URL(ctx context.Context) (string, error)
	CanonicalURL() string
}

type HoverResolver interface {
	Markdown() Markdown
	Range() RangeResolver
}

type Markdown string

func (m Markdown) Text() string {
	return string(m)
}

func (m Markdown) HTML() (string, error) {
	return markdown.Render(string(m))
}

type GitTreeLSIFDataResolver interface {
	Diagnostics(ctx context.Context, args *LSIFDiagnosticsArgs) (DiagnosticConnectionResolver, error)
}

type (
	LSIFDiagnosticsArgs          = ConnectionArgs
	DiagnosticConnectionResolver = PagedConnectionWithTotalCountResolver[DiagnosticResolver]
)

type DiagnosticResolver interface {
	Severity() (*string, error)
	Code() (*string, error)
	Source() (*string, error)
	Message() (*string, error)
	Location(ctx context.Context) (LocationResolver, error)
}

type CodeGraphDataResolver interface {
	Provenance(ctx context.Context) (CodeGraphDataProvenance, error)
	Commit(ctx context.Context) (string, error)
	ToolInfo(ctx context.Context) (*CodeGraphToolInfo, error)
	// Pre-condition: args are Normalized.
	Occurrences(ctx context.Context, args *OccurrencesArgs) (SCIPOccurrenceConnectionResolver, error)
}

type CodeGraphDataProvenance string

const (
	ProvenancePrecise     CodeGraphDataProvenance = "PRECISE"
	ProvenanceSyntactic   CodeGraphDataProvenance = "SYNTACTIC"
	ProvenanceSearchBased CodeGraphDataProvenance = "SEARCH_BASED"
)

type CodeGraphDataProvenanceComparator struct {
	Equals *CodeGraphDataProvenance
}

type CodeGraphDataFilter struct {
	Provenance *CodeGraphDataProvenanceComparator
}

// String is meant as a debugging-only representation without round-trippability
func (f *CodeGraphDataFilter) String() string {
	if f != nil && f.Provenance != nil && f.Provenance.Equals != nil {
		return fmt.Sprintf("provenance == %s", string(*f.Provenance.Equals))
	}
	return ""
}

type CodeGraphDataArgs struct {
	Filter *CodeGraphDataFilter
}

func (args *CodeGraphDataArgs) Attrs() []attribute.KeyValue {
	if args == nil {
		return nil
	}
	return []attribute.KeyValue{attribute.String("args.filter", args.Filter.String())}
}

type ForEachProvenance[T any] struct {
	SearchBased T
	Syntactic   T
	Precise     T
}

func (a *CodeGraphDataArgs) ProvenancesForSCIPData() ForEachProvenance[bool] {
	var out ForEachProvenance[bool]
	if a == nil || a.Filter == nil || a.Filter.Provenance == nil || a.Filter.Provenance.Equals == nil {
		out.Syntactic = true
		out.Precise = true
	} else {
		p := *a.Filter.Provenance.Equals
		switch p {
		case ProvenancePrecise:
			out.Precise = true
		case ProvenanceSyntactic:
			out.Syntactic = true
		case ProvenanceSearchBased:
		}
	}
	return out
}

type CodeGraphDataOpts struct {
	Args   *CodeGraphDataArgs
	Repo   *types.Repo
	Commit api.CommitID
	Path   string
}

func (opts *CodeGraphDataOpts) Attrs() []attribute.KeyValue {
	return append([]attribute.KeyValue{attribute.String("repo", opts.Repo.String()),
		opts.Commit.Attr(),
		attribute.String("path", opts.Path)}, opts.Args.Attrs()...)
}

type CodeGraphToolInfo struct {
	Name_    *string
	Version_ *string
}

func (ti *CodeGraphToolInfo) Name() *string {
	return ti.Name_
}

func (ti *CodeGraphToolInfo) Version() *string {
	return ti.Version_
}

type OccurrencesArgs struct {
	First *int32
	After *string
}

// Normalize returns args for convenience of chaining
func (args *OccurrencesArgs) Normalize(maxPageSize int32) *OccurrencesArgs {
	if args == nil {
		*args = OccurrencesArgs{}
	}
	if args.First == nil || *args.First > maxPageSize {
		args.First = &maxPageSize
	}
	return args
}

type SCIPOccurrenceConnectionResolver interface {
	ConnectionResolver[SCIPOccurrenceResolver]
	PageInfo(ctx context.Context) (*graphqlutil.ConnectionPageInfo[SCIPOccurrenceResolver], error)
}

type SCIPOccurrenceResolver interface {
	Symbol() (*string, error)
	Range() (RangeResolver, error)
	Roles() (*[]SymbolRole, error)
}

type SymbolRole string

// ⚠️ CAUTION: These constants are part of the public GraphQL API
const (
	SymbolRoleDefinition        SymbolRole = "Definition"
	SymbolRoleReference         SymbolRole = "Reference"
	SymbolRoleForwardDefinition SymbolRole = "ForwardDefinition"
)
