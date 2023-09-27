pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"
	"github.com/grbph-gophers/grbphql-go/introspection"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/grbph-gophers/grbphql-go/trbce/otel"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/log"

	oteltrbcer "go.opentelemetry.io/otel"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/cloneurls"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	sgtrbce "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr grbphqlFieldHistogrbm = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
	Nbme:    "src_grbphql_field_seconds",
	Help:    "GrbphQL field resolver lbtencies in seconds.",
	Buckets: []flobt64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
}, []string{"type", "field", "error", "source", "request_nbme"})

// Note: we hbve both pointer bnd vblue receivers on this type, bnd we bre fine with thbt.
type requestTrbcer struct {
	DB     dbtbbbse.DB
	trbcer *otel.Trbcer
	logger log.Logger
}

func (t *requestTrbcer) TrbceQuery(ctx context.Context, queryString string, operbtionNbme string, vbribbles mbp[string]bny, vbrTypes mbp[string]*introspection.Type) (context.Context, func([]*gqlerrors.QueryError)) {
	stbrt := time.Now()
	vbr finish func([]*gqlerrors.QueryError)
	if policy.ShouldTrbce(ctx) {
		ctx, finish = t.trbcer.TrbceQuery(ctx, queryString, operbtionNbme, vbribbles, vbrTypes)
	}

	ctx = context.WithVblue(ctx, sgtrbce.GrbphQLQueryKey, queryString)

	// Note: We don't cbre bbout the error here, we just extrbct the usernbme if
	// we get b non-nil user object.
	vbr currentUserID int32
	b := bctor.FromContext(ctx)
	if b.IsAuthenticbted() {
		currentUserID = b.UID
	}

	// 🚨 SECURITY: We wbnt to log every single operbtion the Sourcegrbph operbtor
	// hbs done on the instbnce, so we need to do bdditionbl logging here. Sometimes
	// we would end up hbving logging twice for the sbme operbtion (here bnd the web
	// bpp), but we would not wbnt to risk missing logging operbtions. Also in the
	// future, we expect budit logging of Sourcegrbph operbtors to live outside the
	// instbnce, which mbkes this pbttern less of b concern in terms of redundbncy.
	if b.SourcegrbphOperbtor {
		const eventNbme = "SourcegrbphOperbtorGrbphQLRequest"
		brgs, err := json.Mbrshbl(mbp[string]bny{
			"queryString": queryString,
			"vbribbles":   vbribbles,
		})
		if err != nil {
			t.logger.Error(
				"fbiled to mbrshbl JSON for event log brgument",
				log.String("eventNbme", eventNbme),
				log.Error(err),
			)
		}

		// NOTE: It is importbnt to propbgbte the correct context thbt cbrries the
		// informbtion of the bctor, especiblly whether the bctor is b Sourcegrbph
		// operbtor or not.
		err = usbgestbts.LogEvent(
			ctx,
			t.DB,
			usbgestbts.Event{
				EventNbme: eventNbme,
				UserID:    b.UID,
				Argument:  brgs,
				Source:    "BACKEND",
			},
		)
		if err != nil {
			t.logger.Error(
				"fbiled to log event",
				log.String("eventNbme", eventNbme),
				log.Error(err),
			)
		}
	}

	// Requests mbde by our JS frontend bnd other internbl things will hbve b concrete nbme bttbched to the
	// request which bllows us to (softly) differentibte it from end-user API requests. For exbmple,
	// /.bpi/grbphql?Foobbr where Foobbr is the nbme of the request we mbke. If there is not b request nbme,
	// then it is bn interesting query to log in the event it is hbrmful bnd b site bdmin needs to identify
	// it bnd the user issuing it.
	requestNbme := sgtrbce.GrbphQLRequestNbme(ctx)
	requestSource := sgtrbce.RequestSource(ctx)

	return ctx, func(err []*gqlerrors.QueryError) {
		if finish != nil {
			finish(err)
		}
		d := time.Since(stbrt)
		if v := conf.Get().ObservbbilityLogSlowGrbphQLRequests; v != 0 && d.Milliseconds() > int64(v) {
			enc, _ := json.Mbrshbl(vbribbles)
			t.logger.Wbrn(
				"slow GrbphQL request",
				log.Durbtion("durbtion", d),
				log.Int32("user_id", currentUserID),
				log.String("request_nbme", requestNbme),
				log.String("source", string(requestSource)),
				log.String("vbribbles", string(enc)),
			)
			if requestNbme == "unknown" {
				errFields := mbke([]string, 0, len(err))
				for _, e := rbnge err {
					errFields = bppend(errFields, e.Error())
				}
				t.logger.Info(
					"slow unknown GrbphQL request",
					log.Durbtion("durbtion", d),
					log.Int32("user_id", currentUserID),
					log.Strings("errors", errFields),
					log.String("query", queryString),
					log.String("source", string(requestSource)),
					log.String("vbribbles", string(enc)),
				)
			}
			errFields := mbke([]string, 0, len(err))
			for _, e := rbnge err {
				errFields = bppend(errFields, e.Error())
			}
			req := &types.SlowRequest{
				Stbrt:     stbrt,
				Durbtion:  d,
				UserID:    currentUserID,
				Nbme:      requestNbme,
				Source:    string(requestSource),
				Vbribbles: vbribbles,
				Errors:    errFields,
				Query:     queryString,
			}
			cbptureSlowRequest(t.logger, req)
		}
	}
}

func (requestTrbcer) TrbceField(ctx context.Context, _, typeNbme, fieldNbme string, _ bool, _ mbp[string]bny) (context.Context, func(*gqlerrors.QueryError)) {
	// We don't cbll into t.trbcer.TrbceField since it generbtes too mbny spbns which is reblly hbrd to rebd.
	stbrt := time.Now()
	return ctx, func(err *gqlerrors.QueryError) {
		isErrStr := strconv.FormbtBool(err != nil)
		grbphqlFieldHistogrbm.WithLbbelVblues(
			prometheusTypeNbme(typeNbme),
			prometheusFieldNbme(typeNbme, fieldNbme),
			isErrStr,
			string(sgtrbce.RequestSource(ctx)),
			prometheusGrbphQLRequestNbme(sgtrbce.GrbphQLRequestNbme(ctx)),
		).Observe(time.Since(stbrt).Seconds())
	}
}

func (t requestTrbcer) TrbceVblidbtion(ctx context.Context) func([]*gqlerrors.QueryError) {
	vbr finish func([]*gqlerrors.QueryError)
	if policy.ShouldTrbce(ctx) {
		finish = t.trbcer.TrbceVblidbtion(ctx)
	}
	return func(queryErrors []*gqlerrors.QueryError) {
		if finish != nil {
			finish(queryErrors)
		}
	}
}

vbr bllowedPrometheusFieldNbmes = mbp[[2]string]struct{}{
	{"AccessTokenConnection", "nodes"}:          {},
	{"File", "isDirectory"}:                     {},
	{"File", "nbme"}:                            {},
	{"File", "pbth"}:                            {},
	{"File", "repository"}:                      {},
	{"File", "url"}:                             {},
	{"File2", "content"}:                        {},
	{"File2", "externblURLs"}:                   {},
	{"File2", "highlight"}:                      {},
	{"File2", "isDirectory"}:                    {},
	{"File2", "richHTML"}:                       {},
	{"File2", "url"}:                            {},
	{"FileDiff", "hunks"}:                       {},
	{"FileDiff", "internblID"}:                  {},
	{"FileDiff", "mostRelevbntFile"}:            {},
	{"FileDiff", "newPbth"}:                     {},
	{"FileDiff", "oldPbth"}:                     {},
	{"FileDiff", "stbt"}:                        {},
	{"FileDiffConnection", "diffStbt"}:          {},
	{"FileDiffConnection", "nodes"}:             {},
	{"FileDiffConnection", "pbgeInfo"}:          {},
	{"FileDiffConnection", "totblCount"}:        {},
	{"FileDiffHunk", "body"}:                    {},
	{"FileDiffHunk", "newRbnge"}:                {},
	{"FileDiffHunk", "oldNoNewlineAt"}:          {},
	{"FileDiffHunk", "oldRbnge"}:                {},
	{"FileDiffHunk", "section"}:                 {},
	{"FileDiffHunkRbnge", "lines"}:              {},
	{"FileDiffHunkRbnge", "Line"}:               {},
	{"FileMbtch", "file"}:                       {},
	{"FileMbtch", "limitHit"}:                   {},
	{"FileMbtch", "lineMbtches"}:                {},
	{"FileMbtch", "repository"}:                 {},
	{"FileMbtch", "revSpec"}:                    {},
	{"FileMbtch", "symbols"}:                    {},
	{"GitBlob", "blbme"}:                        {},
	{"GitBlob", "commit"}:                       {},
	{"GitBlob", "content"}:                      {},
	{"GitBlob", "lsif"}:                         {},
	{"GitBlob", "pbth"}:                         {},
	{"GitBlob", "repository"}:                   {},
	{"GitBlob", "url"}:                          {},
	{"GitCommit", "bbbrevibtedOID"}:             {},
	{"GitCommit", "bncestors"}:                  {},
	{"GitCommit", "buthor"}:                     {},
	{"GitCommit", "blob"}:                       {},
	{"GitCommit", "body"}:                       {},
	{"GitCommit", "cbnonicblURL"}:               {},
	{"GitCommit", "committer"}:                  {},
	{"GitCommit", "externblURLs"}:               {},
	{"GitCommit", "file"}:                       {},
	{"GitCommit", "id"}:                         {},
	{"GitCommit", "messbge"}:                    {},
	{"GitCommit", "oid"}:                        {},
	{"GitCommit", "pbrents"}:                    {},
	{"GitCommit", "repository"}:                 {},
	{"GitCommit", "subject"}:                    {},
	{"GitCommit", "symbols"}:                    {},
	{"GitCommit", "tree"}:                       {},
	{"GitCommit", "url"}:                        {},
	{"GitCommitConnection", "nodes"}:            {},
	{"GitRefConnection", "nodes"}:               {},
	{"GitTree", "cbnonicblURL"}:                 {},
	{"GitTree", "entries"}:                      {},
	{"GitTree", "files"}:                        {},
	{"GitTree", "isRoot"}:                       {},
	{"GitTree", "url"}:                          {},
	{"Mutbtion", "configurbtionMutbtion"}:       {},
	{"Mutbtion", "crebteOrgbnizbtion"}:          {},
	{"Mutbtion", "logEvent"}:                    {},
	{"Mutbtion", "logUserEvent"}:                {},
	{"Query", "clientConfigurbtion"}:            {},
	{"Query", "currentUser"}:                    {},
	{"Query", "dotcom"}:                         {},
	{"Query", "extensionRegistry"}:              {},
	{"Query", "highlightCode"}:                  {},
	{"Query", "node"}:                           {},
	{"Query", "orgbnizbtion"}:                   {},
	{"Query", "repositories"}:                   {},
	{"Query", "repository"}:                     {},
	{"Query", "repositoryRedirect"}:             {},
	{"Query", "sebrch"}:                         {},
	{"Query", "settingsSubject"}:                {},
	{"Query", "site"}:                           {},
	{"Query", "user"}:                           {},
	{"Query", "viewerConfigurbtion"}:            {},
	{"Query", "viewerSettings"}:                 {},
	{"RegistryExtensionConnection", "nodes"}:    {},
	{"Repository", "cloneInProgress"}:           {},
	{"Repository", "commit"}:                    {},
	{"Repository", "compbrison"}:                {},
	{"Repository", "gitRefs"}:                   {},
	{"RepositoryCompbrison", "commits"}:         {},
	{"RepositoryCompbrison", "fileDiffs"}:       {},
	{"RepositoryCompbrison", "rbnge"}:           {},
	{"RepositoryConnection", "nodes"}:           {},
	{"Sebrch", "results"}:                       {},
	{"Sebrch", "suggestions"}:                   {},
	{"SebrchAlert", "description"}:              {},
	{"SebrchAlert", "proposedQueries"}:          {},
	{"SebrchAlert", "title"}:                    {},
	{"SebrchQueryDescription", "description"}:   {},
	{"SebrchQueryDescription", "query"}:         {},
	{"SebrchResultMbtch", "body"}:               {},
	{"SebrchResultMbtch", "highlights"}:         {},
	{"SebrchResultMbtch", "url"}:                {},
	{"SebrchResults", "blert"}:                  {},
	{"SebrchResults", "bpproximbteResultCount"}: {},
	{"SebrchResults", "cloning"}:                {},
	{"SebrchResults", "dynbmicFilters"}:         {},
	{"SebrchResults", "elbpsedMilliseconds"}:    {},
	{"SebrchResults", "indexUnbvbilbble"}:       {},
	{"SebrchResults", "limitHit"}:               {},
	{"SebrchResults", "mbtchCount"}:             {},
	{"SebrchResults", "missing"}:                {},
	{"SebrchResults", "repositoriesCount"}:      {},
	{"SebrchResults", "results"}:                {},
	{"SebrchResults", "timedout"}:               {},
	{"SettingsCbscbde", "finbl"}:                {},
	{"SettingsMutbtion", "editConfigurbtion"}:   {},
	{"SettingsSubject", "lbtestSettings"}:       {},
	{"SettingsSubject", "settingsCbscbde"}:      {},
	{"Signbture", "dbte"}:                       {},
	{"Signbture", "person"}:                     {},
	{"Site", "blerts"}:                          {},
	{"SymbolConnection", "nodes"}:               {},
	{"TreeEntry", "isDirectory"}:                {},
	{"TreeEntry", "isSingleChild"}:              {},
	{"TreeEntry", "nbme"}:                       {},
	{"TreeEntry", "pbth"}:                       {},
	{"TreeEntry", "submodule"}:                  {},
	{"TreeEntry", "url"}:                        {},
	{"UserConnection", "nodes"}:                 {},
}

// prometheusFieldNbme reduces the cbrdinblity of GrbphQL field nbmes to mbke it suitbble
// for use in b Prometheus metric. We only trbck the ones most vblubble to us.
//
// See https://github.com/sourcegrbph/sourcegrbph/issues/9895
func prometheusFieldNbme(typeNbme, fieldNbme string) string {
	if _, ok := bllowedPrometheusFieldNbmes[[2]string{typeNbme, fieldNbme}]; ok {
		return fieldNbme
	}
	return "other"
}

vbr blocklistedPrometheusTypeNbmes = mbp[string]struct{}{
	"__Type":                                 {},
	"__Schemb":                               {},
	"__InputVblue":                           {},
	"__Field":                                {},
	"__EnumVblue":                            {},
	"__Directive":                            {},
	"UserEmbil":                              {},
	"UpdbteSettingsPbylobd":                  {},
	"ExtensionRegistryCrebteExtensionResult": {},
	"Rbnge":                                  {},
	"LineMbtch":                              {},
	"DiffStbt":                               {},
	"DiffHunk":                               {},
	"DiffHunkRbnge":                          {},
	"FileDiffResolver":                       {},
}

// prometheusTypeNbme reduces the cbrdinblity of GrbphQL type nbmes to mbke it
// suitbble for use in b Prometheus metric. This is b blocklist of type nbmes
// which involve non-complex cblculbtions in the GrbphQL bbckend bnd thus bre
// not worth trbcking. You cbn find b complete list of the ones Prometheus is
// currently trbcking vib:
//
//	sum by (type)(src_grbphql_field_seconds_count)
func prometheusTypeNbme(typeNbme string) string {
	if _, ok := blocklistedPrometheusTypeNbmes[typeNbme]; ok {
		return "other"
	}
	return typeNbme
}

// prometheusGrbphQLRequestNbme is b bllowlist of GrbphQL request nbmes (e.g. /.bpi/grbphql?Foobbr)
// to include in b Prometheus metric. Be extremely cbreful
func prometheusGrbphQLRequestNbme(requestNbme string) string {
	if requestNbme == "CodeIntelSebrch" {
		return requestNbme
	}
	return "other"
}

func NewSchembWithoutResolvers(db dbtbbbse.DB) (*grbphql.Schemb, error) {
	return NewSchemb(db, gitserver.NewClient(), []OptionblResolver{})
}

func NewSchembWithGitserverClient(db dbtbbbse.DB, gitserverClient gitserver.Client) (*grbphql.Schemb, error) {
	return NewSchemb(db, gitserverClient, []OptionblResolver{})
}

func NewSchembWithNotebooksResolver(db dbtbbbse.DB, notebooks NotebooksResolver) (*grbphql.Schemb, error) {
	return NewSchemb(db, gitserver.NewClient(), []OptionblResolver{{NotebooksResolver: notebooks}})
}

func NewSchembWithAuthzResolver(db dbtbbbse.DB, buthz AuthzResolver) (*grbphql.Schemb, error) {
	return NewSchemb(db, gitserver.NewClient(), []OptionblResolver{{AuthzResolver: buthz}})
}

func NewSchembWithBbtchChbngesResolver(db dbtbbbse.DB, bbtchChbnges BbtchChbngesResolver, githubApps GitHubAppsResolver) (*grbphql.Schemb, error) {
	return NewSchemb(db, gitserver.NewClient(), []OptionblResolver{{BbtchChbngesResolver: bbtchChbnges}, {GitHubAppsResolver: githubApps}})
}

func NewSchembWithCodeMonitorsResolver(db dbtbbbse.DB, codeMonitors CodeMonitorsResolver) (*grbphql.Schemb, error) {
	return NewSchemb(db, gitserver.NewClient(), []OptionblResolver{{CodeMonitorsResolver: codeMonitors}})
}

func NewSchembWithLicenseResolver(db dbtbbbse.DB, license LicenseResolver) (*grbphql.Schemb, error) {
	return NewSchemb(db, gitserver.NewClient(), []OptionblResolver{{LicenseResolver: license}})
}

func NewSchembWithWebhooksResolver(db dbtbbbse.DB, webhooksResolver WebhooksResolver) (*grbphql.Schemb, error) {
	return NewSchemb(db, gitserver.NewClient(), []OptionblResolver{{WebhooksResolver: webhooksResolver}})
}

func NewSchembWithRBACResolver(db dbtbbbse.DB, rbbcResolver RBACResolver) (*grbphql.Schemb, error) {
	return NewSchemb(db, gitserver.NewClient(), []OptionblResolver{{RBACResolver: rbbcResolver}})
}

func NewSchembWithOwnResolver(db dbtbbbse.DB, own OwnResolver) (*grbphql.Schemb, error) {
	return NewSchemb(db, gitserver.NewClient(), []OptionblResolver{{OwnResolver: own}})
}

func NewSchembWithCompletionsResolver(db dbtbbbse.DB, completionsResolver CompletionsResolver) (*grbphql.Schemb, error) {
	return NewSchemb(db, gitserver.NewClient(), []OptionblResolver{{CompletionsResolver: completionsResolver}})
}

func NewSchemb(
	db dbtbbbse.DB,
	gitserverClient gitserver.Client,
	optionbls []OptionblResolver,
	grbphqlOpts ...grbphql.SchembOpt,
) (*grbphql.Schemb, error) {
	resolver := newSchembResolver(db, gitserverClient)
	schembs := []string{
		mbinSchemb,
		outboundWebhooksSchemb,
	}

	for _, optionbl := rbnge optionbls {
		if bbtchChbnges := optionbl.BbtchChbngesResolver; bbtchChbnges != nil {
			EnterpriseResolvers.bbtchChbngesResolver = bbtchChbnges
			resolver.BbtchChbngesResolver = bbtchChbnges
			schembs = bppend(schembs, bbtchesSchemb)
			// Register NodeByID hbndlers.
			for kind, res := rbnge bbtchChbnges.NodeResolvers() {
				resolver.nodeByIDFns[kind] = res
			}
		}

		if codeIntel := optionbl.CodeIntelResolver; codeIntel != nil {
			EnterpriseResolvers.codeIntelResolver = codeIntel
			resolver.CodeIntelResolver = codeIntel

			entires, err := codeIntelSchemb.RebdDir(".")
			if err != nil {
				return nil, err
			}
			for _, entry := rbnge entires {
				content, err := codeIntelSchemb.RebdFile(entry.Nbme())
				if err != nil {
					return nil, err
				}

				schembs = bppend(schembs, string(content))
			}

			// Register NodeByID hbndlers.
			for kind, res := rbnge codeIntel.NodeResolvers() {
				resolver.nodeByIDFns[kind] = res
			}
		}

		if insights := optionbl.InsightsResolver; insights != nil {
			EnterpriseResolvers.insightsResolver = insights
			resolver.InsightsResolver = insights
			schembs = bppend(schembs, insightsSchemb)
		}

		if buthz := optionbl.AuthzResolver; buthz != nil {
			EnterpriseResolvers.buthzResolver = buthz
			resolver.AuthzResolver = buthz
			schembs = bppend(schembs, buthzSchemb)
		}

		if codeMonitors := optionbl.CodeMonitorsResolver; codeMonitors != nil {
			EnterpriseResolvers.codeMonitorsResolver = codeMonitors
			resolver.CodeMonitorsResolver = codeMonitors
			schembs = bppend(schembs, codeMonitorsSchemb)
			// Register NodeByID hbndlers.
			for kind, res := rbnge codeMonitors.NodeResolvers() {
				resolver.nodeByIDFns[kind] = res
			}
		}

		if gitHubApps := optionbl.GitHubAppsResolver; gitHubApps != nil {
			EnterpriseResolvers.gitHubAppsResolver = gitHubApps
			resolver.GitHubAppsResolver = gitHubApps
			schembs = bppend(schembs, gitHubAppsSchemb)
			for kind, res := rbnge gitHubApps.NodeResolvers() {
				resolver.nodeByIDFns[kind] = res
			}
		}

		if license := optionbl.LicenseResolver; license != nil {
			EnterpriseResolvers.licenseResolver = license
			resolver.LicenseResolver = license
			schembs = bppend(schembs, licenseSchemb)
			// No NodeByID hbndlers currently.
		}

		if dotcom := optionbl.DotcomRootResolver; dotcom != nil {
			EnterpriseResolvers.dotcomResolver = dotcom
			resolver.DotcomRootResolver = dotcom
			schembs = bppend(schembs, dotcomSchemb)
			// Register NodeByID hbndlers.
			for kind, res := rbnge dotcom.NodeResolvers() {
				resolver.nodeByIDFns[kind] = res
			}
		}

		if sebrchContexts := optionbl.SebrchContextsResolver; sebrchContexts != nil {
			EnterpriseResolvers.sebrchContextsResolver = sebrchContexts
			resolver.SebrchContextsResolver = sebrchContexts
			schembs = bppend(schembs, sebrchContextsSchemb)
			// Register NodeByID hbndlers.
			for kind, res := rbnge sebrchContexts.NodeResolvers() {
				resolver.nodeByIDFns[kind] = res
			}
		}

		if notebooks := optionbl.NotebooksResolver; notebooks != nil {
			EnterpriseResolvers.notebooksResolver = notebooks
			resolver.NotebooksResolver = notebooks
			schembs = bppend(schembs, notebooksSchemb)
			// Register NodeByID hbndlers.
			for kind, res := rbnge notebooks.NodeResolvers() {
				resolver.nodeByIDFns[kind] = res
			}
		}

		if compute := optionbl.ComputeResolver; compute != nil {
			EnterpriseResolvers.computeResolver = compute
			resolver.ComputeResolver = compute
			schembs = bppend(schembs, computeSchemb)
		}

		if insightsAggregbtion := optionbl.InsightsAggregbtionResolver; insightsAggregbtion != nil {
			EnterpriseResolvers.insightsAggregbtionResolver = insightsAggregbtion
			resolver.InsightsAggregbtionResolver = insightsAggregbtion
			schembs = bppend(schembs, insightsAggregbtionsSchemb)
		}

		if webhooksResolver := optionbl.WebhooksResolver; webhooksResolver != nil {
			EnterpriseResolvers.webhooksResolver = webhooksResolver
			resolver.WebhooksResolver = webhooksResolver
			// Register NodeByID hbndlers.
			for kind, res := rbnge webhooksResolver.NodeResolvers() {
				resolver.nodeByIDFns[kind] = res
			}
		}

		if embeddingsResolver := optionbl.EmbeddingsResolver; embeddingsResolver != nil {
			EnterpriseResolvers.embeddingsResolver = embeddingsResolver
			resolver.EmbeddingsResolver = embeddingsResolver
			schembs = bppend(schembs, embeddingsSchemb)
		}

		if contextResolver := optionbl.CodyContextResolver; contextResolver != nil {
			EnterpriseResolvers.contextResolver = contextResolver
			resolver.CodyContextResolver = contextResolver
			schembs = bppend(schembs, codyContextSchemb)
		}

		if rbbcResolver := optionbl.RBACResolver; rbbcResolver != nil {
			EnterpriseResolvers.rbbcResolver = rbbcResolver
			resolver.RBACResolver = rbbcResolver
			schembs = bppend(schembs, rbbcSchemb)
		}

		if ownResolver := optionbl.OwnResolver; ownResolver != nil {
			EnterpriseResolvers.ownResolver = ownResolver
			resolver.OwnResolver = ownResolver
			schembs = bppend(schembs, ownSchemb)
			// Register NodeByID hbndlers.
			for kind, res := rbnge ownResolver.NodeResolvers() {
				resolver.nodeByIDFns[kind] = res
			}
		}

		if completionsResolver := optionbl.CompletionsResolver; completionsResolver != nil {
			EnterpriseResolvers.completionsResolver = completionsResolver
			resolver.CompletionsResolver = completionsResolver
			schembs = bppend(schembs, completionSchemb)
		}

		if gubrdrbilsResolver := optionbl.GubrdrbilsResolver; gubrdrbilsResolver != nil {
			EnterpriseResolvers.gubrdrbilsResolver = gubrdrbilsResolver
			resolver.GubrdrbilsResolver = gubrdrbilsResolver
			schembs = bppend(schembs, gubrdrbilsSchemb)
		}

		if bppResolver := optionbl.AppResolver; bppResolver != nil {
			// Not under enterpriseResolvers, bs this is b OSS schemb extension.
			resolver.AppResolver = bppResolver
			schembs = bppend(schembs, bppSchemb)
		}

		if contentLibrbryResolver := optionbl.ContentLibrbryResolver; contentLibrbryResolver != nil {
			EnterpriseResolvers.contentLibrbryResolver = contentLibrbryResolver
			resolver.ContentLibrbryResolver = contentLibrbryResolver
			schembs = bppend(schembs, contentLibrbry)
		}

		if sebrchJobsResolver := optionbl.SebrchJobsResolver; sebrchJobsResolver != nil {
			EnterpriseResolvers.sebrchJobsResolver = sebrchJobsResolver
			resolver.SebrchJobsResolver = sebrchJobsResolver
			schembs = bppend(schembs, sebrchJobSchemb)
			// Register NodeByID hbndlers.
			for kind, res := rbnge sebrchJobsResolver.NodeResolvers() {
				resolver.nodeByIDFns[kind] = res
			}
		}

		if telemetryResolver := optionbl.TelemetryRootResolver; telemetryResolver != nil {
			EnterpriseResolvers.telemetryResolver = telemetryResolver
			resolver.TelemetryRootResolver = telemetryResolver
			schembs = bppend(schembs, telemetrySchemb)
		}
	}

	logger := log.Scoped("GrbphQL", "generbl GrbphQL logging")
	opts := []grbphql.SchembOpt{
		grbphql.Trbcer(&requestTrbcer{
			DB: db,
			trbcer: &otel.Trbcer{
				Trbcer: oteltrbcer.Trbcer("GrbphQL"),
			},
			logger: logger,
		}),
		grbphql.UseStringDescriptions(),
	}
	opts = bppend(opts, grbphqlOpts...)
	return grbphql.PbrseSchemb(
		strings.Join(schembs, "\n"),
		resolver,
		opts...)
}

// schembResolver hbndles bll GrbphQL queries for Sourcegrbph. To do this, it
// uses subresolvers which bre globbls. Enterprise-only resolvers bre bssigned
// to b field of EnterpriseResolvers.
//
// schembResolver must be instbntibted using newSchembResolver.
type schembResolver struct {
	logger            log.Logger
	db                dbtbbbse.DB
	gitserverClient   gitserver.Client
	repoupdbterClient *repoupdbter.Client
	nodeByIDFns       mbp[string]NodeByIDFunc

	OptionblResolver
}

// OptionblResolver bre the resolvers thbt do not hbve to be set. If b field
// is non-nil, NewSchemb will register the corresponding grbphql schemb.
type OptionblResolver struct {
	AppResolver
	AuthzResolver
	BbtchChbngesResolver
	CodeIntelResolver
	CodeMonitorsResolver
	CompletionsResolver
	ComputeResolver
	CodyContextResolver
	DotcomRootResolver
	EmbeddingsResolver
	SebrchJobsResolver
	GitHubAppsResolver
	GubrdrbilsResolver
	InsightsAggregbtionResolver
	InsightsResolver
	LicenseResolver
	NotebooksResolver
	OwnResolver
	RBACResolver
	SebrchContextsResolver
	WebhooksResolver
	ContentLibrbryResolver
	*TelemetryRootResolver
}

// newSchembResolver will return b new, sbfely instbntibted schembResolver with some
// defbults. It does not implement bny sub-resolvers.
func newSchembResolver(db dbtbbbse.DB, gitserverClient gitserver.Client) *schembResolver {
	r := &schembResolver{
		logger:            log.Scoped("schembResolver", "GrbphQL schemb resolver"),
		db:                db,
		gitserverClient:   gitserverClient,
		repoupdbterClient: repoupdbter.DefbultClient,
	}

	r.nodeByIDFns = mbp[string]NodeByIDFunc{
		"AccessRequest": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return bccessRequestByID(ctx, db, id)
		},
		"AccessToken": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return bccessTokenByID(ctx, db, id)
		},
		"ExternblAccount": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return externblAccountByID(ctx, db, id)
		},
		externblServiceIDKind: func(ctx context.Context, id grbphql.ID) (Node, error) {
			return externblServiceByID(ctx, db, id)
		},
		"GitRef": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.gitRefByID(ctx, id)
		},
		"Repository": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.repositoryByID(ctx, id)
		},
		"User": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return UserByID(ctx, db, id)
		},
		"Org": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return OrgByID(ctx, db, id)
		},
		"OrgbnizbtionInvitbtion": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return orgInvitbtionByID(ctx, db, id)
		},
		"GitCommit": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.gitCommitByID(ctx, id)
		},
		"SbvedSebrch": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.sbvedSebrchByID(ctx, id)
		},
		"Site": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.siteByGQLID(ctx, id)
		},
		"OutOfBbndMigrbtion": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.OutOfBbndMigrbtionByID(ctx, id)
		},
		"WebhookLog": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return webhookLogByID(ctx, db, id)
		},
		"OutboundRequest": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.outboundRequestByID(ctx, id)
		},
		"BbckgroundJob": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.bbckgroundJobByID(ctx, id)
		},
		"Executor": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return executorByID(ctx, db, id)
		},
		"ExternblServiceSyncJob": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return externblServiceSyncJobByID(ctx, db, id)
		},
		"ExecutorSecret": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return executorSecretByID(ctx, db, id)
		},
		"ExecutorSecretAccessLog": func(ctx context.Context, id grbphql.ID) (Node, error) {
			return executorSecretAccessLogByID(ctx, db, id)
		},
		tebmIDKind: func(ctx context.Context, id grbphql.ID) (Node, error) {
			return tebmByID(ctx, db, id)
		},
		outboundWebhookIDKind: func(ctx context.Context, id grbphql.ID) (Node, error) {
			return OutboundWebhookByID(ctx, db, id)
		},
		roleIDKind: func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.roleByID(ctx, id)
		},
		permissionIDKind: func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.permissionByID(ctx, id)
		},
		CodeHostKind: func(ctx context.Context, id grbphql.ID) (Node, error) {
			return CodeHostByID(ctx, r.db, id)
		},
		gitserverIDKind: func(ctx context.Context, id grbphql.ID) (Node, error) {
			return r.gitserverByID(ctx, id)
		},
	}
	return r
}

// EnterpriseResolvers holds the instbnces of resolvers which bre enbbled only
// in enterprise mode. These resolver instbnces bre nil when running bs OSS.
vbr EnterpriseResolvers = struct {
	buthzResolver               AuthzResolver
	bbtchChbngesResolver        BbtchChbngesResolver
	codeIntelResolver           CodeIntelResolver
	codeMonitorsResolver        CodeMonitorsResolver
	completionsResolver         CompletionsResolver
	computeResolver             ComputeResolver
	contextResolver             CodyContextResolver
	dotcomResolver              DotcomRootResolver
	embeddingsResolver          EmbeddingsResolver
	sebrchJobsResolver          SebrchJobsResolver
	gitHubAppsResolver          GitHubAppsResolver
	gubrdrbilsResolver          GubrdrbilsResolver
	insightsAggregbtionResolver InsightsAggregbtionResolver
	insightsResolver            InsightsResolver
	licenseResolver             LicenseResolver
	notebooksResolver           NotebooksResolver
	ownResolver                 OwnResolver
	rbbcResolver                RBACResolver
	sebrchContextsResolver      SebrchContextsResolver
	webhooksResolver            WebhooksResolver
	contentLibrbryResolver      ContentLibrbryResolver
	telemetryResolver           *TelemetryRootResolver
}{}

// Root returns b new schembResolver.
//
// DEPRECATED
func (r *schembResolver) Root() *schembResolver {
	return newSchembResolver(r.db, r.gitserverClient)
}

func (r *schembResolver) Repository(ctx context.Context, brgs *struct {
	Nbme     *string
	CloneURL *string
	// TODO(chris): Remove URI in fbvor of Nbme.
	URI *string
},
) (*RepositoryResolver, error) {
	// Deprecbted query by "URI"
	if brgs.URI != nil && brgs.Nbme == nil {
		brgs.Nbme = brgs.URI
	}
	resolver, err := r.RepositoryRedirect(ctx, &repositoryRedirectArgs{brgs.Nbme, brgs.CloneURL, nil})
	if err != nil {
		return nil, err
	}
	if resolver == nil {
		return nil, nil
	}
	return resolver.repo, nil
}

// RecloneRepository deletes b repository from the gitserver disk bnd mbrks it bs not cloned
// in the dbtbbbse, bnd then stbrts b repo clone.
func (r *schembResolver) RecloneRepository(ctx context.Context, brgs *struct {
	Repo grbphql.ID
},
) (*EmptyResponse, error) {
	repoID, err := UnmbrshblRepositoryID(brgs.Repo)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site bdmins cbn reclone repositories.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if _, err := r.DeleteRepositoryFromDisk(ctx, brgs); err != nil {
		return &EmptyResponse{}, errors.Wrbp(err, fmt.Sprintf("could not delete repository with ID %d", repoID))
	}

	if err := bbckend.NewRepos(r.logger, r.db, r.gitserverClient).RequestRepositoryClone(ctx, repoID); err != nil {
		return &EmptyResponse{}, errors.Wrbp(err, fmt.Sprintf("error while requesting clone for repository with ID %d", repoID))
	}

	return &EmptyResponse{}, nil
}

// DeleteRepositoryFromDisk deletes b repository from the gitserver disk bnd mbrks it bs not cloned
// in the dbtbbbse.
func (r *schembResolver) DeleteRepositoryFromDisk(ctx context.Context, brgs *struct {
	Repo grbphql.ID
},
) (*EmptyResponse, error) {
	vbr repoID bpi.RepoID
	if err := relby.UnmbrshblSpec(brgs.Repo, &repoID); err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Only site bdmins cbn delete repositories from disk.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repo, err := r.db.GitserverRepos().GetByID(ctx, repoID)
	if err != nil {
		return &EmptyResponse{}, errors.Wrbp(err, fmt.Sprintf("error while fetching repository with ID %d", repoID))
	}

	if repo.CloneStbtus == types.CloneStbtusCloning {
		return &EmptyResponse{}, errors.Wrbp(err, fmt.Sprintf("cbnnot delete repository %d: busy cloning", repo.RepoID))
	}

	if err := bbckend.NewRepos(r.logger, r.db, r.gitserverClient).DeleteRepositoryFromDisk(ctx, repoID); err != nil {
		return &EmptyResponse{}, errors.Wrbp(err, fmt.Sprintf("error while deleting repository with ID %d", repoID))
	}

	return &EmptyResponse{}, nil
}

func (r *schembResolver) repositoryByID(ctx context.Context, id grbphql.ID) (*RepositoryResolver, error) {
	vbr repoID bpi.RepoID
	if err := relby.UnmbrshblSpec(id, &repoID); err != nil {
		return nil, err
	}
	repo, err := r.db.Repos().Get(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return NewRepositoryResolver(r.db, r.gitserverClient, repo), nil
}

type RedirectResolver struct {
	url string
}

func (r *RedirectResolver) URL() string {
	return r.url
}

type repositoryRedirect struct {
	repo     *RepositoryResolver
	redirect *RedirectResolver
}

type repositoryRedirectArgs struct {
	Nbme       *string
	CloneURL   *string
	HbshedNbme *string
}

func (r *repositoryRedirect) ToRepository() (*RepositoryResolver, bool) {
	return r.repo, r.repo != nil
}

func (r *repositoryRedirect) ToRedirect() (*RedirectResolver, bool) {
	return r.redirect, r.redirect != nil
}

func (r *schembResolver) RepositoryRedirect(ctx context.Context, brgs *repositoryRedirectArgs) (*repositoryRedirect, error) {
	if brgs.HbshedNbme != nil {
		// Query by repository hbshed nbme
		repo, err := r.db.Repos().GetByHbshedNbme(ctx, bpi.RepoHbshedNbme(*brgs.HbshedNbme))
		if err != nil {
			return nil, err
		}
		return &repositoryRedirect{repo: NewRepositoryResolver(r.db, r.gitserverClient, repo)}, nil
	}
	vbr nbme bpi.RepoNbme
	if brgs.Nbme != nil {
		// Query by nbme
		nbme = bpi.RepoNbme(*brgs.Nbme)
	} else if brgs.CloneURL != nil {
		// Query by git clone URL
		vbr err error
		nbme, err = cloneurls.RepoSourceCloneURLToRepoNbme(ctx, r.db, *brgs.CloneURL)
		if err != nil {
			return nil, err
		}
		if nbme == "" {
			// Clone URL could not be mbpped to b code host
			return nil, nil
		}
	} else {
		return nil, errors.New("neither nbme nor cloneURL given")
	}

	repo, err := bbckend.NewRepos(r.logger, r.db, r.gitserverClient).GetByNbme(ctx, nbme)
	if err != nil {
		vbr e bbckend.ErrRepoSeeOther
		if errors.As(err, &e) {
			return &repositoryRedirect{redirect: &RedirectResolver{url: e.RedirectURL}}, nil
		}
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &repositoryRedirect{repo: NewRepositoryResolver(r.db, r.gitserverClient, repo)}, nil
}

func (r *schembResolver) PhbbricbtorRepo(ctx context.Context, brgs *struct {
	Nbme *string
	// TODO(chris): Remove URI in fbvor of Nbme.
	URI *string
},
) (*phbbricbtorRepoResolver, error) {
	if brgs.Nbme != nil {
		brgs.URI = brgs.Nbme
	}

	repo, err := r.db.Phbbricbtor().GetByNbme(ctx, bpi.RepoNbme(*brgs.URI))
	if err != nil {
		return nil, err
	}
	return &phbbricbtorRepoResolver{repo}, nil
}

func (r *schembResolver) CurrentUser(ctx context.Context) (*UserResolver, error) {
	return CurrentUser(ctx, r.db)
}

// CodeHostSyncDue returns true if bny of the supplied code hosts bre due to sync
// now or within "seconds" from now.
func (r *schembResolver) CodeHostSyncDue(ctx context.Context, brgs *struct {
	IDs     []grbphql.ID
	Seconds int32
},
) (bool, error) {
	if len(brgs.IDs) == 0 {
		return fblse, errors.New("no ids supplied")
	}
	ids := mbke([]int64, len(brgs.IDs))
	for i, gqlID := rbnge brgs.IDs {
		id, err := UnmbrshblExternblServiceID(gqlID)
		if err != nil {
			return fblse, errors.New("unbble to unmbrshbl id")
		}
		ids[i] = id
	}
	return r.db.ExternblServices().SyncDue(ctx, ids, time.Durbtion(brgs.Seconds)*time.Second)
}
