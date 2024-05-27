package cloud

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/regexp"
	cloudapiv1 "github.com/sourcegraph/cloud-api/go/cloudapi/v1"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

var ErrLeaseTimeNotSet error = errors.New("lease time not set")

const (
	// EphemeralInstanceType is the instance type for ephemeral instances. An instance is considered ephemeral if it
	// contains "ephemeral_instance": "true" in its Instance Features
	EphemeralInstanceType = "ephemeral"

	// InternalInstanceType is the instance type for internal instances. An instance is considered internal if it it is
	// in the Dev cloud environment and does not contain "ephemeral_instance": "true" in its Instance Features
	InternalInstanceType = "internal"

	InstanceStatusUnspecified = "unspecified"
	InstanceStatusCompleted   = "completed"
	InstanceStatusInProgress  = "in-progress"
	InstanceStatusFailed      = "failed"
	InstanceStatusUnknown     = "unknown"
)

type Instance struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	InstanceType string `json:"instanceType"`
	Environment  string `json:"environment"`
	Version      string `json:"version"`
	URL          string `json:"hostname"`
	AdminEmail   string `json:"adminEmail"`

	CreatedAt time.Time `json:"createdAt"`
	DeletedAt time.Time `json:"deletedAt"`
	ExpiresAt time.Time `json:"ExpiresAt"`

	Project string         `json:"project"`
	Region  string         `json:"region"`
	Status  InstanceStatus `json:"status"`
	// contains various key value pairs that are specific to the instance type
	features *InstanceFeatures
}

func (i *Instance) String() string {
	// Protobuf returns the unix zero epoch if the time is nil, so we check for that
	// and we also check if we do have a valid time that it is not zero
	fmtTime := func(t time.Time) string {
		if isUnixEpochZero(t) || t.IsZero() {
			return "n/a"
		}
		return t.Format(time.RFC3339)
	}
	return fmt.Sprintf(`ID           : %s
Name         : %s
InstanceType : %s
Environment  : %s
Version      : %s
URL          : %s
AdminEmail   : %s
CreatedAt    : %s
DeletetAt    : %s
ExpiresAt    : %s
Project      : %s
Region       : %s
%s
`, i.ID, i.Name, i.InstanceType, i.Environment, i.Version, i.URL, i.AdminEmail,
		fmtTime(i.CreatedAt), fmtTime(i.DeletedAt), fmtTime(i.ExpiresAt), i.Project, i.Region,
		i.Status.String())
}

func (i *Instance) IsEphemeral() bool {
	return i.InstanceType == EphemeralInstanceType
}

func (i *Instance) IsInternal() bool {
	return i.InstanceType == InternalInstanceType
}

func (i *Instance) IsExpired() bool {
	if i.ExpiresAt.IsZero() {
		return false
	}

	return time.Now().After(i.ExpiresAt)
}

func (i *Instance) HasStatus(status string) bool {
	return i.Status.Status == status
}

type InstanceStatus struct {
	Status string       `json:"status"`
	Reason StatusReason `json:"reason"`
	Error  string       `json:"error"`
}

type StatusReason struct {
	Step     string `json:"step"`
	Phase    string `json:"phase"`
	JobCount int    `json:"job_count"`
	JobURL   string `json:"job_url"`
	JobState string `json:"job_state"`
	Overall  string `json:"overall"`
}

func newStatusReason(reason string) (StatusReason, error) {
	if reason == "" {
		return StatusReason{}, nil
	}

	// TODO(burmudar): handle storing of multiple jobs
	jobCount := 1
	// if the reason contains a semicolon it means there are multiple jobs, we only want the last job
	if strings.Contains(reason, ";") {
		parts := strings.Split(reason, ";")
		// we want to know the number of jobs
		jobCount = len(parts)
		// we want to know the last job
		reason = strings.TrimSpace(parts[jobCount-1])
	}
	// step 1/3:creating instance, job-url:https://github.com/sourcegraph/cloud/actions/runs/9209264595, state:in_progress, conclusion: failed
	statusRegex := regexp.MustCompile(`^step (\d\/\d):(.*), job-url:(.*), state:(\w+)(, conclusion:(\w+))?$`)
	matches := statusRegex.FindStringSubmatch(reason)
	// if matches is 7, it means we have a conclusion, if matches is 5 then we don't have aa conclusion
	if len(matches) != 5 && len(matches) != 7 {
		return StatusReason{}, errors.Newf("failed to parse status reason: %q", reason)
	}
	var conclusion string
	if len(matches) == 7 {
		conclusion = matches[6]
	}
	return StatusReason{
		Step:     matches[1],
		Phase:    matches[2],
		JobCount: jobCount,
		JobURL:   matches[3],
		JobState: matches[4],
		Overall:  conclusion,
	}, nil
}

func (s *StatusReason) GetStepPhaseString() string {
	if s.Step == "" || s.Phase == "" {
		return ""
	}

	return fmt.Sprintf("(%s %s)", s.Step, s.Phase)
}

func (s *StatusReason) GetJobURLStateString() string {
	value := s.JobURL
	if s.JobState != "" {
		value = fmt.Sprintf("%s (%s)", value, s.JobState)
	}
	return value
}

func (s *InstanceStatus) String() string {
	return fmt.Sprintf(`Status       : %s
Step         : %s
Phase        : %s
JobURL       : %s
JobState     : %s
Overall      : %s`, s.Status, s.Reason.Step, s.Reason.Phase, s.Reason.JobURL, s.Reason.JobState, s.Reason.Overall)
}

type InstanceFeatures struct {
	features map[string]string
}

func newInstanceStatus(src *cloudapiv1.InstanceState) (*InstanceStatus, error) {
	reason, err := newStatusReason(src.GetReason())
	if err != nil {
		return nil, err
	}

	status := InstanceStatus{
		Reason: reason,
	}
	switch src.GetInstanceStatus() {
	case cloudapiv1.InstanceStatus_INSTANCE_STATUS_UNSPECIFIED:
		status.Status = InstanceStatusUnspecified
	case cloudapiv1.InstanceStatus_INSTANCE_STATUS_OK:
		status.Status = InstanceStatusCompleted
	case cloudapiv1.InstanceStatus_INSTANCE_STATUS_PROGRESSING:
		status.Status = InstanceStatusInProgress
	case cloudapiv1.InstanceStatus_INSTANCE_STATUS_FAILED:
		status.Status = InstanceStatusFailed
		status.Error = src.GetReason()
	default:
		status.Status = InstanceStatusUnknown
	}

	return &status, nil
}

func newInstance(src *cloudapiv1.Instance) (*Instance, error) {
	details := src.GetInstanceDetails()
	platform := src.GetPlatformDetails()
	status, err := newInstanceStatus(src.GetInstanceState())
	if err != nil {
		return nil, err
	}
	features := newInstanceFeaturesFrom(details.GetInstanceFeatures())
	expiresAt, err := features.GetEphemeralLeaseTime()
	if err != nil && !errors.Is(err, ErrLeaseTimeNotSet) {
		return nil, err
	}

	instanceType := InternalInstanceType
	if features.IsEphemeralInstance() {
		instanceType = EphemeralInstanceType
	}

	return &Instance{
		ID:           src.GetId(),
		Name:         details.Name,
		Environment:  DevEnvironment,
		InstanceType: instanceType,
		Version:      details.Version,
		URL:          pointers.DerefZero(details.Url),
		AdminEmail:   pointers.DerefZero(details.AdminEmail),
		CreatedAt:    platform.GetCreatedAt().AsTime(),
		DeletedAt:    platform.GetDeletedAt().AsTime(),
		ExpiresAt:    expiresAt,
		Project:      platform.GetGcpProjectId(),
		Region:       platform.GetGcpRegion(),
		Status:       *status,
		features:     features,
	}, nil
}

func isUnixEpochZero(t time.Time) bool {
	return t.Unix() == 0
}

func toInstances(items ...*cloudapiv1.Instance) ([]*Instance, error) {
	converted := []*Instance{}
	for _, item := range items {
		inst, err := newInstance(item)
		if err != nil {
			return nil, err
		}
		converted = append(converted, inst)
	}
	return converted, nil
}

func newInstanceFeaturesFrom(src map[string]string) *InstanceFeatures {
	return &InstanceFeatures{
		features: src,
	}
}
func newInstanceFeatures() *InstanceFeatures {
	return &InstanceFeatures{features: make(map[string]string)}
}

func (f *InstanceFeatures) IsEphemeralInstance() bool {
	v, ok := f.features["ephemeral_instance"]
	if !ok {
		return false
	}
	val, err := strconv.ParseBool(v)
	if err != nil {
		return false
	}

	return val
}

func (f *InstanceFeatures) SetEphemeralInstance(v bool) {
	f.features["ephemeral"] = strconv.FormatBool(v)
}

func (f *InstanceFeatures) SetEphemeralLeaseTime(expiresAt time.Time) {
	f.features["ephemeral_instance_lease_time"] = strconv.FormatInt(expiresAt.Unix(), 10)
}

func (f *InstanceFeatures) GetEphemeralLeaseTime() (time.Time, error) {
	seconds, ok := f.features["ephemeral_instance_lease_time"]
	if !ok {
		return time.Time{}, ErrLeaseTimeNotSet
	}
	secondsInt, err := strconv.ParseInt(seconds, 10, 64)
	if err != nil {
		return time.Time{}, errors.Newf("failed to convert 'ephemeral_instance_lease_time' value %q to int64: %v", seconds, err)
	}
	leaseTime := time.Unix(secondsInt, 0)
	return leaseTime, nil
}

func (f *InstanceFeatures) Value() map[string]string {
	return f.features
}
