package telemetrygateway_test

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/lib/pointers"

	"github.com/sourcegraph/sourcegraph/internal/trace/tracetest"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

var (
	createSnapshot = flag.Bool("snapshot", false, "generate a new snapshot")
	snapshotsDir   = filepath.Join("testdata", "snapshots")

	// sampleEvent should have all fields populated - a snapshot from this event
	// is generated with the -snapshot flag, see TestBackcompat() for more
	// details.
	sampleEvent = &telemetrygatewayv1.Event{
		Id:        "1234",
		Feature:   "Feature",
		Action:    "Action",
		Timestamp: timestamppb.New(must(time.Parse(time.RFC3339, "2023-02-24T14:48:30Z"))),
		Interaction: &telemetrygatewayv1.EventInteraction{
			TraceId: func() *string {
				tid, _ := (tracetest.StaticTraceIDGenerator{}).NewIDs(context.Background())
				return pointers.Ptr(tid.String())
			}(),
			Geolocation: &telemetrygatewayv1.EventInteraction_Geolocation{
				CountryCode: "US",
			},
		},
		Source: &telemetrygatewayv1.EventSource{
			Server: &telemetrygatewayv1.EventSource_Server{
				Version: "dev",
			},
			Client: &telemetrygatewayv1.EventSource_Client{
				Name:    "CLIENT",
				Version: pointers.Ptr("VERSION"),
			},
		},
		Parameters: &telemetrygatewayv1.EventParameters{
			Version: 1,
			Metadata: map[string]float64{
				"metadata": 1,
				"float":    1.3,
			},
			PrivateMetadata: must(structpb.NewStruct(map[string]any{"private": "data"})),
			BillingMetadata: &telemetrygatewayv1.EventBillingMetadata{
				Product:  "Product",
				Category: "Category",
			},
		},
		User: &telemetrygatewayv1.EventUser{
			UserId:          pointers.Ptr(int64(1234)),
			AnonymousUserId: pointers.Ptr("anonymous"),
		},
		FeatureFlags: &telemetrygatewayv1.EventFeatureFlags{
			Flags: map[string]string{"feature": "true"},
		},
		MarketingTracking: &telemetrygatewayv1.EventMarketingTracking{
			Url:             pointers.Ptr("value"),
			FirstSourceUrl:  pointers.Ptr("value"),
			CohortId:        pointers.Ptr("value"),
			Referrer:        pointers.Ptr("value"),
			LastSourceUrl:   pointers.Ptr("value"),
			DeviceSessionId: pointers.Ptr("value"),
			SessionReferrer: pointers.Ptr("value"),
			SessionFirstUrl: pointers.Ptr("value"),
		},
	}
)

// TestBackcompat asserts that past events marshalled in the proto wire format,
// tracked in internal/telemetrygateway/v1/testdata/snapshots, continue to be
// able to be marshalled by the current v1 types to ensure we don't introduce
// any breaking changes.
//
// New snapshots should be manually created as the spec evolves by updating
// sampleEvent and running the test with the '-snapshot' flag:
//
//	go test -v ./internal/telemetrygateway/v1 -snapshot
//
// Without the '-snapshot' flag, this test just loads existing snapshots and
// asserts they can still be unmarshalled.
func TestBackcompat(t *testing.T) {
	if *createSnapshot {
		data, err := proto.Marshal(sampleEvent)
		require.NoError(t, err)

		f := filepath.Join(snapshotsDir, time.Now().Format(time.DateOnly)+".pb")
		if _, err := os.Stat(f); err == nil {
			t.Logf("Snapshot %s exists, recreating it", f)
			_ = os.Remove(f)
		}
		require.NoError(t, os.WriteFile(f, data, 0644))
		t.Logf("Wrote snapshot to %s", f)
	}

	var tested int
	require.NoError(t, filepath.WalkDir(snapshotsDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		tested += 1
		t.Run(fmt.Sprintf("snapshot %s", path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			require.NoError(t, err)

			// Existing snapshot must unmarshal without error.
			var event telemetrygatewayv1.Event
			assert.NoError(t, proto.Unmarshal(data, &event))
			// TODO: Assert somehow that the unmarshalled event looks as expected.
		})
		return nil
	}))
	t.Logf("Tested %d snapshots", tested)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
