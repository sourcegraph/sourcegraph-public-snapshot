package instancehealth

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

// run with -v for output
func TestCheckPermissionsSyncing(t *testing.T) {
	for _, tt := range []struct {
		name           string
		instanceHealth Indicators

		wantEmojis []string
		wantErr    string
	}{{
		name: "no jobs",
		instanceHealth: Indicators{
			PermissionsSyncJobs: struct{ Nodes []permissionSyncJob }{
				Nodes: nil,
			},
		},
		wantEmojis: []string{output.EmojiWarning},
		wantErr:    "",
	}, {
		name: "healthy",
		instanceHealth: Indicators{
			PermissionsSyncJobs: struct{ Nodes []permissionSyncJob }{
				Nodes: []permissionSyncJob{{
					FinishedAt: time.Now(),
					State:      "SUCCESS",
					CodeHostStates: []permissionsProviderStatus{{
						ProviderType: "github",
						ProviderID:   "https://github.com/",
						Status:       "SUCCESS",
					}},
				}},
			},
		},
		wantEmojis: []string{output.EmojiSuccess},
		wantErr:    "",
	}, {
		name: "unhealthy",
		instanceHealth: Indicators{
			PermissionsSyncJobs: struct{ Nodes []permissionSyncJob }{
				Nodes: []permissionSyncJob{{
					FinishedAt:     time.Now(),
					State:          "ERROR",
					FailureMessage: "oh no!",
					CodeHostStates: []permissionsProviderStatus{{
						ProviderType: "github",
						ProviderID:   "https://github.com/",
						Status:       "ERROR",
					}},
				}},
			},
		},
		wantEmojis: []string{output.EmojiFailure},
		wantErr:    "permissions sync errors",
	}} {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := checkPermissionsSyncing(output.NewOutput(io.MultiWriter(os.Stderr, &out), output.OutputOpts{}), time.Hour, tt.instanceHealth)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
			if len(tt.wantEmojis) > 0 {
				data := out.String()
				for _, emoji := range tt.wantEmojis {
					assert.Contains(t, data, emoji)
				}
			}
		})
	}
}
