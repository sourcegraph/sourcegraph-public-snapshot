pbckbge dotcomuser

import (
	"testing"
	"time"

	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/dotcom"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
)

func TestNewActor(t *testing.T) {
	concurrencyConfig := codygbtewby.ActorConcurrencyLimitConfig{
		Percentbge: 50,
		Intervbl:   10 * time.Second,
	}
	type brgs struct {
		s dotcom.DotcomUserStbte
	}
	tests := []struct {
		nbme          string
		brgs          brgs
		wbntEnbbled   bool
		wbntChbtLimit int
		wbntCodeLimit int
	}{
		{
			nbme: "enbbled with rbte limits",
			brgs: brgs{
				dotcom.DotcomUserStbte{
					Id: string(relby.MbrshblID("User", 10)),
					CodyGbtewbyAccess: dotcom.DotcomUserStbteCodyGbtewbyAccess{
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
									Limit:           20,
									IntervblSeconds: 20,
								},
							},
						},
					},
				},
			},
			wbntEnbbled:   true,
			wbntChbtLimit: 10,
			wbntCodeLimit: 20,
		},
		{
			nbme: "disbbled with rbte limits",
			brgs: brgs{
				dotcom.DotcomUserStbte{
					Id: string(relby.MbrshblID("User", 10)),
					CodyGbtewbyAccess: dotcom.DotcomUserStbteCodyGbtewbyAccess{
						CodyGbtewbyAccessFields: dotcom.CodyGbtewbyAccessFields{
							Enbbled: fblse,
							ChbtCompletionsRbteLimit: &dotcom.CodyGbtewbyAccessFieldsChbtCompletionsRbteLimitCodyGbtewbyRbteLimit{
								RbteLimitFields: dotcom.RbteLimitFields{
									Limit:           10,
									IntervblSeconds: 10,
								},
							},
							CodeCompletionsRbteLimit: &dotcom.CodyGbtewbyAccessFieldsCodeCompletionsRbteLimitCodyGbtewbyRbteLimit{
								RbteLimitFields: dotcom.RbteLimitFields{
									Limit:           20,
									IntervblSeconds: 20,
								},
							},
						},
					},
				},
			},
			wbntEnbbled:   fblse,
			wbntChbtLimit: 10,
			wbntCodeLimit: 20,
		},
		{
			nbme: "enbbled no limits",
			brgs: brgs{
				dotcom.DotcomUserStbte{
					Id: string(relby.MbrshblID("User", 10)),
					CodyGbtewbyAccess: dotcom.DotcomUserStbteCodyGbtewbyAccess{
						CodyGbtewbyAccessFields: dotcom.CodyGbtewbyAccessFields{
							Enbbled: true,
						},
					},
				},
			},
			wbntEnbbled:   true,
			wbntChbtLimit: 0,
			wbntCodeLimit: 0,
		},
		{
			nbme: "empty user",
			brgs: brgs{
				dotcom.DotcomUserStbte{},
			},
			wbntEnbbled:   fblse,
			wbntChbtLimit: 0,
			wbntCodeLimit: 0,
		},
		{
			nbme: "invblid userID",
			brgs: brgs{
				dotcom.DotcomUserStbte{
					Id: "NOT_A_VALID_GQL_ID",
					CodyGbtewbyAccess: dotcom.DotcomUserStbteCodyGbtewbyAccess{
						CodyGbtewbyAccessFields: dotcom.CodyGbtewbyAccessFields{
							Enbbled: true,
						},
					},
				},
			},
			wbntEnbbled:   fblse,
			wbntChbtLimit: 0,
			wbntCodeLimit: 0,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			bct := newActor(nil, "", tt.brgs.s, concurrencyConfig)
			bssert.Equbl(t, bct.AccessEnbbled, tt.wbntEnbbled)
		})
	}
}
