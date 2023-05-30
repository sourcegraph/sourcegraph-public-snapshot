package types

import (
	"encoding/json"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type DequeueRequest struct {
	Queues       []string `json:"queues,omitempty"`
	ExecutorName string   `json:"executorName"`
	Version      string   `json:"version"`
	NumCPUs      int      `json:"numCPUs,omitempty"`
	Memory       string   `json:"memory,omitempty"`
	DiskSpace    string   `json:"diskSpace,omitempty"`
}

type JobOperationRequest struct {
	ExecutorName string `json:"executorName"`
	JobID        int    `json:"jobId"`
}

type AddExecutionLogEntryRequest struct {
	JobOperationRequest
	executor.ExecutionLogEntry
}

type UpdateExecutionLogEntryRequest struct {
	JobOperationRequest
	EntryID int `json:"entryId"`
	executor.ExecutionLogEntry
}

type MarkCompleteRequest struct {
	JobOperationRequest
}

type MarkErroredRequest struct {
	JobOperationRequest
	ErrorMessage string `json:"errorMessage"`
}

// HeartbeatRequest is the payload sent by executors to the executor service to indicate that they are still alive.
type HeartbeatRequest struct {
	ExecutorName string   `json:"executorName"`
	JobIDs       []string `json:"jobIds"`

	// Telemetry data.
	OS              string `json:"os"`
	Architecture    string `json:"architecture"`
	DockerVersion   string `json:"dockerVersion"`
	ExecutorVersion string `json:"executorVersion"`
	GitVersion      string `json:"gitVersion"`
	IgniteVersion   string `json:"igniteVersion"`
	SrcCliVersion   string `json:"srcCliVersion"`

	PrometheusMetrics string `json:"prometheusMetrics"`
}

// HeartbeatRequestV1 is the payload sent by executors to the executor service to indicate that they are still alive.
// Job IDs are ints instead of strings to support backwards compatibility.
// TODO: Remove this in Sourcegraph 5.2
type HeartbeatRequestV1 struct {
	ExecutorName string `json:"executorName"`
	JobIDs       []int  `json:"jobIds"`

	// Telemetry data.
	OS              string `json:"os"`
	Architecture    string `json:"architecture"`
	DockerVersion   string `json:"dockerVersion"`
	ExecutorVersion string `json:"executorVersion"`
	GitVersion      string `json:"gitVersion"`
	IgniteVersion   string `json:"igniteVersion"`
	SrcCliVersion   string `json:"srcCliVersion"`

	PrometheusMetrics string `json:"prometheusMetrics"`
}

type heartbeatRequestUnmarshaller struct {
	ExecutorName string `json:"executorName"`
	JobIDs       []any  `json:"jobIds"`

	// Telemetry data.
	OS              string `json:"os"`
	Architecture    string `json:"architecture"`
	DockerVersion   string `json:"dockerVersion"`
	ExecutorVersion string `json:"executorVersion"`
	GitVersion      string `json:"gitVersion"`
	IgniteVersion   string `json:"igniteVersion"`
	SrcCliVersion   string `json:"srcCliVersion"`

	PrometheusMetrics string `json:"prometheusMetrics"`
}

// UnmarshalJSON is a custom unmarshaler for HeartbeatRequest that allows for backwards compatibility when job IDs are
// ints instead of strings.
// TODO: Remove this in Sourcegraph 5.2
func (h *HeartbeatRequest) UnmarshalJSON(b []byte) error {
	var req heartbeatRequestUnmarshaller
	if err := json.Unmarshal(b, &req); err != nil {
		return err
	}
	h.ExecutorName = req.ExecutorName
	h.OS = req.OS
	h.Architecture = req.Architecture
	h.DockerVersion = req.DockerVersion
	h.ExecutorVersion = req.ExecutorVersion
	h.GitVersion = req.GitVersion
	h.IgniteVersion = req.IgniteVersion
	h.SrcCliVersion = req.SrcCliVersion
	h.PrometheusMetrics = req.PrometheusMetrics

	for _, id := range req.JobIDs {
		switch jobId := id.(type) {
		case int:
			h.JobIDs = append(h.JobIDs, strconv.Itoa(jobId))
		case float32:
			h.JobIDs = append(h.JobIDs, strconv.FormatFloat(float64(jobId), 'f', -1, 32))
		case float64:
			h.JobIDs = append(h.JobIDs, strconv.FormatFloat(jobId, 'f', -1, 64))
		case string:
			h.JobIDs = append(h.JobIDs, jobId)
		default:
			return errors.Newf("unknown type for job ID: %T", id)
		}
	}
	return nil
}

type HeartbeatResponse struct {
	KnownIDs  []string `json:"knownIds"`
	CancelIDs []string `json:"cancelIds"`
}

// TODO: Deprecated. Can be removed in Sourcegraph 4.4.
type CanceledJobsRequest struct {
	KnownJobIDs  []string `json:"knownJobIds"`
	ExecutorName string   `json:"executorName"`
}
