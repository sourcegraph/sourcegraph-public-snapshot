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
