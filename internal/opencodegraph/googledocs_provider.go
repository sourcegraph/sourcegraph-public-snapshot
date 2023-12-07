package opencodegraph

import (
	"context"
	"fmt"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	RegisterProvider(googledocsProvider{})
}

type googledocsProvider struct{}

func (googledocsProvider) Name() string { return "googledocs" }

func (googledocsProvider) Capabilities(ctx context.Context, params schema.CapabilitiesParams) (*schema.CapabilitiesResult, error) {
	return &schema.CapabilitiesResult{
		Selector: []*schema.Selector{
			{ContentContains: "https://docs.google.com/document/d/"},
		},
	}, nil
}

func (googledocsProvider) Annotations(ctx context.Context, params schema.AnnotationsParams) (*schema.AnnotationsResult, error) {
	var result schema.AnnotationsResult

	docIDs, ranges := gdocIDs(params.Content)
	for i, docID := range docIDs {
		id := fmt.Sprintf("%s:%d", docID, i)
		item := &schema.OpenCodeGraphItem{
			Id:         id,
			Title:      "📝 GDoc: " + gdocTitles[docID],
			Url:        gdocURLPrefix + docID,
			Preview:    true,
			PreviewUrl: "https://docs.google.com/document/d/" + docID + "/preview",
		}

		result.Items = append(result.Items, item)
		result.Annotations = append(result.Annotations, &schema.OpenCodeGraphAnnotation{
			Item:  schema.OpenCodeGraphItemRef{Id: id},
			Range: ranges[i],
		})
	}

	return &result, nil
}

// TODO(sqs): fetch titles instead of hardcoding
var gdocTitles = map[string]string{
	"1Z1Yp7G61WYlQ1B4vO5-mIXVtmvzGmD7PqYHNBQV-2Ik": "Telemetry Export (ELE) rollout plan",
}

const gdocURLPrefix = "https://docs.google.com/document/d/"

var gdocURL = regexp.MustCompile(regexp.QuoteMeta(gdocURLPrefix) + `([\w-]+)`)

func gdocIDs(content string) (ids []string, ranges []schema.OpenCodeGraphRange) {
	return firstSubmatchNamesAndRanges(gdocURL, content)
}
