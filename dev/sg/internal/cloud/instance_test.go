package cloud

import (
	"testing"

	cloudapiv1 "github.com/sourcegraph/cloud-api/go/cloudapi/v1"
)

func TestInstanceStatus(t *testing.T) {
	tt := []struct {
		name      string
		actionURL string
		status    string
		reason    string
		errText   string
	}{
		{
			name:      "failed with actionURL and status",
			actionURL: "http://test.com/action/123",
			status:    "failed",
			reason:    "url:http://test.com/action/123, status: failed",
		},
		{
			name:      "completed with no actionURL and no status",
			actionURL: "",
			status:    "",
			reason:    "",
		},
		{
			name:    "incorrect reason format",
			reason:  "https://test.com/action/123",
			errText: "invalid satus reason format",
		},
		{
			name:    "incorrect reason field format",
			reason:  "actionURL=https://test.com/action/123, status=completed",
			errText: "invalid field value",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			src := &cloudapiv1.InstanceState{
				InstanceStatus: cloudapiv1.InstanceStatus_INSTANCE_STATUS_FAILED,
				Reason:         &tc.reason,
			}

			instanceSatus, err := newInstanceStatus(src)
			if err != nil {
				if tc.errText == "" || !strings.Contains(err.Error(), tc.errText) {
					t.Errorf("incorrect error. want=%s have=%s", tc.errText, err.Error())
				} else {
					t.Fatal(err)
				}
			}

			if instanceSatus.ActionURL != tc.actionURL {
				t.Errorf("incorrect action url. want=%s have=%s", tc.actionURL, instanceSatus.ActionURL)
			}
			if instanceSatus.Status != tc.status {
				t.Errorf("incorrect status. want=%s have=%s", tc.status, instanceSatus.Status)
			}
		})
	}
}
