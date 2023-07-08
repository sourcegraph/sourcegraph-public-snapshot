package graphqlbackend

import (
	"context"
	"html/template"

	"github.com/gogo/protobuf/jsonpb"

	"github.com/sourcegraph/sourcegraph/internal/gosyntect"
	"github.com/sourcegraph/sourcegraph/internal/highlight"
	searchresult "github.com/sourcegraph/sourcegraph/internal/search/result"
)

type highlightedRangeResolver struct {
	inner searchresult.HighlightedRange
}

func (h highlightedRangeResolver) Line() int32      { return h.inner.Line }
func (h highlightedRangeResolver) Character() int32 { return h.inner.Character }
func (h highlightedRangeResolver) Length() int32    { return h.inner.Length }

type highlightedStringResolver struct {
	inner searchresult.HighlightedString
}

func (s *highlightedStringResolver) Value() string { return s.inner.Value }
func (s *highlightedStringResolver) Highlights() []highlightedRangeResolver {
	res := make([]highlightedRangeResolver, len(s.inner.Highlights))
	for i, hl := range s.inner.Highlights {
		res[i] = highlightedRangeResolver{hl}
	}
	return res
}

type HighlightArgs struct {
	DisableTimeout     bool
	IsLightTheme       *bool
	HighlightLongLines bool
	Format             string
	StartLine          *int32
	EndLine            *int32
}

type HighlightedFileResolver struct {
	aborted  bool
	response *highlight.HighlightedCode
}

func (h *HighlightedFileResolver) Aborted() bool { return h.aborted }
func (h *HighlightedFileResolver) HTML() string {
	html, err := h.response.HTML()
	if err != nil {
		return ""
	}

	return string(html)
}
func (h *HighlightedFileResolver) LSIF() string {
	if h.response == nil {
		return "{}"
	}

	marshaller := &jsonpb.Marshaler{
		EnumsAsInts:  true,
		EmitDefaults: false,
	}

	// TODO(tjdevries): We could probably serialize the error, but it wouldn't do anything for now.
	lsif, err := marshaller.MarshalToString(h.response.LSIF())
	if err != nil {
		return "{}"
	}

	return lsif
}
func (h *HighlightedFileResolver) LineRanges(args *struct{ Ranges []highlight.LineRange }) ([][]string, error) {
	if h.response != nil && h.response.LSIF() != nil {
		return h.response.LinesForRanges(args.Ranges)
	}

	return highlight.SplitLineRanges(template.HTML(h.HTML()), args.Ranges)
}

func highlightContent(ctx context.Context, args *HighlightArgs, content, path string, metadata highlight.Metadata) (*HighlightedFileResolver, error) {
	var (
		resolver        = &HighlightedFileResolver{}
		err             error
		simulateTimeout = metadata.RepoName == "github.com/sourcegraph/AlwaysHighlightTimeoutTest"
	)

	response, aborted, err := highlight.Code(ctx, highlight.Params{
		Content:            []byte(content),
		Filepath:           path,
		DisableTimeout:     args.DisableTimeout,
		HighlightLongLines: args.HighlightLongLines,
		SimulateTimeout:    simulateTimeout,
		Metadata:           metadata,
		Format:             gosyntect.GetResponseFormat(args.Format),
	})

	resolver.aborted = aborted
	resolver.response = response

	if err != nil {
		return nil, err
	}

	return resolver, nil
}
