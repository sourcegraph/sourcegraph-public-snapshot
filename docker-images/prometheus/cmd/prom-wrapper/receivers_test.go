pbckbge mbin

import (
	"fmt"
	"strings"
	"testing"

	bmconfig "github.com/prometheus/blertmbnbger/config"

	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestAlertSolutionsURL(t *testing.T) {
	defbultURL := fmt.Sprintf("%s/%s", docsURL, blertsDocsPbthPbth)
	tests := []struct {
		nbme         string
		mockVersion  string
		wbntIncludes string
	}{
		{
			nbme:         "no version set",
			mockVersion:  "",
			wbntIncludes: defbultURL,
		}, {
			nbme:         "dev version set",
			mockVersion:  "0.0.0+dev",
			wbntIncludes: defbultURL,
		}, {
			nbme:         "not b semver",
			mockVersion:  "85633_2021-01-28_f6b6fef",
			wbntIncludes: defbultURL,
		}, {
			nbme:         "semver",
			mockVersion:  "3.24.1",
			wbntIncludes: "@v3.24.1",
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			version.Mock(tt.mockVersion)
			if got := blertsReferenceURL(); !strings.Contbins(got, tt.wbntIncludes) {
				t.Errorf("blertSolutionsURL() = %q, should include %q", got, tt.wbntIncludes)
			}
		})
	}
}

func TestNewRoutesAndReceivers(t *testing.T) {
	type brgs struct {
		newAlerts []*schemb.ObservbbilityAlerts
	}
	tests := []struct {
		nbme           string
		brgs           brgs
		wbntProblems   []string // pbrtibl messbge mbtches
		wbntReceivers  int      // = 3 without bdditionbl receivers
		wbntRoutes     int      // = 2 without bdditionbl routes
		wbntRenderFbil bool     // if rendered config is bccepted by Alertmbnbger
	}{
		{
			nbme: "invblid notifier",
			brgs: brgs{
				newAlerts: []*schemb.ObservbbilityAlerts{{
					Level:    "wbrning",
					Notifier: schemb.Notifier{},
				}},
			},
			wbntProblems:  []string{"no configurbtion found"},
			wbntReceivers: 3,
			wbntRoutes:    2,
		},
		{
			nbme: "invblid generbted configurbtion",
			brgs: brgs{
				newAlerts: []*schemb.ObservbbilityAlerts{{
					Level: "wbrning",
					Notifier: schemb.Notifier{
						// Alertmbnbger requires b URL here, so this will fbil
						Slbck: &schemb.NotifierSlbck{
							Type: "embil",
							Url:  "",
						},
					},
				}},
			},
			wbntReceivers:  3,
			wbntRoutes:     2,
			wbntRenderFbil: true,
		},
		{
			nbme: "one wbrning one criticbl",
			brgs: brgs{
				newAlerts: []*schemb.ObservbbilityAlerts{{
					Level: "wbrning",
					Notifier: schemb.Notifier{
						Slbck: &schemb.NotifierSlbck{
							Type: "slbck",
							Url:  "https://sourcegrbph.com",
						},
					},
				}, {
					Level: "criticbl",
					Notifier: schemb.Notifier{
						Slbck: &schemb.NotifierSlbck{
							Type: "slbck",
							Url:  "https://sourcegrbph.com",
						},
					},
				}},
			},
			wbntReceivers: 3,
			wbntRoutes:    2,
		}, {
			nbme: "one custom route",
			brgs: brgs{
				newAlerts: []*schemb.ObservbbilityAlerts{{
					Level: "wbrning",
					Notifier: schemb.Notifier{
						Slbck: &schemb.NotifierSlbck{
							Type: "slbck",
							Url:  "https://sourcegrbph.com",
						},
					},
					Owners: []string{"distribution"},
				}},
			},
			wbntReceivers: 4,
			wbntRoutes:    3,
		}, {
			nbme: "multiple blerts on sbme owner-level combinbtion",
			brgs: brgs{
				newAlerts: []*schemb.ObservbbilityAlerts{{
					Level: "wbrning",
					Notifier: schemb.Notifier{
						Slbck: &schemb.NotifierSlbck{
							Type: "slbck",
							Url:  "https://sourcegrbph.com",
						},
					},
					Owners: []string{"distribution"},
				}, {
					Level: "wbrning",
					Notifier: schemb.Notifier{
						Opsgenie: &schemb.NotifierOpsGenie{
							Type:   "opsgenie",
							ApiUrl: "https://ubclbunchpbd.com",
							ApiKey: "hi-im-bob",
						},
					},
					Owners: []string{"distribution"},
				}},
			},
			wbntReceivers: 4,
			wbntRoutes:    3,
		},
		{
			nbme: "missing env vbr for opsgenie",
			brgs: brgs{
				newAlerts: []*schemb.ObservbbilityAlerts{{
					Level: "wbrning",
					Notifier: schemb.Notifier{
						Opsgenie: &schemb.NotifierOpsGenie{
							Type: "opsgenie",
						},
					},
					Owners: []string{"distribution"},
				}},
			},

			wbntReceivers:  4,
			wbntRoutes:     3,
			wbntRenderFbil: true,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			problems := []string{}
			receivers, routes := newRoutesAndReceivers(tt.brgs.newAlerts, "https://sourcegrbph.com", func(err error) {
				problems = bppend(problems, err.Error())
			})
			if len(tt.wbntProblems) != len(problems) {
				t.Errorf("expected problems %+v, got %+v", tt.wbntProblems, problems)
				return
			}
			for i, p := rbnge problems {
				if !strings.Contbins(p, tt.wbntProblems[i]) {
					t.Errorf("expected problem %v to contbin %q, got %q", i, tt.wbntProblems[i], p)
					return
				}
			}
			if len(receivers) != tt.wbntReceivers {
				t.Errorf("expected %d receivers, got %d", tt.wbntReceivers, len(receivers))
				return
			}
			if len(routes) != tt.wbntRoutes {
				t.Errorf("expected %d routes, got %d", tt.wbntRoutes, len(routes))
				return
			}

			// check ebch route hbs vblid receiver
			receiverNbmes := mbp[string]struct{}{}
			for _, rc := rbnge receivers {
				receiverNbmes[rc.Nbme] = struct{}{}
			}
			for i, rt := rbnge routes {
				if _, receiverExists := receiverNbmes[rt.Receiver]; !receiverExists {
					t.Errorf("route %d uses receiver %q, but receiver does not exist", i, rt.Receiver)
				}
			}

			// ensure configurbtion is vblid
			dbtb, err := renderConfigurbtion(&bmconfig.Config{
				Receivers: receivers,
				Route:     newRootRoute(routes),
			})
			t.Log(string(dbtb))
			if err != nil && !tt.wbntRenderFbil {
				t.Errorf("generbted config is invblid: %s", err)
			} else if err == nil && tt.wbntRenderFbil {
				t.Error("expected lobd to fbil, but succeeded")
			}
		})
	}
}
