package monitoring

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/monitoringalertpolicy"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack/options/googleprovider"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// Common
// - Container (8)
//    - run.googleapis.com/container/billable_instance_time
//    - run.googleapis.com/container/cpu/allocation_time
//    * run.googleapis.com/container/cpu/utilizations
//    - run.googleapis.com/container/memory/allocation_time
//    * run.googleapis.com/container/memory/utilizations
//    * run.googleapis.com/container/startup_latencies
//    - run.googleapis.com/container/network/received_bytes_count
//    - run.googleapis.com/container/network/sent_bytes_count
// - Log-based metrics (2)
//    - logging.googleapis.com/byte_count
//    - logging.googleapis.com/log_entry_count
// Cloud Run Job
// - Job (4)
//    - run.googleapis.com/job/completed_execution_count
//    * run.googleapis.com/job/completed_task_attempt_count
//    - run.googleapis.com/job/running_executions
//    - run.googleapis.com/job/running_task_attempts
// Cloud Run Service
// - Container (9)
//    - run.googleapis.com/container/completed_probe_attempt_count
//    - run.googleapis.com/container/completed_probe_count
//    - run.googleapis.com/container/probe_attempt_latencies
//    - run.googleapis.com/container/probe_latencies
//    * run.googleapis.com/container/instance_count
//    - run.googleapis.com/container/max_request_concurrencies
//    - run.googleapis.com/container/cpu/usage
//    - run.googleapis.com/container/containers
//    - run.googleapis.com/container/memory/usage
// - Request_count (1)
//    - run.googleapis.com/request_count
// - Request_latencies (1)
//    * run.googleapis.com/request_latencies
// - Pending_queue (1)
//    - run.googleapis.com/pending_queue/pending_requests

type CrossStackOutput struct{}

type Variables struct {
	ProjectID  string
	Service    spec.ServiceSpec
	Monitoring spec.MonitoringSpec
	MaxCount   *int
}

const StackName = "monitoring"

func NewStack(stacks *stack.Set, vars Variables) (*CrossStackOutput, error) {
	stack, _, err := stacks.New(StackName, googleprovider.With(vars.ProjectID))
	if err != nil {
		return nil, err
	}

	id := resourceid.New("monitoring")
	err = commonAlerts(stack, id, vars)
	if err != nil {
		return nil, err
	}

	switch pointers.Deref(vars.Service.Kind, spec.ServiceKindService) {
	case spec.ServiceKindService:
		err = serviceAlerts(stack, id, vars)
		if err != nil {
			return nil, err
		}

		if vars.Monitoring.Alerts.ResponseCodeRatios != nil {
			err = responseCodeMetrics(stack, id, vars)
		}

		if err != nil {
			return nil, err
		}
	case spec.ServiceKindJob:
		err = jobAlerts(stack, id, vars)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown service kind")
	}

	return &CrossStackOutput{}, nil
}

func commonAlerts(stack cdktf.TerraformStack, id resourceid.ID, vars Variables) error {
	for _, config := range []monitoringalertpolicy.Config{
		{
			ID:          "cpu",
			Name:        "High Container CPU Utilization",
			Description: pointers.Ptr("High CPU Usage - it may be neccessaru to reduce load or increase CPU allocation"),
			ThresholdAggregation: &monitoringalertpolicy.ThresholdAggregation{
				Filters:   map[string]string{"metric.type": "run.googleapis.com/container/cpu/utilizations"},
				Aligner:   monitoringalertpolicy.MonitoringAlignPercentile99,
				Reducer:   monitoringalertpolicy.MonitoringReduceMax,
				Period:    "300s",
				Threshold: 0.8,
			},
		},
		{
			ID:          "memory",
			Name:        "High Container Memory Utilization",
			Description: pointers.Ptr("High Memory Usage - it may be neccessary to reduce load or increase memory allocation"),
			ThresholdAggregation: &monitoringalertpolicy.ThresholdAggregation{
				Filters:   map[string]string{"metric.type": "run.googleapis.com/container/memory/utilizations"},
				Aligner:   monitoringalertpolicy.MonitoringAlignPercentile99,
				Reducer:   monitoringalertpolicy.MonitoringReduceMax,
				Period:    "300s",
				Threshold: 0.8,
			},
		},
		{
			ID:          "startup",
			Name:        "Container Startup Latency",
			Description: pointers.Ptr("Instance is taking a long time to start up - something may be blocking startup"),
			ThresholdAggregation: &monitoringalertpolicy.ThresholdAggregation{
				Filters:   map[string]string{"metric.type": "run.googleapis.com/container/startup_latencies"},
				Aligner:   monitoringalertpolicy.MonitoringAlignPercentile99,
				Reducer:   monitoringalertpolicy.MonitoringReduceMax,
				Period:    "60s",
				Threshold: 10000,
			},
		},
	} {

		config.ProjectID = vars.ProjectID
		config.ServiceName = vars.Service.ID
		config.ServiceKind = pointers.Deref(vars.Service.Kind, "service")
		_, err := monitoringalertpolicy.New(stack, id, &config)
		if err != nil {
			return err
		}
	}

	return nil
}

func serviceAlerts(stack cdktf.TerraformStack, id resourceid.ID, vars Variables) error {
	// Only provision if MaxCount is specified above 5
	if pointers.Deref(vars.MaxCount, 0) > 5 {
		_, err := monitoringalertpolicy.New(stack, id, &monitoringalertpolicy.Config{
			ID:          "instance_count",
			Name:        "Container Instance Count",
			Description: pointers.Ptr("There are a lot of Cloud Run instances running - we may need to increase per-instance requests make make sure we won't hit the configured max instance count"),
			ProjectID:   vars.ProjectID,
			ServiceName: vars.Service.ID,
			ServiceKind: spec.ServiceKindService,
			ThresholdAggregation: &monitoringalertpolicy.ThresholdAggregation{
				Filters: map[string]string{"metric.type": "run.googleapis.com/container/instance_count"},
				Aligner: monitoringalertpolicy.MonitoringAlignMax,
				Reducer: monitoringalertpolicy.MonitoringReduceMax,
				Period:  "60s",
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func jobAlerts(stack cdktf.TerraformStack, id resourceid.ID, vars Variables) error {
	// Alert whenever a Cloud Run Job fails
	_, err := monitoringalertpolicy.New(stack, id, &monitoringalertpolicy.Config{
		ID:          "job_failures",
		Name:        "Cloud Run Job Failures",
		Description: pointers.Ptr("Failed executions of Cloud Run Job"),
		ProjectID:   vars.ProjectID,
		ServiceName: vars.Service.ID,
		ServiceKind: spec.ServiceKindJob,
		ThresholdAggregation: &monitoringalertpolicy.ThresholdAggregation{
			Filters: map[string]string{
				"metric.type":          "run.googleapis.com/job/completed_task_attempt_count",
				"metric.labels.result": "failed",
			},
			GroupByField: "metric.label.result",
			Aligner:      monitoringalertpolicy.MonitoringAlignCount,
			Reducer:      monitoringalertpolicy.MonitoringReduceSum,
			Period:       "60s",
			Threshold:    0,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func responseCodeMetrics(stack cdktf.TerraformStack, id resourceid.ID, vars Variables) error {
	for _, config := range vars.Monitoring.Alerts.ResponseCodeRatios {

		_, err := monitoringalertpolicy.New(stack, id, &monitoringalertpolicy.Config{
			ID:          config.ID,
			ProjectID:   vars.ProjectID,
			Name:        config.Name,
			ServiceName: vars.Service.ID,
			ServiceKind: spec.ServiceKindService,
			ResponseCodeMetric: &monitoringalertpolicy.ResponseCodeMetric{
				Code:         config.Code,
				CodeClass:    config.CodeClass,
				ExcludeCodes: config.ExcludeCodes,
				Ratio:        config.Ratio,
				Duration:     config.Duration,
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
