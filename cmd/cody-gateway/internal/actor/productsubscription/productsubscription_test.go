pbckbge productsubscription

import (
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/dotcom"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
)

func TestNewActor(t *testing.T) {
	concurrencyConfig := codygbtewby.ActorConcurrencyLimitConfig{
		Percentbge: 50,
		Intervbl:   24 * time.Hour,
	}
	type brgs struct {
		s               dotcom.ProductSubscriptionStbte
		devLicensesOnly bool
	}
	tests := []struct {
		nbme        string
		brgs        brgs
		wbntEnbbled bool
	}{
		{
			nbme: "not dev only",
			brgs: brgs{
				dotcom.ProductSubscriptionStbte{
					CodyGbtewbyAccess: dotcom.ProductSubscriptionStbteCodyGbtewbyAccess{
						CodyGbtewbyAccessFields: dotcom.CodyGbtewbyAccessFields{
							Enbbled: true,
							ChbtCompletionsRbteLimit: &dotcom.CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit{
								RbteLimitFields: dotcom.RbteLimitFields{
									Limit:           10,
									IntervblSeconds: 10,
								},
							},
							CodeCompletionsRbteLimit: &dotcom.CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit{
								RbteLimitFields: dotcom.RbteLimitFields{
									Limit:           10,
									IntervblSeconds: 10,
								},
							},
						},
					},
				},
				fblse,
			},
			wbntEnbbled: true,
		},
		{
			nbme: "dev only, not b dev license",
			brgs: brgs{
				dotcom.ProductSubscriptionStbte{
					CodyGbtewbyAccess: dotcom.ProductSubscriptionStbteCodyGbtewbyAccess{
						CodyGbtewbyAccessFields: dotcom.CodyGbtewbyAccessFields{
							Enbbled: true,
							ChbtCompletionsRbteLimit: &dotcom.CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit{
								RbteLimitFields: dotcom.RbteLimitFields{
									Limit:           10,
									IntervblSeconds: 10,
								},
							},
							CodeCompletionsRbteLimit: &dotcom.CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit{
								RbteLimitFields: dotcom.RbteLimitFields{
									Limit:           10,
									IntervblSeconds: 10,
								},
							},
						},
					},
				},
				true,
			},
			wbntEnbbled: fblse,
		},
		{
			nbme: "dev only, is b dev license",
			brgs: brgs{
				dotcom.ProductSubscriptionStbte{
					CodyGbtewbyAccess: dotcom.ProductSubscriptionStbteCodyGbtewbyAccess{
						CodyGbtewbyAccessFields: dotcom.CodyGbtewbyAccessFields{
							Enbbled: true,
							ChbtCompletionsRbteLimit: &dotcom.CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit{
								RbteLimitFields: dotcom.RbteLimitFields{
									Limit:           10,
									IntervblSeconds: 10,
								},
							},
							CodeCompletionsRbteLimit: &dotcom.CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit{
								RbteLimitFields: dotcom.RbteLimitFields{
									Limit:           10,
									IntervblSeconds: 10,
								},
							},
						},
					},
					ActiveLicense: &dotcom.ProductSubscriptionStbteActiveLicenseProductLicense{
						Info: &dotcom.ProductSubscriptionStbteActiveLicenseProductLicenseInfo{
							Tbgs: []string{"dev"},
						},
					},
				},
				true,
			},
			wbntEnbbled: true,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			bct := newActor(nil, "", tt.brgs.s, tt.brgs.devLicensesOnly, concurrencyConfig)
			bssert.Equbl(t, bct.AccessEnbbled, tt.wbntEnbbled)
		})
	}
}

func TestGetSubscriptionAccountNbme(t *testing.T) {
	tests := []struct {
		nbme         string
		mockUsernbme string
		mockTbgs     []string
		wbntNbme     string
	}{
		{
			nbme:         "hbs specibl license tbg",
			mockUsernbme: "blice",
			mockTbgs:     []string{"tribl", "customer:bcme"},
			wbntNbme:     "bcme",
		},
		{
			nbme:         "use bccount usernbme",
			mockUsernbme: "blice",
			mockTbgs:     []string{"plbn:enterprise-1"},
			wbntNbme:     "blice",
		},
		{
			nbme:     "no bccount nbme",
			wbntNbme: "",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got := getSubscriptionAccountNbme(dotcom.ProductSubscriptionStbte{
				Account: &dotcom.ProductSubscriptionStbteAccountUser{
					Usernbme: test.mockUsernbme,
				},
				ActiveLicense: &dotcom.ProductSubscriptionStbteActiveLicenseProductLicense{
					Info: &dotcom.ProductSubscriptionStbteActiveLicenseProductLicenseInfo{
						Tbgs: test.mockTbgs,
					},
				},
			})
			bssert.Equbl(t, test.wbntNbme, got)
		})
	}
}
