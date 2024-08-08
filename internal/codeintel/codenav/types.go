package codenav

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// visibleUpload pairs an upload visible from the current target commit with the
// current target path and a matcher to the data within the underlying index.
//
// Pre-condition: TargetPath must be a sub-path of Upload.Root
// TODO: Make the fields private and have this go through a New* function.
type visibleUpload struct {
	Upload        uploadsshared.CompletedUpload
	TargetPath    core.RepoRelPath
	TargetMatcher shared.Matcher
}

func (vu visibleUpload) TargetPathWithoutRoot() core.UploadRelPath {
	return core.NewUploadRelPath(vu.Upload, vu.TargetPath)
}

type qualifiedMonikerSet struct {
	monikers       []precise.QualifiedMonikerData
	monikerHashMap map[string]struct{}
}

func newQualifiedMonikerSet() *qualifiedMonikerSet {
	return &qualifiedMonikerSet{
		monikerHashMap: map[string]struct{}{},
	}
}

// add the given qualified moniker to the set if it is distinct from all elements
// currently in the set.
func (s *qualifiedMonikerSet) add(qualifiedMoniker precise.QualifiedMonikerData) {
	monikerHash := strings.Join([]string{
		qualifiedMoniker.PackageInformationData.Name,
		qualifiedMoniker.PackageInformationData.Version,
		qualifiedMoniker.MonikerData.Scheme,
		qualifiedMoniker.PackageInformationData.Manager,
		qualifiedMoniker.MonikerData.Identifier,
	}, ":")

	if _, ok := s.monikerHashMap[monikerHash]; ok {
		return
	}

	s.monikerHashMap[monikerHash] = struct{}{}
	s.monikers = append(s.monikers, qualifiedMoniker)
}

type RequestArgs struct {
	RepositoryID api.RepoID
	Commit       api.CommitID
	Limit        int
	RawCursor    string
}

func (args *RequestArgs) Attrs() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Int("repositoryID", int(args.RepositoryID)),
		attribute.String("commit", string(args.Commit)),
		attribute.Int("limit", args.Limit),
	}
}

type PositionalRequestArgs struct {
	RequestArgs
	Path      core.RepoRelPath
	Line      int
	Character int
}

type OccurrenceRequestArgs struct {
	RepositoryID api.RepoID
	Path         core.RepoRelPath
	Commit       api.CommitID
	Limit        int
	RawCursor    string
	Matcher      shared.Matcher
}

func (args *OccurrenceRequestArgs) RequestArgs() RequestArgs {
	return RequestArgs{
		RepositoryID: args.RepositoryID,
		Commit:       args.Commit,
		Limit:        args.Limit,
		RawCursor:    args.RawCursor,
	}
}

func (args *OccurrenceRequestArgs) Attrs() []attribute.KeyValue {
	return append(
		[]attribute.KeyValue{
			attribute.Int("repositoryID", int(args.RepositoryID)),
			attribute.String("commit", string(args.Commit)),
			attribute.Int("limit", args.Limit),
		}, args.Matcher.Attrs()...)
}

func (args *PositionalRequestArgs) Attrs() []attribute.KeyValue {
	return append(args.RequestArgs.Attrs(),
		attribute.String("path", args.Path.RawValue()),
		attribute.Int("line", args.Line),
		attribute.Int("character", args.Character),
	)
}

// DiagnosticAtUpload is a diagnostic from within a particular upload. The adjusted commit denotes
// the target commit for which the location was adjusted (the originally requested commit).
type DiagnosticAtUpload struct {
	shared.Diagnostic[core.RepoRelPath]
	Upload         uploadsshared.CompletedUpload
	AdjustedCommit string
	AdjustedRange  shared.Range
}

// AdjustedCodeIntelligenceRange stores definition, reference, and hover information for all ranges
// within a block of lines. The definition and reference locations have been adjusted to fit the
// target (originally requested) commit.
type AdjustedCodeIntelligenceRange struct {
	Range           shared.Range
	Definitions     []shared.UploadLocation
	References      []shared.UploadLocation
	Implementations []shared.UploadLocation
	HoverText       string
}

// PreciseCursor is a struct that holds the state necessary to resume a locations query from a second or
// subsequent request. This struct is used internally as a request-specific context object that is
// mutated as the locations request is fulfilled. This struct is serialized to JSON then base64
// encoded to make an opaque string that is handed to a future request to get the remainder of the
// result set.
type PreciseCursor struct {
	// the following fields...
	// track the current phase and offset within phase

	Phase                string `json:"p"`    // ""/"local", "remote", or "done"
	LocalUploadOffset    int    `json:"l_uo"` // number of consumed visible uploads
	LocalLocationOffset  int    `json:"l_lo"` // offset within locations of VisibleUploads[LocalUploadOffset:]
	RemoteUploadOffset   int    `json:"r_uo"` // number of searched (to completion) uploads
	RemoteLocationOffset int    `json:"r_lo"` // offset within locations of the current upload batch

	// the following fields...
	// track associated visible/definition uploads and current batch of referencing uploads

	VisibleUploads []CursorVisibleUpload `json:"vus"` // root uploads covering a particular code location
	DefinitionIDs  []int                 `json:"dus"` // identifiers of uploads defining relevant symbol names
	UploadIDs      []int                 `json:"rus"` // current batch of uploads in which to search

	// the following fields...
	// are populated during the local phase, used in the remote phase

	SymbolNames []string `json:"ss"` // symbol names extracted from visible uploads
	// SkipPathsByUploadID maps UploadID -> UploadRelPath.
	SkipPathsByUploadID map[int]string `json:"pm"` // paths to skip for particular uploads in the remote phase
}

type CursorVisibleUpload struct {
	UploadID      int           `json:"id"`
	TargetPath    string        `json:"path"`
	TargetMatcher CursorMatcher `json:"mt"`
}

type CursorMatcher struct {
	ExactSymbol string          `json:"sym"`
	Start       shared.Position `json:"s"`
	End         shared.Position `json:"e"`
	HasEnd      bool            `json:"he"`
}

func (m CursorMatcher) ToShared() shared.Matcher {
	if m.HasEnd {
		return shared.NewSCIPBasedMatcher(scip.Range{m.Start.ToSCIPPosition(), m.End.ToSCIPPosition()}, m.ExactSymbol)
	}
	return shared.NewStartPositionMatcher(m.Start.ToSCIPPosition())
}

func NewCursorMatcher(matcher shared.Matcher) CursorMatcher {
	if sym, range_, ok := matcher.SymbolBased(); ok {
		return CursorMatcher{
			// OK to use "" here as lookups based on "" are not allowed
			ExactSymbol: sym.UnwrapOr(""),
			Start:       shared.TranslatePosition(range_.Start),
			End:         shared.TranslatePosition(range_.End),
			HasEnd:      true,
		}
	}
	if pos, ok := matcher.PositionBased(); ok {
		return CursorMatcher{
			ExactSymbol: "",
			Start:       shared.TranslatePosition(pos),
			End:         shared.Position{},
			HasEnd:      false,
		}
	}
	panic(fmt.Sprintf("Unhandled case for matcher: %+v", matcher))
}

var exhaustedCursor = PreciseCursor{Phase: "done"}

func (c PreciseCursor) Encode() string {
	return encodeViaJSON(c)
}

func DecodeCursor(rawEncoded string) (PreciseCursor, error) {
	return decodeViaJSON[PreciseCursor](rawEncoded)
}

func (c PreciseCursor) BumpLocalLocationOffset(n, totalCount int) PreciseCursor {
	c.LocalLocationOffset += n
	if c.LocalLocationOffset >= totalCount {
		// We've consumed this upload completely. Skip it the next time we find
		// ourselves in this loop, and ensure that we start with a zero offset on
		// the next upload we process (if any).
		c.LocalUploadOffset++
		c.LocalLocationOffset = 0
	}

	return c
}

func (c PreciseCursor) BumpRemoteUploadOffset(n, totalCount int) PreciseCursor {
	c.RemoteUploadOffset += n
	if c.RemoteUploadOffset >= totalCount {
		// We've consumed all upload batches
		c.RemoteUploadOffset = -1
	}

	return c
}

func (c PreciseCursor) BumpRemoteUsageOffset(n, totalCount int) PreciseCursor {
	c.RemoteLocationOffset += n
	if c.RemoteLocationOffset >= totalCount {
		// We've consumed the locations for this set of uploads. Reset this slice value in the
		// cursor so that the next call to this function will query the new set of uploads to
		// search in while resolving the next page. We also ensure we start on a zero offset
		// for the next page of results for a fresh set of uploads (if any).
		c.UploadIDs = nil
		c.RemoteLocationOffset = 0
	}

	return c
}

// referencesCursor stores (enough of) the state of a previous References request used to
// calculate the offset into the result set to be returned by the current request.
type ReferencesCursor struct {
	CursorsToVisibleUploads []CursorToVisibleUpload        `json:"adjustedUploads"`
	OrderedMonikers         []precise.QualifiedMonikerData `json:"orderedMonikers"`
	Phase                   string                         `json:"phase"`
	LocalCursor             LocalCursor                    `json:"localCursor"`
	RemoteCursor            RemoteCursor                   `json:"remoteCursor"`
}

// ImplementationsCursor stores (enough of) the state of a previous Implementations request used to
// calculate the offset into the result set to be returned by the current request.
type ImplementationsCursor struct {
	CursorsToVisibleUploads       []CursorToVisibleUpload        `json:"visibleUploads"`
	OrderedImplementationMonikers []precise.QualifiedMonikerData `json:"orderedImplementationMonikers"`
	OrderedExportMonikers         []precise.QualifiedMonikerData `json:"orderedExportMonikers"`
	Phase                         string                         `json:"phase"`
	LocalCursor                   LocalCursor                    `json:"localCursor"`
	RemoteCursor                  RemoteCursor                   `json:"remoteCursor"`
}

// cursorAdjustedUpload
type CursorToVisibleUpload struct {
	UploadID              int             `json:"uploadID"`
	TargetPath            string          `json:"adjustedPath"`
	TargetPosition        shared.Position `json:"adjustedPosition"`
	TargetPathWithoutRoot string          `json:"adjustedPathInBundle"`
}

// localCursor is an upload offset and a location offset within that upload.
type LocalCursor struct {
	UploadOffset int `json:"uploadOffset"`
	// The location offset within the associated upload.
	LocationOffset int `json:"locationOffset"`
}

// RemoteCursor is an upload offset, the current batch of uploads, and a location offset within the batch of uploads.
type RemoteCursor struct {
	UploadOffset   int   `json:"batchOffset"`
	UploadBatchIDs []int `json:"uploadBatchIDs"`
	// The location offset within the associated batch of uploads.
	LocationOffset int `json:"locationOffset"`
}

type SyntacticCursor struct {
	SeenFiles []string `json:"files"`
}

type UsagesForSymbolResolvedArgs struct {
	// Symbol is either nil or all the fields are populated for the equality check.
	Symbol   *ResolvedSymbolComparator
	Repo     types.Repo
	CommitID api.CommitID
	Path     core.RepoRelPath
	Range    scip.Range
	Filter   *ResolvedUsagesFilter

	RemainingCount int32
	Cursor         UsagesCursor
}

type ResolvedUsagesFilter struct {
	Not        *ResolvedUsagesFilter
	Repository *ResolvedRepositoryFilter
}

type ResolvedRepositoryFilter struct {
	NameEquals string
	// Resolved from above name
	RepoEquals types.Repo
}

type ResolvedSymbolComparator struct {
	EqualsName       string
	EqualsProvenance CodeGraphDataProvenance
	EqualsSymbol     *scip.Symbol
}

type ResolvedSymbolNameComparator struct {
	Equals       string
	EqualsSymbol scip.SymbolInformation
}

func (s *ResolvedSymbolComparator) ProvenancesForSCIPData() ForEachProvenance[bool] {
	var out ForEachProvenance[bool]
	if s == nil {
		out.Precise = true
		out.Syntactic = true
		out.SearchBased = true
	} else {
		switch s.EqualsProvenance {
		case ProvenancePrecise:
			out.Precise = true
		case ProvenanceSyntactic:
			out.Syntactic = true
		case ProvenanceSearchBased:
			out.SearchBased = true
		}
	}
	return out
}

// CodeGraphDataProvenance corresponds to the matching type in the GraphQL API.
//
// Make sure this type maintains its marshaling/unmarshaling behavior in
// case the type definition is changed.
type CodeGraphDataProvenance string

const (
	ProvenancePrecise     CodeGraphDataProvenance = "PRECISE"
	ProvenanceSyntactic   CodeGraphDataProvenance = "SYNTACTIC"
	ProvenanceSearchBased CodeGraphDataProvenance = "SEARCH_BASED"
)

type ForEachProvenance[T any] struct {
	SearchBased T
	Syntactic   T
	Precise     T
}

// CursorType's string representation is used for debugging.
type CursorType string

const (
	CursorTypeDefinitions     CursorType = "definitions"
	CursorTypeImplementations CursorType = "implementations"
	CursorTypePrototypes      CursorType = "prototypes"
	CursorTypeReferences      CursorType = "references"
	CursorTypeSyntactic       CursorType = "syntactic"
	CursorTypeSearchBased     CursorType = "searchBased"
	CursorTypeDone            CursorType = "done"
)

type UsagesCursor struct {
	CursorType      CursorType      `json:"ty"`
	PreciseCursor   PreciseCursor   `json:"pc"`
	SyntacticCursor SyntacticCursor `json:"sc"` // TODO(GRAPH-696)
}

func (c UsagesCursor) IsPrecise() bool {
	switch c.CursorType {
	case CursorTypeDefinitions, CursorTypeImplementations, CursorTypePrototypes, CursorTypeReferences:
		return true
	default:
		return false
	}
}

func (c UsagesCursor) IsSyntactic() bool {
	return c.CursorType == CursorTypeSyntactic
}

func (c UsagesCursor) IsSearchBased() bool {
	return c.CursorType == CursorTypeSearchBased
}

func (c UsagesCursor) IsDone() bool {
	return c.CursorType == CursorTypeDone
}

func (c UsagesCursor) AdvanceCursor(nextCursor core.Option[UsagesCursor], provenances ForEachProvenance[bool]) UsagesCursor {
	if next, isSome := nextCursor.Get(); isSome {
		return next
	}
	if c.IsPrecise() && provenances.Syntactic {
		return UsagesCursor{CursorType: CursorTypeSyntactic}
	} else if (c.IsPrecise() || c.IsSyntactic()) && provenances.SearchBased {
		return UsagesCursor{CursorType: CursorTypeSearchBased}
	} else {
		return UsagesCursor{CursorType: CursorTypeDone}
	}
}

func InitialCursor(provenances ForEachProvenance[bool]) UsagesCursor {
	if provenances.Precise {
		return UsagesCursor{CursorType: CursorTypeDefinitions}
	} else if provenances.Syntactic {
		return UsagesCursor{CursorType: CursorTypeSyntactic}
	} else if provenances.SearchBased {
		return UsagesCursor{CursorType: CursorTypeSearchBased}
	} else {
		return UsagesCursor{CursorType: CursorTypeDone}
	}
}

func (c UsagesCursor) Encode() string {
	return encodeViaJSON(c)
}

func DecodeUsagesCursor(rawEncoded string) (UsagesCursor, error) {
	return decodeViaJSON[UsagesCursor](rawEncoded)
}

func encodeViaJSON[T any](t T) string {
	bytes, _ := json.Marshal(t)
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func decodeViaJSON[T any](rawEncoded string) (T, error) {
	var val T
	if rawEncoded == "" {
		return val, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(rawEncoded)
	if err != nil {
		return val, err
	}
	err = json.Unmarshal(raw, &val)
	return val, err
}
