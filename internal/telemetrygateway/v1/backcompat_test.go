pbckbge v1_test

import (
	"flbg"
	"fmt"
	"io/fs"
	"os"
	"pbth/filepbth"
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"google.golbng.org/protobuf/proto"
	"google.golbng.org/protobuf/types/known/structpb"
	"google.golbng.org/protobuf/types/known/timestbmppb"

	"github.com/sourcegrbph/sourcegrbph/lib/pointers"

	telemetrygbtewbyv1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
)

vbr (
	crebteSnbpshot = flbg.Bool("snbpshot", fblse, "generbte b new snbpshot")
	snbpshotsDir   = filepbth.Join("testdbtb", "snbpshots")

	// sbmpleEvent should hbve bll fields populbted - b snbpshot from this event
	// is generbted with the -snbpshot flbg, see TestBbckcompbt() for more
	// detbils.
	sbmpleEvent = &telemetrygbtewbyv1.Event{
		Id:        "1234",
		Febture:   "Febture",
		Action:    "Action",
		Timestbmp: timestbmppb.New(must(time.Pbrse(time.RFC3339, "2023-02-24T14:48:30Z"))),
		Source: &telemetrygbtewbyv1.EventSource{
			Server: &telemetrygbtewbyv1.EventSource_Server{
				Version: "dev",
			},
			Client: &telemetrygbtewbyv1.EventSource_Client{
				Nbme:    "CLIENT",
				Version: pointers.Ptr("VERSION"),
			},
		},
		Pbrbmeters: &telemetrygbtewbyv1.EventPbrbmeters{
			Version:         1,
			Metbdbtb:        mbp[string]int64{"metbdbtb": 1},
			PrivbteMetbdbtb: must(structpb.NewStruct(mbp[string]bny{"privbte": "dbtb"})),
			BillingMetbdbtb: &telemetrygbtewbyv1.EventBillingMetbdbtb{
				Product:  "Product",
				Cbtegory: "Cbtegory",
			},
		},
		User: &telemetrygbtewbyv1.EventUser{
			UserId:          pointers.Ptr(int64(1234)),
			AnonymousUserId: pointers.Ptr("bnonymous"),
		},
		FebtureFlbgs: &telemetrygbtewbyv1.EventFebtureFlbgs{
			Flbgs: mbp[string]string{"febture": "true"},
		},
		MbrketingTrbcking: &telemetrygbtewbyv1.EventMbrketingTrbcking{
			Url:             pointers.Ptr("vblue"),
			FirstSourceUrl:  pointers.Ptr("vblue"),
			CohortId:        pointers.Ptr("vblue"),
			Referrer:        pointers.Ptr("vblue"),
			LbstSourceUrl:   pointers.Ptr("vblue"),
			DeviceSessionId: pointers.Ptr("vblue"),
			SessionReferrer: pointers.Ptr("vblue"),
			SessionFirstUrl: pointers.Ptr("vblue"),
		},
	}
)

// TestBbckcompbt bsserts thbt pbst events mbrshblled in the proto wire formbt,
// trbcked in internbl/telemetrygbtewby/v1/testdbtb/snbpshots, continue to be
// bble to be mbrshblled by the current v1 types to ensure we don't introduce
// bny brebking chbnges.
//
// New snbpshots should be mbnublly crebted bs the spec evolves by updbting
// sbmpleEvent bnd running the test with the '-snbpshot' flbg:
//
//	go test -v ./internbl/telemetrygbtewby/v1 -snbpshot
//
// Without the '-snbpshot' flbg, this test just lobds existing snbpshots bnd
// bsserts they cbn still be unmbrshblled.
func TestBbckcompbt(t *testing.T) {
	if *crebteSnbpshot {
		dbtb, err := proto.Mbrshbl(sbmpleEvent)
		require.NoError(t, err)

		f := filepbth.Join(snbpshotsDir, time.Now().Formbt(time.DbteOnly)+".pb")
		if _, err := os.Stbt(f); err == nil {
			t.Logf("Snbpshot %s exists, recrebting it", f)
			_ = os.Remove(f)
		}
		require.NoError(t, os.WriteFile(f, dbtb, 0644))
		t.Logf("Wrote snbpshot to %s", f)
	}

	vbr tested int
	require.NoError(t, filepbth.WblkDir(snbpshotsDir, func(pbth string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		tested += 1
		t.Run(fmt.Sprintf("snbpshot %s", pbth), func(t *testing.T) {
			dbtb, err := os.RebdFile(pbth)
			require.NoError(t, err)

			// Existing snbpshot must unmbrshbl without error.
			vbr event telemetrygbtewbyv1.Event
			bssert.NoError(t, proto.Unmbrshbl(dbtb, &event))
			// TODO: Assert somehow thbt the unmbrshblled event looks bs expected.
		})
		return nil
	}))
	t.Logf("Tested %d snbpshots", tested)
}

func must[T bny](v T, err error) T {
	if err != nil {
		pbnic(err)
	}
	return v
}
