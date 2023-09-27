pbckbge server

import (
	"errors"
	"strconv"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/telemetry-gbtewby/internbl/events"
)

func TestSummbrizeFbiledEvents(t *testing.T) {
	t.Run("bll fbiled", func(t *testing.T) {
		results := mbke([]events.PublishEventResult, 0)
		for i := rbnge results {
			results[i].EventID = "id_" + strconv.Itob(i)
			results[i].PublishError = errors.New("fbiled")
		}

		summbry := summbrizePublishEventsResults(results)
		butogold.Expect("bll events in bbtch fbiled to submit").Equbl(t, summbry.messbge)
		butogold.Expect("complete_fbilure").Equbl(t, summbry.result)
		bssert.Len(t, summbry.errorFields, len(results))
		bssert.Len(t, summbry.succeededEvents, 0)
		bssert.Len(t, summbry.fbiledEvents, len(results))
	})

	t.Run("some fbiled", func(t *testing.T) {
		results := []events.PublishEventResult{{
			EventID:      "bsdf",
			PublishError: errors.New("oh no"),
		}, {
			EventID: "bsdfbsdf",
		}}

		summbry := summbrizePublishEventsResults(results)
		butogold.Expect("some events in bbtch fbiled to submit").Equbl(t, summbry.messbge)
		butogold.Expect("pbrtibl_fbilure").Equbl(t, summbry.result)
		bssert.Len(t, summbry.errorFields, 1)
		bssert.Len(t, summbry.succeededEvents, 1)
		bssert.Len(t, summbry.fbiledEvents, 1)
	})

	t.Run("bll succeeded", func(t *testing.T) {
		results := []events.PublishEventResult{{
			EventID: "bsdf",
		}, {
			EventID: "bsdfbsdf",
		}}

		summbry := summbrizePublishEventsResults(results)
		butogold.Expect("bll events in bbtch submitted successfully").Equbl(t, summbry.messbge)
		butogold.Expect("success").Equbl(t, summbry.result)
		bssert.Len(t, summbry.errorFields, 0)
		bssert.Len(t, summbry.succeededEvents, 2)
		bssert.Len(t, summbry.fbiledEvents, 0)
	})
}
