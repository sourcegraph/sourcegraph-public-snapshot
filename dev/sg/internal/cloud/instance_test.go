package cloud

import (
	"strconv"
	"strings"
	"testing"
	"time"

	cloudapiv1 "github.com/sourcegraph/cloud-api/go/cloudapi/v1"
)

func TestInstanceStatus(t *testing.T) {
	tt := []struct {
		name       string
		actionURL  string
		statusEnum cloudapiv1.InstanceStatus
		statusText string
		reason     string
		errText    string
	}{
		{
			name:       "failed with actionURL and status",
			actionURL:  "http://test.com/action/123",
			statusEnum: cloudapiv1.InstanceStatus_INSTANCE_STATUS_FAILED,
			statusText: "failed",
			reason:     "url:http://test.com/action/123, status: failed",
		},
		{
			name:       "completed with no actionURL and no status",
			actionURL:  "",
			statusEnum: cloudapiv1.InstanceStatus_INSTANCE_STATUS_OK,
			statusText: "completed",
			reason:     "",
		},
		{
			name:    "incorrect reason format",
			reason:  "https://test.com/action/123",
			errText: "invalid status reason format",
		},
		{
			name:    "incorrect reason field format",
			reason:  "actionURL=https://test.com/action/123, status=completed",
			errText: "field error",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			src := &cloudapiv1.InstanceState{
				InstanceStatus: tc.statusEnum,
				Reason:         &tc.reason,
			}

			instanceSatus, err := newInstanceStatus(src)
			if err != nil {
				if tc.errText == "" {
					t.Fatal(err)
				} else if !strings.Contains(err.Error(), tc.errText) {
					t.Errorf("incorrect error. want=%s have=%s", tc.errText, err.Error())
				}
				return
			}

			if instanceSatus.ActionURL != tc.actionURL {
				t.Errorf("incorrect action url. want=%s have=%s", tc.actionURL, instanceSatus.ActionURL)
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
