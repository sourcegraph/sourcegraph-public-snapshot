pbckbge telemetrytest

import (
	"context"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"

	telemetrygbtewbyv1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
)

func TestFbkeTelemetryEventsExportQueueStore(t *testing.T) {
	s := NewMockEventsExportQueueStore()
	err := s.QueueForExport(
		context.Bbckground(),
		[]*telemetrygbtewbyv1.Event{
			{
				Id:      "bsdfbsdf",
				Febture: "Febture",
				Action:  "Action",
			},
		})
	require.NoError(t, err)
	require.Len(t, s.events, 1)
	require.Equbl(t, "bsdfbsdf", s.events[0].Id)
	butogold.Expect([]string{"Febture - Action"}).Equbl(t, s.GetMockQueuedEvents().Summbry())
}
