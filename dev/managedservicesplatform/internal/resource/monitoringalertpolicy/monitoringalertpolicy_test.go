package monitoringalertpolicy

import (
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestBuildFilter(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config Config
		want   autogold.Value
	}{
		{
			name: "Service Metric",
			config: Config{
				ServiceName: "my-service-name",
				ServiceKind: "service",
				ThresholdAggregation: &ThresholdAggregation{
					Filters: map[string]string{
						"metric.type": "run.googleapis.com/container/startup_latencies",
					},
				},
			},
			want: autogold.Expect(`metric.type = "run.googleapis.com/container/startup_latencies" AND resource.type = "cloud_run_revision" AND resource.labels.service_name = "my-service-name"`),
		},
		{
			name: "Job Metric",
			config: Config{
				ServiceName: "my-job-name",
				ServiceKind: "job",
				ThresholdAggregation: &ThresholdAggregation{
					Filters: map[string]string{
						"metric.type":          "run.googleapis.com/job/completed_task_attempt_count",
						"metric.labels.result": "failed",
					},
				},
			},
			want: autogold.Expect(`metric.labels.result = "failed" AND metric.type = "run.googleapis.com/job/completed_task_attempt_count" AND resource.type = "cloud_run_job" AND resource.labels.job_name = "my-job-name"`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := buildFilter(&tc.config)
			tc.want.Equal(t, got)
		})
	}
}

func TestResponseCodeBuilder(t *testing.T) {
	for _, tc := range []struct {
		name string
		ResponseCodeMetric
		want autogold.Value
	}{
		{
			name: "Single Response Code",
			ResponseCodeMetric: ResponseCodeMetric{
				Code:  pointers.Ptr(404),
				Ratio: 0.1,
			},
			want: autogold.Expect(`fetch cloud_run_revision
| metric 'run.googleapis.com/request_count'
| group_by 15s, [value_request_count_aggregate: aggregate(value.request_count)]
| every 15s
| {
  group_by [metric.response_code],
  [response_code_count_aggregate: aggregate(value_request_count_aggregate)]
  | filter (metric.response_code = '404')
; group_by [],
  [value_request_count_aggregate_aggregate: aggregate(value_request_count_aggregate)]
}
| join
| value [response_code_ratio: val(0) / val(1)]
| condition gt(val(), 0.1)
`),
		},
		{
			name: "Response Code Class",
			ResponseCodeMetric: ResponseCodeMetric{
				CodeClass: pointers.Ptr("4xx"),
				Ratio:     0.4,
			},
			want: autogold.Expect(`fetch cloud_run_revision
| metric 'run.googleapis.com/request_count'
| group_by 15s, [value_request_count_aggregate: aggregate(value.request_count)]
| every 15s
| {
  group_by [metric.response_code, metric.response_code_class],
  [response_code_count_aggregate: aggregate(value_request_count_aggregate)]
  | filter (metric.response_code_class = '4xx')
; group_by [],
  [value_request_count_aggregate_aggregate: aggregate(value_request_count_aggregate)]
}
| join
| value [response_code_ratio: val(0) / val(1)]
| condition gt(val(), 0.4)
`),
		},
		{
			name: "Response Code Class + Exclude",
			ResponseCodeMetric: ResponseCodeMetric{
				CodeClass:    pointers.Ptr("4xx"),
				ExcludeCodes: []string{"404", "429"},
				Ratio:        0.8,
			},
			want: autogold.Expect(`fetch cloud_run_revision
| metric 'run.googleapis.com/request_count'
| group_by 15s, [value_request_count_aggregate: aggregate(value.request_count)]
| every 15s
| {
  group_by [metric.response_code, metric.response_code_class],
  [response_code_count_aggregate: aggregate(value_request_count_aggregate)]
  | filter (metric.response_code_class = '4xx')
  | filter (metric.response_code != '404')
  | filter (metric.response_code != '429')
; group_by [],
  [value_request_count_aggregate_aggregate: aggregate(value_request_count_aggregate)]
}
| join
| value [response_code_ratio: val(0) / val(1)]
| condition gt(val(), 0.8)
`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := responseCodeBuilder(&Config{
				ServiceName:        "test-service",
				ResponseCodeMetric: &tc.ResponseCodeMetric,
			})
			tc.want.Equal(t, got)
		})
	}
}
