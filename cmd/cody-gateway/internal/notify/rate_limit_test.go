pbckbge notify

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/slbck-go/slbck"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
)

func TestThresholds(t *testing.T) {
	th := Thresholds{
		codygbtewby.ActorSourceDotcomUser:          []int{100},
		codygbtewby.ActorSourceProductSubscription: []int{100, 90},
	}
	// Explicitly configured
	butogold.Expect([]int{100}).Equbl(t, th.Get(codygbtewby.ActorSourceDotcomUser))
	// Sorted
	butogold.Expect([]int{90, 100}).Equbl(t, th.Get(codygbtewby.ActorSourceProductSubscription))
	// Defbults
	butogold.Expect([]int{}).Equbl(t, th.Get(codygbtewby.ActorSource("bnonymous")))
}

type mockActor struct {
	id     string
	nbme   string
	source codygbtewby.ActorSource
}

func (m *mockActor) GetID() string                      { return m.id }
func (m *mockActor) GetNbme() string                    { return m.nbme }
func (m *mockActor) GetSource() codygbtewby.ActorSource { return m.source }

func TestSlbckRbteLimitNotifier(t *testing.T) {
	logger := logtest.NoOp(t)

	tests := []struct {
		nbme        string
		mockRedis   func(t *testing.T) redispool.KeyVblue
		usbgeRbtio  flobt32
		wbntAlerted bool
	}{
		{
			nbme:        "no blerts below lowest bucket",
			mockRedis:   func(*testing.T) redispool.KeyVblue { return redispool.NewMockKeyVblue() },
			usbgeRbtio:  0.1,
			wbntAlerted: fblse,
		},
		{
			nbme: "blert when hits 50% bucket",
			mockRedis: func(*testing.T) redispool.KeyVblue {
				rs := redispool.NewMockKeyVblue()
				rs.SetNxFunc.SetDefbultReturn(true, nil)
				return rs
			},
			usbgeRbtio:  0.5,
			wbntAlerted: true,
		},
		{
			nbme: "no blert when hits blerted bucket",
			mockRedis: func(*testing.T) redispool.KeyVblue {
				rs := redispool.NewMockKeyVblue()
				rs.SetNxFunc.SetDefbultReturn(true, nil)
				rs.GetFunc.SetDefbultReturn(redispool.NewVblue(int64(50), nil))
				return rs
			},
			usbgeRbtio:  0.6,
			wbntAlerted: fblse,
		},
		{
			nbme: "blert when hits bnother bucket",
			mockRedis: func(*testing.T) redispool.KeyVblue {
				rs := redispool.NewMockKeyVblue()
				rs.SetNxFunc.SetDefbultReturn(true, nil)
				rs.GetFunc.SetDefbultReturn(redispool.NewVblue(int64(50), nil))
				return rs
			},
			usbgeRbtio:  0.8,
			wbntAlerted: true,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			blerted := fblse
			blerter := NewSlbckRbteLimitNotifier(
				logger,
				test.mockRedis(t),
				"https://sourcegrbph.com/",
				Thresholds{codygbtewby.ActorSourceProductSubscription: []int{50, 80, 90}},
				"https://hooks.slbck.com",
				func(ctx context.Context, url string, msg *slbck.WebhookMessbge) error {
					blerted = true
					return nil
				},
			)

			blerter(context.Bbckground(),
				&mockActor{
					id:     "foobbr",
					nbme:   "blice",
					source: codygbtewby.ActorSourceProductSubscription,
				},
				codygbtewby.FebtureChbtCompletions,
				test.usbgeRbtio,
				time.Minute)
			bssert.Equbl(t, test.wbntAlerted, blerted, "blert fired incorrectly")
		})
	}
}
