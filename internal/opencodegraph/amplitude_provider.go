package opencodegraph

import (
	"context"
	"fmt"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	RegisterProvider(amplitudeProvider{})
}

type amplitudeProvider struct{}

func (amplitudeProvider) Name() string { return "amplitude" }

func (amplitudeProvider) Capabilities(ctx context.Context, params schema.CapabilitiesParams) (*schema.CapabilitiesResult, error) {
	return &schema.CapabilitiesResult{
		Selector: []*schema.Selector{
			{Path: "**/*.ts?(x)", ContentContains: "eventLogger.log"},
		},
	}, nil
}

func (amplitudeProvider) Annotations(ctx context.Context, params schema.AnnotationsParams) (*schema.AnnotationsResult, error) {
	var result schema.AnnotationsResult

	events, ranges := telemetryCalls(params.Content)
	for i, ev := range events {
		id := fmt.Sprintf("%s:%d", ev, i)
		item := &schema.OpenCodeGraphItem{
			Id:         id,
			Title:      "📊 Analytics: " + ev,
			Url:        "https://example.com/#not-yet-implemented",
			Preview:    true,
			PreviewUrl: "https://example.com/#not-yet-implemented",
		}

		result.Items = append(result.Items, item)
		result.Annotations = append(result.Annotations, &schema.OpenCodeGraphAnnotation{
			Item:  schema.OpenCodeGraphItemRef{Id: id},
			Range: ranges[i],
		})
	}

	return &result, nil
}

const anyQuote = "\"'`"

var logViewEventCall = regexp.MustCompile(`eventLogger.log(?:ViewEvent)?\([` + anyQuote + `]([^` + anyQuote + `]+)[` + anyQuote + `]`)

func telemetryCalls(content string) (events []string, ranges []schema.OpenCodeGraphRange) {
	return firstSubmatchNamesAndRanges(logViewEventCall, content)
}
