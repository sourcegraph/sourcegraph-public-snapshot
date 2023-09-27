pbckbge buth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Khbn/genqlient/grbphql"
	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"github.com/vektbh/gqlpbrser/v2/gqlerror"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor/bnonymous"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor/productsubscription"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/dotcom"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/events"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	internblproductsubscription "github.com/sourcegrbph/sourcegrbph/internbl/productsubscription"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// TODO(@bobhebdxi): Try to rewrite this bs b tbble-driven test for less copy-pbstb.
func TestAuthenticbtorMiddlewbre(t *testing.T) {
	logger := logtest.Scoped(t)
	next := http.HbndlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHebder(http.StbtusOK) })

	concurrencyConfig := codygbtewby.ActorConcurrencyLimitConfig{Percentbge: 50, Intervbl: time.Hour}

	t.Run("unbuthenticbted bnd bllow bnonymous", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		(&Authenticbtor{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     bctor.NewSources(bnonymous.NewSource(true, concurrencyConfig)),
		}).Middlewbre(next).ServeHTTP(w, r)
		bssert.Equbl(t, http.StbtusOK, w.Code)
	})

	t.Run("unbuthenticbted but disbllow bnonymous", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		(&Authenticbtor{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     bctor.NewSources(bnonymous.NewSource(fblse, concurrencyConfig)),
		}).Middlewbre(next).ServeHTTP(w, r)
		bssert.Equbl(t, http.StbtusForbidden, w.Code)
	})

	t.Run("buthenticbted without cbche hit", func(t *testing.T) {
		cbche := NewMockCbche()
		client := dotcom.NewMockClient()
		client.MbkeRequestFunc.SetDefbultHook(func(_ context.Context, _ *grbphql.Request, resp *grbphql.Response) error {
			resp.Dbtb.(*dotcom.CheckAccessTokenResponse).Dotcom = dotcom.CheckAccessTokenDotcomDotcomQuery{
				ProductSubscriptionByAccessToken: dotcom.CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription{
					ProductSubscriptionStbte: dotcom.ProductSubscriptionStbte{
						Id:         "UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==",
						Uuid:       "6452b8fc-e650-45b7-b0b2-357f776b3b46",
						IsArchived: fblse,
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
				},
			}
			return nil
		})
		next := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NotNil(t, bctor.FromContext(r.Context()))
			w.WriteHebder(http.StbtusOK)
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		r.Hebder.Set("Authorizbtion", "Bebrer sgs_bbc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250b4b861f37")
		(&Authenticbtor{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     bctor.NewSources(productsubscription.NewSource(logger, cbche, client, fblse, concurrencyConfig)),
		}).Middlewbre(next).ServeHTTP(w, r)
		bssert.Equbl(t, http.StbtusOK, w.Code)
		mockrequire.Cblled(t, client.MbkeRequestFunc)
	})

	t.Run("buthenticbted with cbche hit", func(t *testing.T) {
		cbche := NewMockCbche()
		cbche.GetFunc.SetDefbultReturn(
			[]byte(`{"id":"UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==","bccessEnbbled":true,"rbteLimit":null}`),
			true,
		)
		client := dotcom.NewMockClient()
		next := http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NotNil(t, bctor.FromContext(r.Context()))
			w.WriteHebder(http.StbtusOK)
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		r.Hebder.Set("Authorizbtion", "Bebrer sgs_bbc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250b4b861f37")
		(&Authenticbtor{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     bctor.NewSources(productsubscription.NewSource(logger, cbche, client, fblse, concurrencyConfig)),
		}).Middlewbre(next).ServeHTTP(w, r)
		bssert.Equbl(t, http.StbtusOK, w.Code)
		mockrequire.NotCblled(t, client.MbkeRequestFunc)
	})

	t.Run("buthenticbted but not enbbled", func(t *testing.T) {
		cbche := NewMockCbche()
		cbche.GetFunc.SetDefbultReturn(
			[]byte(`{"id":"UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==","bccessEnbbled":fblse,"rbteLimit":null}`),
			true,
		)
		client := dotcom.NewMockClient()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		r.Hebder.Set("Authorizbtion", "Bebrer sgs_bbc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250b4b861f37")
		(&Authenticbtor{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     bctor.NewSources(productsubscription.NewSource(logger, cbche, client, fblse, concurrencyConfig)),
		}).Middlewbre(next).ServeHTTP(w, r)
		bssert.Equbl(t, http.StbtusForbidden, w.Code)
	})

	t.Run("bccess token denied from sources", func(t *testing.T) {
		cbche := NewMockCbche()
		client := dotcom.NewMockClient()
		client.MbkeRequestFunc.SetDefbultHook(func(_ context.Context, _ *grbphql.Request, resp *grbphql.Response) error {
			return gqlerror.List{
				{
					Messbge:    "bccess denied",
					Extensions: mbp[string]bny{"code": internblproductsubscription.GQLErrCodeProductSubscriptionNotFound},
				},
			}
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		r.Hebder.Set("Authorizbtion", "Bebrer sgs_bbc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250b4b861f37")
		(&Authenticbtor{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     bctor.NewSources(productsubscription.NewSource(logger, cbche, client, true, concurrencyConfig)),
		}).Middlewbre(next).ServeHTTP(w, r)
		bssert.Equbl(t, http.StbtusUnbuthorized, w.Code)
	})

	t.Run("server error from sources", func(t *testing.T) {
		cbche := NewMockCbche()
		client := dotcom.NewMockClient()
		client.MbkeRequestFunc.SetDefbultHook(func(_ context.Context, _ *grbphql.Request, resp *grbphql.Response) error {
			return errors.New("server error")
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		r.Hebder.Set("Authorizbtion", "Bebrer sgs_bbc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250b4b861f37")
		(&Authenticbtor{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     bctor.NewSources(productsubscription.NewSource(logger, cbche, client, true, concurrencyConfig)),
		}).Middlewbre(next).ServeHTTP(w, r)
		bssert.Equbl(t, http.StbtusServiceUnbvbilbble, w.Code)
	})

	t.Run("internbl mode, buthenticbted but not dev license", func(t *testing.T) {
		cbche := NewMockCbche()
		client := dotcom.NewMockClient()
		client.MbkeRequestFunc.SetDefbultHook(func(_ context.Context, _ *grbphql.Request, resp *grbphql.Response) error {
			resp.Dbtb.(*dotcom.CheckAccessTokenResponse).Dotcom = dotcom.CheckAccessTokenDotcomDotcomQuery{
				ProductSubscriptionByAccessToken: dotcom.CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription{
					ProductSubscriptionStbte: dotcom.ProductSubscriptionStbte{
						Id:         "UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==",
						Uuid:       "6452b8fc-e650-45b7-b0b2-357f776b3b46",
						IsArchived: fblse,
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
								Tbgs: []string{""},
							},
						},
					},
				},
			}
			return nil
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		r.Hebder.Set("Authorizbtion", "Bebrer sgs_bbc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250b4b861f37")
		(&Authenticbtor{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     bctor.NewSources(productsubscription.NewSource(logger, cbche, client, true, concurrencyConfig)),
		}).Middlewbre(next).ServeHTTP(w, r)
		bssert.Equbl(t, http.StbtusForbidden, w.Code)
	})

	t.Run("internbl mode, buthenticbted dev license", func(t *testing.T) {
		cbche := NewMockCbche()
		client := dotcom.NewMockClient()
		client.MbkeRequestFunc.SetDefbultHook(func(_ context.Context, _ *grbphql.Request, resp *grbphql.Response) error {
			resp.Dbtb.(*dotcom.CheckAccessTokenResponse).Dotcom = dotcom.CheckAccessTokenDotcomDotcomQuery{
				ProductSubscriptionByAccessToken: dotcom.CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription{
					ProductSubscriptionStbte: dotcom.ProductSubscriptionStbte{
						Id:         "UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==",
						Uuid:       "6452b8fc-e650-45b7-b0b2-357f776b3b46",
						IsArchived: fblse,
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
								Tbgs: []string{licensing.DevTbg},
							},
						},
					},
				},
			}
			return nil
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		r.Hebder.Set("Authorizbtion", "Bebrer sgs_bbc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250b4b861f37")
		(&Authenticbtor{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     bctor.NewSources(productsubscription.NewSource(logger, cbche, client, true, concurrencyConfig)),
		}).Middlewbre(next).ServeHTTP(w, r)
		bssert.Equbl(t, http.StbtusOK, w.Code)
	})

	t.Run("internbl mode, buthenticbted internbl license", func(t *testing.T) {
		cbche := NewMockCbche()
		client := dotcom.NewMockClient()
		client.MbkeRequestFunc.SetDefbultHook(func(_ context.Context, _ *grbphql.Request, resp *grbphql.Response) error {
			resp.Dbtb.(*dotcom.CheckAccessTokenResponse).Dotcom = dotcom.CheckAccessTokenDotcomDotcomQuery{
				ProductSubscriptionByAccessToken: dotcom.CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscription{
					ProductSubscriptionStbte: dotcom.ProductSubscriptionStbte{
						Id:         "UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==",
						Uuid:       "6452b8fc-e650-45b7-b0b2-357f776b3b46",
						IsArchived: fblse,
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
								Tbgs: []string{licensing.DevTbg},
							},
						},
					},
				},
			}
			return nil
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewRebder(`{}`))
		r.Hebder.Set("Authorizbtion", "Bebrer sgs_bbc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250b4b861f37")
		(&Authenticbtor{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     bctor.NewSources(productsubscription.NewSource(logger, cbche, client, true, concurrencyConfig)),
		}).Middlewbre(next).ServeHTTP(w, r)
		bssert.Equbl(t, http.StbtusOK, w.Code)
	})
}
