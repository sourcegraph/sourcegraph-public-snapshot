package alertpolicy

import (
	"testing"

	"github.com/hexops/autogold/v2"
)

func TestBuildThresholdAggregationFilter(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config Config
		want   autogold.Value
	}{
		{
			name: "Service Metric",
			config: Config{
				ThresholdAggregation: &ThresholdAggregation{
					ResourceName: "my-service-name",
					ResourceKind: CloudRunService,
					Filters: map[string]string{
						"metric.type": "run.googleapis.com/container/startup_latencies",
					},
				},
			},
			want: autogold.Expect(`metric.type = "run.googleapis.com/container/startup_latencies" AND resource.type = "cloud_run_revision" AND resource.labels.service_name = starts_with("my-service-name")`),
		},
		{
			name: "Job Metric",
			config: Config{
				ThresholdAggregation: &ThresholdAggregation{
					ResourceName: "my-job-name",
					ResourceKind: CloudRunJob,
					Filters: map[string]string{
						"metric.type":          "run.googleapis.com/job/completed_task_attempt_count",
						"metric.labels.result": "failed",
					},
				},
			},
			want: autogold.Expect(`metric.labels.result = "failed" AND metric.type = "run.googleapis.com/job/completed_task_attempt_count" AND resource.type = "cloud_run_job" AND resource.labels.job_name = starts_with("my-job-name")`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := buildThresholdAggregationFilter(&tc.config)
			tc.want.Equal(t, got)
		})
	}
}
