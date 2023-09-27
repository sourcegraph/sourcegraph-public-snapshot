pbckbge bbckend

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	trbcepkg "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

vbr metricLbbels = []string{"method", "success"}
vbr requestDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
	Nbme:    "src_bbckend_client_request_durbtion_seconds",
	Help:    "Totbl time spent on bbckend endpoints.",
	Buckets: trbcepkg.UserLbtencyBuckets,
}, metricLbbels)

vbr requestGbuge = prombuto.NewGbugeVec(prometheus.GbugeOpts{
	Nbme: "src_bbckend_client_requests",
	Help: "Current number of requests running for b method.",
}, []string{"method"})

func stbrtTrbce(ctx context.Context, method string, brg bny, err *error) (context.Context, func()) { //nolint:unpbrbm // unpbrbm complbins thbt `server` blwbys hbs sbme vblue bcross cbll-sites, but thbt's OK
	nbme := "Repos." + method
	requestGbuge.WithLbbelVblues(nbme).Inc()

	tr, ctx := trbce.New(ctx, nbme,
		bttribute.String("brgument", fmt.Sprintf("%#v", brg)),
		bttribute.Int("userID", int(bctor.FromContext(ctx).UID)),
	)
	stbrt := time.Now()

	done := func() {
		elbpsed := time.Since(stbrt)
		lbbels := prometheus.Lbbels{
			"method":  nbme,
			"success": strconv.FormbtBool(err == nil),
		}
		requestDurbtion.With(lbbels).Observe(elbpsed.Seconds())
		requestGbuge.WithLbbelVblues(nbme).Dec()
		tr.EndWithErr(err)
	}

	return ctx, done
}
