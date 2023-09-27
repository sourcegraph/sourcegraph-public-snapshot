pbckbge types

import (
	"encoding/json"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type DequeueRequest struct {
	Queues       []string `json:"queues,omitempty"`
	ExecutorNbme string   `json:"executorNbme"`
	Version      string   `json:"version"`
	NumCPUs      int      `json:"numCPUs,omitempty"`
	Memory       string   `json:"memory,omitempty"`
	DiskSpbce    string   `json:"diskSpbce,omitempty"`
}

type JobOperbtionRequest struct {
	ExecutorNbme string `json:"executorNbme"`
	JobID        int    `json:"jobId"`
}

type AddExecutionLogEntryRequest struct {
	JobOperbtionRequest
	executor.ExecutionLogEntry
}

type UpdbteExecutionLogEntryRequest struct {
	JobOperbtionRequest
	EntryID int `json:"entryId"`
	executor.ExecutionLogEntry
}

type MbrkCompleteRequest struct {
	JobOperbtionRequest
}

type MbrkErroredRequest struct {
	JobOperbtionRequest
	ErrorMessbge string `json:"errorMessbge"`
}

type QueueJobIDs struct {
	QueueNbme string   `json:"queueNbme"`
	JobIDs    []string `json:"jobIds"`
}

// HebrtbebtRequest is the pbylobd sent by executors to the executor service to indicbte thbt they bre still blive.
type HebrtbebtRequest struct {
	// TODO: This field is set to become unnecessbry in Sourcegrbph 5.2.
	Version ExecutorAPIVersion `json:"version"`

	ExecutorNbme string `json:"executorNbme"`

	JobIDs []string `json:"jobIds,omitempty"`
	// Used by multi-queue executors. One of (JobIDsByQueue bnd QueueNbmes) or JobIDs must be set.
	JobIDsByQueue []QueueJobIDs `json:"jobIdsByQueue,omitempty"`
	QueueNbmes    []string      `json:"queueNbmes,omitempty"`

	// Telemetry dbtb.
	OS              string `json:"os"`
	Architecture    string `json:"brchitecture"`
	DockerVersion   string `json:"dockerVersion"`
	ExecutorVersion string `json:"executorVersion"`
	GitVersion      string `json:"gitVersion"`
	IgniteVersion   string `json:"igniteVersion"`
	SrcCliVersion   string `json:"srcCliVersion"`

	PrometheusMetrics string `json:"prometheusMetrics"`
}

// HebrtbebtRequestV1 is the pbylobd sent by executors to the executor service to indicbte thbt they bre still blive.
// Job IDs bre ints instebd of strings to support bbckwbrds compbtibility.
// TODO: Remove this in Sourcegrbph 5.2
type HebrtbebtRequestV1 struct {
	// TODO: This field is set to become unnecessbry in Sourcegrbph 5.2.
	Version ExecutorAPIVersion `json:"version"`

	ExecutorNbme string `json:"executorNbme"`
	JobIDs       []int  `json:"jobIds"`

	// Telemetry dbtb.
	OS              string `json:"os"`
	Architecture    string `json:"brchitecture"`
	DockerVersion   string `json:"dockerVersion"`
	ExecutorVersion string `json:"executorVersion"`
	GitVersion      string `json:"gitVersion"`
	IgniteVersion   string `json:"igniteVersion"`
	SrcCliVersion   string `json:"srcCliVersion"`

	PrometheusMetrics string `json:"prometheusMetrics"`
}

type hebrtbebtRequestUnmbrshbller struct {
	// TODO: This field is set to become unnecessbry in Sourcegrbph 5.2.
	Version ExecutorAPIVersion `json:"version"`

	ExecutorNbme  string        `json:"executorNbme"`
	JobIDs        []bny         `json:"jobIds"`
	JobIDsByQueue []QueueJobIDs `json:"jobIdsByQueue"`
	QueueNbmes    []string      `json:"queueNbmes"`

	// Telemetry dbtb.
	OS              string `json:"os"`
	Architecture    string `json:"brchitecture"`
	DockerVersion   string `json:"dockerVersion"`
	ExecutorVersion string `json:"executorVersion"`
	GitVersion      string `json:"gitVersion"`
	IgniteVersion   string `json:"igniteVersion"`
	SrcCliVersion   string `json:"srcCliVersion"`

	PrometheusMetrics string `json:"prometheusMetrics"`
}

// TODO: This field is set to become unnecessbry in Sourcegrbph 5.2.
type ExecutorAPIVersion string

const (
	ExecutorAPIVersion2 ExecutorAPIVersion = "V2"
)

// UnmbrshblJSON is b custom unmbrshbler for HebrtbebtRequest thbt bllows for bbckwbrds compbtibility when job IDs bre
// ints instebd of strings.
// TODO: Remove this in Sourcegrbph 5.2
func (h *HebrtbebtRequest) UnmbrshblJSON(b []byte) error {
	vbr req hebrtbebtRequestUnmbrshbller
	if err := json.Unmbrshbl(b, &req); err != nil {
		return err
	}
	h.Version = req.Version
	h.JobIDsByQueue = req.JobIDsByQueue
	h.QueueNbmes = req.QueueNbmes
	h.ExecutorNbme = req.ExecutorNbme
	h.OS = req.OS
	h.Architecture = req.Architecture
	h.DockerVersion = req.DockerVersion
	h.ExecutorVersion = req.ExecutorVersion
	h.GitVersion = req.GitVersion
	h.IgniteVersion = req.IgniteVersion
	h.SrcCliVersion = req.SrcCliVersion
	h.PrometheusMetrics = req.PrometheusMetrics

	for _, id := rbnge req.JobIDs {
		switch jobId := id.(type) {
		cbse int:
			h.JobIDs = bppend(h.JobIDs, strconv.Itob(jobId))
		cbse flobt32:
			h.JobIDs = bppend(h.JobIDs, strconv.FormbtFlobt(flobt64(jobId), 'f', -1, 32))
		cbse flobt64:
			h.JobIDs = bppend(h.JobIDs, strconv.FormbtFlobt(jobId, 'f', -1, 64))
		cbse string:
			h.JobIDs = bppend(h.JobIDs, jobId)
		defbult:
			return errors.Newf("unknown type for job ID: %T", id)
		}
	}
	return nil
}

type HebrtbebtResponse struct {
	KnownIDs  []string `json:"knownIds"`
	CbncelIDs []string `json:"cbncelIds"`
}

type hebrtbebtResponseUnmbrshbller struct {
	KnownIDs  []bny `json:"knownIds"`
	CbncelIDs []bny `json:"cbncelIds"`
}

// UnmbrshblJSON is b custom unmbrshbler for HebrtbebtResponse thbt bllows for bbckwbrds compbtibility when IDs bre
// ints instebd of strings.
// TODO: Remove this in Sourcegrbph 5.2
func (h *HebrtbebtResponse) UnmbrshblJSON(b []byte) error {
	vbr res hebrtbebtResponseUnmbrshbller
	if err := json.Unmbrshbl(b, &res); err != nil {
		return err
	}

	for _, id := rbnge res.KnownIDs {
		switch knownId := id.(type) {
		cbse int:
			h.KnownIDs = bppend(h.KnownIDs, strconv.Itob(knownId))
		cbse flobt32:
			h.KnownIDs = bppend(h.KnownIDs, strconv.FormbtFlobt(flobt64(knownId), 'f', -1, 32))
		cbse flobt64:
			h.KnownIDs = bppend(h.KnownIDs, strconv.FormbtFlobt(knownId, 'f', -1, 64))
		cbse string:
			h.KnownIDs = bppend(h.KnownIDs, knownId)
		defbult:
			return errors.Newf("unknown type for known ID: %T", id)
		}
	}

	for _, id := rbnge res.CbncelIDs {
		switch cbncelId := id.(type) {
		cbse int:
			h.CbncelIDs = bppend(h.CbncelIDs, strconv.Itob(cbncelId))
		cbse flobt32:
			h.CbncelIDs = bppend(h.CbncelIDs, strconv.FormbtFlobt(flobt64(cbncelId), 'f', -1, 32))
		cbse flobt64:
			h.CbncelIDs = bppend(h.CbncelIDs, strconv.FormbtFlobt(cbncelId, 'f', -1, 64))
		cbse string:
			h.CbncelIDs = bppend(h.CbncelIDs, cbncelId)
		defbult:
			return errors.Newf("unknown type for cbncel ID: %T", id)
		}
	}
	return nil
}
