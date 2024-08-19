package gcplogurl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/compute/metadata"
	"go.opencensus.io/trace"
)

// TraceLogURL construct Cloud Logging URL about this trace (request).
func TraceLogURL(ctx context.Context) (string, error) {
	projectID, _ := metadata.ProjectID()
	if projectID == "" {
		for _, key := range []string{"GCP_PROJECT", "GCLOUD_PROJECT", "GOOGLE_CLOUD_PROJECT"} {
			projectID = os.Getenv(key)
			if projectID != "" {
				break
			}
		}
	}
	if projectID == "" {
		return "", errors.New("failed to detect GCP Project ID")
	}
	span := trace.FromContext(ctx)
	if span == nil {
		return "", errors.New("ctx doesn't have OpenCensus span")
	}
	ex := &Explorer{
		ProjectID: projectID,
		Query:     Query(fmt.Sprintf(`trace="projects/%s/traces/%s"`, projectID, span.SpanContext().TraceID.String())),
		TimeRange: &SpecificTimeWithRange{
			At:    time.Now(),
			Range: 2 * time.Hour,
		},
	}
	return ex.String(), nil
}
