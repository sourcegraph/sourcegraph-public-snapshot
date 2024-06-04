package cloud

import (
	"strconv"
	"testing"
	"time"

	cloudapiv1 "github.com/sourcegraph/cloud-api/go/cloudapi/v1"
)

func TestInstanceStatus(t *testing.T) {
	tt := []struct {
		name       string
		jobURL     string
		jobState   string
		step       string
		phase      string
		statusEnum cloudapiv1.InstanceStatus
		statusText string
		reason     string
		errText    string
	}{
		{
			name:       "in_progress with jobURL and status",
			jobURL:     "https://github.com/sourcegraph/cloud/actions/runs/9209264595",
			jobState:   "in_progress",
			step:       "1/3",
			phase:      "creating instance",
			statusEnum: cloudapiv1.InstanceStatus_INSTANCE_STATUS_PROGRESSING,
			statusText: "in-progress",
			reason:     "step 1/3:creating instance, job-url:https://github.com/sourcegraph/cloud/actions/runs/9209264595, state:in_progress",
		},
		{
			name:       "completed with no actionURL and no status",
			jobURL:     "",
			statusEnum: cloudapiv1.InstanceStatus_INSTANCE_STATUS_OK,
			statusText: "completed",
			reason:     "",
		},
		{
			name:       "incorrect reason format",
			statusEnum: cloudapiv1.InstanceStatus_INSTANCE_STATUS_UNSPECIFIED,
			statusText: "unspecified",
			reason:     "https://test.com/action/123",
			errText:    `failed to parse status reason: "https://test.com/action/123"`,
		},
		{
			name:       "incorrect reason field format",
			statusEnum: cloudapiv1.InstanceStatus_INSTANCE_STATUS_UNSPECIFIED,
			statusText: "unspecified",
			reason:     "actionURL=https://test.com/action/123, status=completed",
			errText:    `failed to parse status reason: "actionURL=https://test.com/action/123, status=completed"`,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			src := &cloudapiv1.InstanceState{
				InstanceStatus: tc.statusEnum,
				Reason:         &tc.reason,
			}

			instanceSatus := newInstanceStatus(src)
			if tc.errText != "" && instanceSatus.Error != tc.errText {
				t.Errorf("incorrect error. want=%s have=%s", tc.errText, instanceSatus.Error)
			}

			if instanceSatus.Reason.JobURL != tc.jobURL {
				t.Errorf("incorrect action url. want=%s have=%s", tc.jobURL, instanceSatus.Reason.JobURL)
			}
			if instanceSatus.Reason.JobState != tc.jobState {
				t.Errorf("incorrect action url. want=%s have=%s", tc.jobState, instanceSatus.Reason.JobState)
			}
			if instanceSatus.Reason.Step != tc.step {
				t.Errorf("incorrect reason step. want=%s have=%s", tc.step, instanceSatus.Reason.Step)
			}
			if instanceSatus.Reason.Phase != tc.phase {
				t.Errorf("incorrect action url. want=%s have=%s", tc.phase, instanceSatus.Reason.Phase)
			}
			if instanceSatus.Status != tc.statusText {
				t.Errorf("incorrect status. want=%s have=%s", tc.statusText, instanceSatus.Status)
			}
		})
	}
}

func TestInstanceFeatures(t *testing.T) {
	now := time.Now()
	features := newInstanceFeaturesFrom(map[string]string{
		"ephemeral_instance":            "true",
		"ephemeral_instance_lease_time": strconv.FormatInt(now.Unix(), 10),
	})

	if !features.IsEphemeralInstance() {
		t.Errorf("expected ephemeral instance to be true")
	}
	lease, err := features.GetEphemeralLeaseTime()
	if err != nil {
		t.Fatalf("failed to instance lease time: %v", err)
	}
	if lease.Unix() != now.Unix() {
		t.Errorf("expected lease to be %d, got %d", now.Unix(), lease.Unix())
	}
}
