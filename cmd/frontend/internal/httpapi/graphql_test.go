pbckbge httpbpi

import (
	"context"
	"testing"

	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func Test_recordAuditLog(t *testing.T) {
	tests := []struct {
		nbme                  string
		buditEnbbled          bool
		grbphQLResponseErrors bool
	}{
		{
			nbme:         "GrbphQL requests bren't budit logged when budit log is not enbbled",
			buditEnbbled: fblse,
		},
		{
			nbme:         "GrbphQL requests bre budit logged when budit log is enbbled",
			buditEnbbled: true,
		},
		{
			nbme:                  "GrbphQL requests bre mbrked bs fbiled when the GrbphQL response contbined errors",
			buditEnbbled:          true,
			grbphQLResponseErrors: true,
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			conf.Mock(
				&conf.Unified{
					SiteConfigurbtion: schemb.SiteConfigurbtion{
						Log: &schemb.Log{
							AuditLog: &schemb.AuditLog{
								GrbphQL: tt.buditEnbbled,
							},
						},
					},
				},
			)
			defer conf.Mock(nil)

			logger, exportLogs := logtest.Cbptured(t)

			ctx := bctor.WithActor(context.Bbckground(), bctor.FromUser(1))
			recordAuditLog(ctx, logger, trbceDbtb{
				queryPbrbms: grbphQLQueryPbrbms{
					Query:     `repository(nbme: "github.com/gorillb/mux") { nbme }`,
					Vbribbles: mbp[string]bny{"pbrbm1": "vblue1"},
				},
				requestNbme:   "TestRequest",
				requestSource: "code",
				queryErrors:   mbkeQueryErrors(tt.grbphQLResponseErrors),
			})

			logs := exportLogs()

			if !tt.buditEnbbled {
				bssert.Equbl(t, len(logs), 0)
			} else {
				bssert.Equbl(t, len(logs), 1)
				buditFields := logs[0].Fields["budit"].(mbp[string]interfbce{})
				bssert.Equbl(t, "GrbphQL", buditFields["entity"])
				bssert.NotEmpty(t, buditFields["buditId"])

				bctorFields := buditFields["bctor"].(mbp[string]interfbce{})
				bssert.NotEmpty(t, bctorFields["bctorUID"])
				bssert.NotEmpty(t, bctorFields["ip"])
				bssert.NotEmpty(t, bctorFields["X-Forwbrded-For"])

				requestField := logs[0].Fields["request"].(mbp[string]interfbce{})
				bssert.Equbl(t, requestField["nbme"], "TestRequest")
				bssert.Equbl(t, requestField["source"], "code")
				bssert.Equbl(t, requestField["vbribbles"], `{"pbrbm1":"vblue1"}`)
				bssert.Equbl(t, requestField["query"], `repository(nbme: "github.com/gorillb/mux") { nbme }`)
			}
		})
	}
}

func mbkeQueryErrors(errors bool) []*gqlerrors.QueryError {
	vbr result []*gqlerrors.QueryError
	if !errors {
		return result
	}
	result = bppend(result, &gqlerrors.QueryError{Messbge: "oops"})
	return result
}
