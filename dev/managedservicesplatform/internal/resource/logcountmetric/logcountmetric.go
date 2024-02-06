package logcountmetric

import (
	"fmt"
	"sort"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/loggingmetric"
	"golang.org/x/exp/maps"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type LabelExtractor struct {
	// Expression which is used to extract data from a log entry field and assign
	// as the label value. Value can be just 'EXTRACT(field)' or
	// 'REGEXP_EXTRACT(field, regex)' to extract a regex match from the target
	// field.
	Expression string
	// Type of data that can be assigned to the label. Default value: "STRING"
	// Possible values: ["BOOL", "INT64", "STRING"].
	Type string
	// Human-readable description.
	Description string
}

type Config struct {
	// Name of the metric with '/' namespace, e.g. msp.sourcegraph.com/my_metric
	Name string

	// LogFilters for matching log messages - see https://cloud.google.com/logging/docs/view/logging-query-language
	LogFilters string

	// A map from a label key string to an extractor
	LabelExtractors map[string]LabelExtractor
}

type Output struct {
	// Metric can be used for the 'metric.type' filter.
	Metric string
}

// New creates a logging metric that looks specifically for counts of a matching entry.
func New(scope constructs.Construct, id resourceid.ID, config *Config) (*Output, error) {
	extractors, descriptors := buildExtractorsAndDescriptors(config.LabelExtractors)

	metric := loggingmetric.NewLoggingMetric(scope, id.TerraformID("metric"), &loggingmetric.LoggingMetricConfig{
		Name:            &config.Name,
		Filter:          pointers.Ptr(config.LogFilters),
		LabelExtractors: extractors,
		MetricDescriptor: &loggingmetric.LoggingMetricMetricDescriptor{
			MetricKind: pointers.Ptr("DELTA"),
			ValueType:  pointers.Ptr("INT64"),
			Labels:     descriptors,
		},
	})
	return &Output{
		Metric: fmt.Sprintf("logging.googleapis.com/user/%s", *metric.Id()),
	}, nil
}

func buildExtractorsAndDescriptors(extractors map[string]LabelExtractor) (*map[string]*string, *[]loggingmetric.LoggingMetricMetricDescriptorLabels) {
	labels := maps.Keys(extractors)
	sort.Strings(labels)

	labelExtractors := map[string]*string{}
	labelDescriptors := []loggingmetric.LoggingMetricMetricDescriptorLabels{}
	for _, l := range labels {
		extractor := extractors[l]
		labelExtractors[l] = pointers.Ptr(extractor.Expression)
		labelDescriptors = append(labelDescriptors, loggingmetric.LoggingMetricMetricDescriptorLabels{
			Key:         pointers.Ptr(l),
			ValueType:   pointers.Ptr(extractor.Type),
			Description: pointers.Ptr(extractor.Description),
		})
	}
	return &labelExtractors, &labelDescriptors
}
