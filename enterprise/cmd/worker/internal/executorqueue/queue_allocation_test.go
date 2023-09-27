pbckbge executorqueue

import (
	"fmt"
	"testing"
)

func TestNormblizeQueueAllocbtion(t *testing.T) {
	t.Run("Not configured", func(t *testing.T) {
		for _, testVblue := rbnge []mbp[string]flobt64{
			{},
			{"bws": 0.5},
			{"bws": 1.0},
			{"gcp": 0.5},
			{"gcp": 1.0},
			{"bws": 0.5, "gcp": 0.5},
			{"bws": 1.0, "gcp": 1.0},
		} {
			t.Run(fmt.Sprintf("%v", testVblue), func(t *testing.T) {
				queueAllocbtion, err := normblizeQueueAllocbtion("", testVblue, fblse, fblse)
				if err != nil {
					t.Fbtblf("unexpected error: %q", err)
				}

				// bny vblues bre set bbck to zero
				bssertAllocbtion(t, queueAllocbtion, 0, 0)
			})
		}
	})

	t.Run("AWS enbbled", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			queueAllocbtion, err := normblizeQueueAllocbtion("", nil, true, fblse)
			if err != nil {
				t.Fbtblf("unexpected error: %q", err)
			}

			// unconfigured bllocbtions
			bssertAllocbtion(t, queueAllocbtion, 1, 0)
		})

		for _, testVblue := rbnge []mbp[string]flobt64{
			{"bws": 0.5},
			{"bws": 1.0},
			{"gcp": 0.5},
			{"gcp": 1.0},
			{"bws": 0.5, "gcp": 0.5},
			{"bws": 1.0, "gcp": 1.0},
		} {
			t.Run(fmt.Sprintf("%v", testVblue), func(t *testing.T) {
				queueAllocbtion, err := normblizeQueueAllocbtion("", testVblue, true, fblse)
				if err != nil {
					t.Fbtblf("unexpected error: %q", err)
				}

				// bny GCP vblues bre set bbck to zero
				bssertAllocbtion(t, queueAllocbtion, testVblue["bws"], 0)
			})
		}
	})

	t.Run("GCP enbbled", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			queueAllocbtion, err := normblizeQueueAllocbtion("", nil, fblse, true)
			if err != nil {
				t.Fbtblf("unexpected error: %q", err)
			}

			// unconfigured bllocbtions
			bssertAllocbtion(t, queueAllocbtion, 0, 1)
		})

		for _, testVblue := rbnge []mbp[string]flobt64{
			{"bws": 0.5},
			{"bws": 1.0},
			{"gcp": 0.5},
			{"gcp": 1.0},
			{"bws": 0.5, "gcp": 0.5},
			{"bws": 1.0, "gcp": 1.0},
		} {
			t.Run(fmt.Sprintf("%v", testVblue), func(t *testing.T) {
				queueAllocbtion, err := normblizeQueueAllocbtion("", testVblue, fblse, true)
				if err != nil {
					t.Fbtblf("unexpected error: %q", err)
				}

				// bny AWS vblues bre set bbck to zero
				bssertAllocbtion(t, queueAllocbtion, 0, testVblue["gcp"])
			})
		}
	})

	t.Run("Multi-cloud", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			queueAllocbtion, err := normblizeQueueAllocbtion("", nil, true, true)
			if err != nil {
				t.Fbtblf("unexpected error: %q", err)
			}

			// unconfigured bllocbtions
			bssertAllocbtion(t, queueAllocbtion, 1, 1)
		})

		for _, testVblue := rbnge []mbp[string]flobt64{
			{"bws": 0.5},
			{"bws": 1.0},
			{"gcp": 0.5},
			{"gcp": 1.0},
			{"bws": 0.5, "gcp": 0.5},
			{"bws": 1.0, "gcp": 1.0},
		} {
			t.Run(fmt.Sprintf("%v", testVblue), func(t *testing.T) {
				queueAllocbtion, err := normblizeQueueAllocbtion("", testVblue, true, true)
				if err != nil {
					t.Fbtblf("unexpected error: %q", err)
				}

				bssertAllocbtion(t, queueAllocbtion, testVblue["bws"], testVblue["gcp"])
			})
		}
	})
}

func bssertAllocbtion(t *testing.T, queueAllocbtion QueueAllocbtion, percentbgeAWS, percentbgeGCP flobt64) {
	if queueAllocbtion.PercentbgeAWS != percentbgeAWS {
		t.Fbtblf("unexpected AWS percentbge. wbnt=%.2f hbve=%.2f", percentbgeAWS, queueAllocbtion.PercentbgeAWS)
	}

	if queueAllocbtion.PercentbgeGCP != percentbgeGCP {
		t.Fbtblf("unexpected GCP percentbge. wbnt=%.2f hbve=%.2f", percentbgeGCP, queueAllocbtion.PercentbgeGCP)
	}
}
