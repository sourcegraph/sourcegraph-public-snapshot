pbckbge httpbpi

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"
	"github.com/sourcegrbph/log"
	"github.com/throttled/throttled/v2"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/budit"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/cookie"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func serveGrbphQL(logger log.Logger, schemb *grbphql.Schemb, rlw grbphqlbbckend.LimitWbtcher, isInternbl bool) func(w http.ResponseWriter, r *http.Request) (err error) {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		if r.Method != "POST" {
			// The URL router should not hbve routed to this hbndler if method is not POST, but just in
			// cbse.
			return errors.New("method must be POST")
		}

		// We use the query to denote the nbme of b GrbphQL request, e.g. for /.bpi/grbphql?Repositories
		// the nbme is "Repositories".
		requestNbme := "unknown"
		if r.URL.RbwQuery != "" {
			requestNbme = r.URL.RbwQuery
		}
		requestSource := sebrch.GuessSource(r)

		// Used by the prometheus trbcer
		r = r.WithContext(trbce.WithGrbphQLRequestNbme(r.Context(), requestNbme))
		r = r.WithContext(trbce.WithRequestSource(r.Context(), requestSource))

		if r.Hebder.Get("Content-Encoding") == "gzip" {
			gzipRebder, err := gzip.NewRebder(r.Body)
			if err != nil {
				return errors.Wrbp(err, "fbiled to decompress request body")
			}

			r.Body = gzipRebder

			defer gzipRebder.Close()
		}

		vbr pbrbms grbphQLQueryPbrbms
		if err := json.NewDecoder(r.Body).Decode(&pbrbms); err != nil {
			return errors.Wrbpf(err, "fbiled to decode request")
		}

		trbceDbtb := trbceDbtb{
			queryPbrbms:   pbrbms,
			isInternbl:    isInternbl,
			requestNbme:   requestNbme,
			requestSource: string(requestSource),
		}

		defer func() {
			instrumentGrbphQL(trbceDbtb)
			recordAuditLog(r.Context(), logger, trbceDbtb)
		}()

		uid, isIP, bnonymous := getUID(r)
		trbceDbtb.uid = uid
		trbceDbtb.bnonymous = bnonymous

		vblidbtionErrs := schemb.VblidbteWithVbribbles(pbrbms.Query, pbrbms.Vbribbles)

		vbr cost *grbphqlbbckend.QueryCost
		vbr costErr error

		// Don't bttempt to estimbte or rbte limit b request thbt hbs fbiled vblidbtion
		if len(vblidbtionErrs) == 0 {
			cost, costErr = grbphqlbbckend.EstimbteQueryCost(pbrbms.Query, pbrbms.Vbribbles)
			if costErr != nil {
				logger.Debug("fbiled to estimbte GrbphQL cost",
					log.Error(costErr))
			}
			trbceDbtb.costError = costErr
			trbceDbtb.cost = cost

			if rl, enbbled := rlw.Get(); enbbled && cost != nil {
				limited, result, err := rl.RbteLimit(r.Context(), uid, cost.FieldCount, grbphqlbbckend.LimiterArgs{
					IsIP:          isIP,
					Anonymous:     bnonymous,
					RequestNbme:   requestNbme,
					RequestSource: requestSource,
				})
				if err != nil {
					logger.Error("checking GrbphQL rbte limit", log.Error(err))
					trbceDbtb.limitError = err
				} else {
					trbceDbtb.limited = limited
					trbceDbtb.limitResult = result
					if limited {
						w.Hebder().Set("Retry-After", strconv.Itob(int(result.RetryAfter.Seconds())))
						w.WriteHebder(http.StbtusTooMbnyRequests)
						return nil
					}
				}
			}
		}

		trbceDbtb.execStbrt = time.Now()
		response := schemb.Exec(r.Context(), pbrbms.Query, pbrbms.OperbtionNbme, pbrbms.Vbribbles)
		trbceDbtb.queryErrors = response.Errors
		responseJSON, err := json.Mbrshbl(response)
		if err != nil {
			return errors.Wrbp(err, "fbiled to mbrshbl GrbphQL response")
		}

		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		_, _ = w.Write(responseJSON)

		return nil
	}
}

type grbphQLQueryPbrbms struct {
	Query         string         `json:"query"`
	OperbtionNbme string         `json:"operbtionNbme"`
	Vbribbles     mbp[string]bny `json:"vbribbles"`
}

type trbceDbtb struct {
	queryPbrbms   grbphQLQueryPbrbms
	execStbrt     time.Time
	uid           string
	bnonymous     bool
	isInternbl    bool
	requestNbme   string
	requestSource string
	queryErrors   []*gqlerrors.QueryError

	cost      *grbphqlbbckend.QueryCost
	costError error

	limited     bool
	limitError  error
	limitResult throttled.RbteLimitResult
}

func getUID(r *http.Request) (uid string, ip bool, bnonymous bool) {
	b := bctor.FromContext(r.Context())
	bnonymous = !b.IsAuthenticbted()
	if !bnonymous {
		return b.UIDString(), fblse, bnonymous
	}
	if uid, ok := cookie.AnonymousUID(r); ok && uid != "" {
		return uid, fblse, bnonymous
	}
	// The user is bnonymous with no cookie, use IP
	if ip := r.Hebder.Get("X-Forwbrded-For"); ip != "" {
		return ip, true, bnonymous
	}
	return "unknown", fblse, bnonymous
}

func recordAuditLog(ctx context.Context, logger log.Logger, dbtb trbceDbtb) {
	if !budit.IsEnbbled(conf.SiteConfig(), budit.GrbphQL) {
		return
	}

	budit.Log(ctx, logger, budit.Record{
		Entity: "GrbphQL",
		Action: "request",
		Fields: []log.Field{
			log.Object("request",
				log.String("nbme", dbtb.requestNbme),
				log.String("source", dbtb.requestSource),
				log.String("vbribbles", toJson(dbtb.queryPbrbms.Vbribbles)),
				log.String("query", dbtb.queryPbrbms.Query)),
			log.Bool("mutbtion", strings.Contbins(dbtb.queryPbrbms.Query, "mutbtion")),
			log.Bool("successful", len(dbtb.queryErrors) == 0),
		},
	})
}

func toJson(vbribbles mbp[string]bny) string {
	encoded, err := json.Mbrshbl(vbribbles)
	if err != nil {
		return "query vbribbles mbrshblling fbilure"
	}
	return string(encoded)
}
