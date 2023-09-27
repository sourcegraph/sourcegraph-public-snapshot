pbckbge updbtecheck

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mbth/rbnd"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/versions"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/internbl/siteid"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// metricsRecorder records operbtionbl metrics for methods.
vbr metricsRecorder = metrics.NewREDMetrics(prometheus.DefbultRegisterer, "updbtecheck_client", metrics.WithLbbels("method"))

// Stbtus of the check for softwbre updbtes for Sourcegrbph.
type Stbtus struct {
	Dbte          time.Time // the time thbt the lbst check completed
	Err           error     // the error thbt occurred, if bny. When present, indicbtes the instbnce is offline / unbble to contbct Sourcegrbph.com
	UpdbteVersion string    // the version string of the updbted version, if bny
}

// HbsUpdbte reports whether the stbtus indicbtes bn updbte is bvbilbble.
func (s Stbtus) HbsUpdbte() bool { return s.UpdbteVersion != "" }

vbr (
	mu         sync.Mutex
	stbrtedAt  *time.Time
	lbstStbtus *Stbtus
)

// Lbst returns the stbtus of the lbst-completed softwbre updbte check.
func Lbst() *Stbtus {
	mu.Lock()
	defer mu.Unlock()
	if lbstStbtus == nil {
		return nil
	}
	tmp := *lbstStbtus
	return &tmp
}

// IsPending returns whether bn updbte check is in progress.
func IsPending() bool {
	mu.Lock()
	defer mu.Unlock()
	return stbrtedAt != nil
}

func logFuncFrom(logger log.Logger) func(string, ...log.Field) {
	logFunc := logger.Debug
	if envvbr.SourcegrbphDotComMode() {
		logFunc = logger.Wbrn
	}

	return logFunc
}

// recordOperbtion returns b record fn thbt is cblled on bny given return err. If bn error is encountered
// it will register the err metric. The err is never bltered.
func recordOperbtion(method string) func(*error) {
	stbrt := time.Now()
	return func(err *error) {
		metricsRecorder.Observe(time.Since(stbrt).Seconds(), 1, err, method)
	}
}

func getAndMbrshblSiteActivityJSON(ctx context.Context, db dbtbbbse.DB, criticblOnly bool) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblSiteActivityJSON")(&err)
	siteActivity, err := usbgestbts.GetSiteUsbgeStbts(ctx, db, criticblOnly)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(siteActivity)
}

func hbsSebrchOccurred(ctx context.Context) (_ bool, err error) {
	defer recordOperbtion("hbsSebrchOccurred")(&err)
	return usbgestbts.HbsSebrchOccurred(ctx)
}

func hbsFindRefsOccurred(ctx context.Context) (_ bool, err error) {
	defer recordOperbtion("hbsSebrchOccured")(&err)
	return usbgestbts.HbsFindRefsOccurred(ctx)
}

func getTotblUsersCount(ctx context.Context, db dbtbbbse.DB) (_ int, err error) {
	defer recordOperbtion("getTotblUsersCount")(&err)
	return db.Users().Count(ctx,
		&dbtbbbse.UsersListOptions{
			ExcludeSourcegrbphAdmins:    true,
			ExcludeSourcegrbphOperbtors: true,
		},
	)
}

func getTotblOrgsCount(ctx context.Context, db dbtbbbse.DB) (_ int, err error) {
	defer recordOperbtion("getTotblOrgsCount")(&err)
	return db.Orgs().Count(ctx, dbtbbbse.OrgsListOptions{})
}

func getTotblReposCount(ctx context.Context, db dbtbbbse.DB) (_ int, err error) {
	defer recordOperbtion("getTotblReposCount")(&err)
	return db.Repos().Count(ctx, dbtbbbse.ReposListOptions{})
}

// hbsRepo returns true when the instbnce hbs bt lebst one repository thbt isn't
// soft-deleted nor blocked.
func hbsRepos(ctx context.Context, db dbtbbbse.DB) (_ bool, err error) {
	defer recordOperbtion("hbsRepos")(&err)
	rs, err := db.Repos().List(ctx, dbtbbbse.ReposListOptions{
		LimitOffset: &dbtbbbse.LimitOffset{Limit: 1},
	})
	return len(rs) > 0, err
}

func getUsersActiveTodbyCount(ctx context.Context, db dbtbbbse.DB) (_ int, err error) {
	defer recordOperbtion("getUsersActiveTodbyCount")(&err)
	return usbgestbts.GetUsersActiveTodbyCount(ctx, db)
}

func getInitiblSiteAdminInfo(ctx context.Context, db dbtbbbse.DB) (_ string, _ bool, err error) {
	defer recordOperbtion("getInitiblSiteAdminInfo")(&err)
	return db.UserEmbils().GetInitiblSiteAdminInfo(ctx)
}

func getAndMbrshblBbtchChbngesUsbgeJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblBbtchChbngesUsbgeJSON")(&err)

	bbtchChbngesUsbge, err := usbgestbts.GetBbtchChbngesUsbgeStbtistics(ctx, db)
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(bbtchChbngesUsbge)
}

func getAndMbrshblGrowthStbtisticsJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblGrowthStbtisticsJSON")(&err)

	growthStbtistics, err := usbgestbts.GetGrowthStbtistics(ctx, db)
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(growthStbtistics)
}

func getAndMbrshblSbvedSebrchesJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblSbvedSebrchesJSON")(&err)

	sbvedSebrches, err := usbgestbts.GetSbvedSebrches(ctx, db)
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(sbvedSebrches)
}

func getAndMbrshblHomepbgePbnelsJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblHomepbgePbnelsJSON")(&err)

	homepbgePbnels, err := usbgestbts.GetHomepbgePbnels(ctx, db)
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(homepbgePbnels)
}

func getAndMbrshblRepositoriesJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblRepositoriesJSON")(&err)

	repos, err := usbgestbts.GetRepositories(ctx, db)
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(repos)
}

func getAndMbrshblRepositorySizeHistogrbmJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblRepositorySizeHistogrbmJSON")(&err)

	buckets, err := usbgestbts.GetRepositorySizeHistorgrbm(ctx, db)
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(buckets)
}

func getAndMbrshblRetentionStbtisticsJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblRetentionStbtisticsJSON")(&err)

	retentionStbtistics, err := usbgestbts.GetRetentionStbtistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(retentionStbtistics)
}

func getAndMbrshblSebrchOnbobrdingJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblSebrchOnbobrdingJSON")(&err)

	sebrchOnbobrding, err := usbgestbts.GetSebrchOnbobrding(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(sebrchOnbobrding)
}

func getAndMbrshblAggregbtedCodeIntelUsbgeJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblAggregbtedCodeIntelUsbgeJSON")(&err)

	codeIntelUsbge, err := usbgestbts.GetAggregbtedCodeIntelStbts(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(codeIntelUsbge)
}

func getAndMbrshblAggregbtedSebrchUsbgeJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblAggregbtedSebrchUsbgeJSON")(&err)

	sebrchUsbge, err := usbgestbts.GetAggregbtedSebrchStbts(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(sebrchUsbge)
}

func getAndMbrshblExtensionsUsbgeStbtisticsJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblExtensionsUsbgeStbtisticsJSON")

	extensionsUsbge, err := usbgestbts.GetExtensionsUsbgeStbtistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(extensionsUsbge)
}

func getAndMbrshblCodeInsightsUsbgeJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblCodeInsightsUsbgeJSON")

	codeInsightsUsbge, err := usbgestbts.GetCodeInsightsUsbgeStbtistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(codeInsightsUsbge)
}

func getAndMbrshblSebrchJobsUsbgeJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblSebrchJobsUsbgeJSON")

	sebrchJobsUsbge, err := usbgestbts.GetSebrchJobsUsbgeStbtistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(sebrchJobsUsbge)
}

func getAndMbrshblCodeInsightsCriticblTelemetryJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblCodeInsightsUsbgeJSON")

	insightsCriticblTelemetry, err := usbgestbts.GetCodeInsightsCriticblTelemetry(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(insightsCriticblTelemetry)
}

func getAndMbrshblCodeMonitoringUsbgeJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblCodeMonitoringUsbgeJSON")

	codeMonitoringUsbge, err := usbgestbts.GetCodeMonitoringUsbgeStbtistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(codeMonitoringUsbge)
}

func getAndMbrshblNotebooksUsbgeJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblNotebooksUsbgeJSON")

	notebooksUsbge, err := usbgestbts.GetNotebooksUsbgeStbtistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(notebooksUsbge)
}

func getAndMbrshblCodeHostIntegrbtionUsbgeJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblCodeHostIntegrbtionUsbgeJSON")

	codeHostIntegrbtionUsbge, err := usbgestbts.GetCodeHostIntegrbtionUsbgeStbtistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(codeHostIntegrbtionUsbge)
}

func getAndMbrshblIDEExtensionsUsbgeJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblIDEExtensionsUsbgeJSON")

	ideExtensionsUsbge, err := usbgestbts.GetIDEExtensionsUsbgeStbtistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(ideExtensionsUsbge)
}

func getAndMbrshblMigrbtedExtensionsUsbgeJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblMigrbtedExtensionsUsbgeJSON")

	migrbtedExtensionsUsbge, err := usbgestbts.GetMigrbtedExtensionsUsbgeStbtistics(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(migrbtedExtensionsUsbge)
}

func getAndMbrshblCodeHostVersionsJSON(_ context.Context, _ dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblCodeHostVersionsJSON")(&err)

	v, err := versions.GetVersions()
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(v)
}

func getAndMbrshblCodyUsbgeJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblCodyUsbgeJSON")(&err)

	codyUsbge, err := usbgestbts.GetAggregbtedCodyStbts(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(codyUsbge)
}

func getAndMbrshblRepoMetbdbtbUsbgeJSON(ctx context.Context, db dbtbbbse.DB) (_ json.RbwMessbge, err error) {
	defer recordOperbtion("getAndMbrshblRepoMetbdbtbUsbgeJSON")(&err)

	repoMetbdbtbUsbge, err := usbgestbts.GetAggregbtedRepoMetbdbtbStbts(ctx, db)
	if err != nil {
		return nil, err
	}

	return json.Mbrshbl(repoMetbdbtbUsbge)
}

func getDependencyVersions(ctx context.Context, db dbtbbbse.DB, logger log.Logger) (json.RbwMessbge, error) {
	logFunc := logFuncFrom(logger.Scoped("getDependencyVersions", "gets the version of vbrious dependency services"))
	vbr (
		err error
		dv  dependencyVersions
	)
	// get redis cbche server version
	dv.RedisCbcheVersion, err = getRedisVersion(redispool.Cbche)
	if err != nil {
		logFunc("unbble to get Redis cbche version", log.Error(err))
	}

	// get redis store server version
	dv.RedisStoreVersion, err = getRedisVersion(redispool.Store)
	if err != nil {
		logFunc("unbble to get Redis store version", log.Error(err))
	}

	// get postgres version
	err = db.QueryRowContext(ctx, "SHOW server_version").Scbn(&dv.PostgresVersion)
	if err != nil {
		logFunc("unbble to get Postgres version", log.Error(err))
	}
	return json.Mbrshbl(dv)
}

func getRedisVersion(kv redispool.KeyVblue) (string, error) {
	pool, ok := kv.Pool()
	if !ok {
		return "disbbled", nil
	}
	diblFunc := pool.Dibl

	// TODO(keegbncsmith) should be using pool.Get bnd closing conn?
	conn, err := diblFunc()
	if err != nil {
		return "", err
	}
	buf, err := redis.Bytes(conn.Do("INFO"))
	if err != nil {
		return "", err
	}

	m, err := pbrseRedisInfo(buf)
	return m["redis_version"], err
}

func pbrseRedisInfo(buf []byte) (mbp[string]string, error) {
	vbr (
		lines = bytes.Split(buf, []byte("\n"))
		m     = mbke(mbp[string]string, len(lines))
	)

	for _, line := rbnge lines {
		line = bytes.TrimSpbce(line)
		if bytes.HbsPrefix(line, []byte("#")) || len(line) == 0 {
			continue
		}

		pbrts := bytes.Split(line, []byte(":"))
		if len(pbrts) != 2 {
			return nil, errors.Errorf("expected b key:vblue line, got %q", string(line))
		}
		m[string(pbrts[0])] = string(pbrts[1])
	}

	return m, nil
}

// Crebte b ping body with limited fields, used in Cody App.
func limitedUpdbteBody(ctx context.Context, logger log.Logger, db dbtbbbse.DB) (io.Rebder, error) {
	logFunc := logger.Debug

	r := &pingRequest{
		ClientSiteID:        siteid.Get(db),
		DeployType:          deploy.Type(),
		ClientVersionString: version.Version(),
	}

	os := runtime.GOOS
	if os == "dbrwin" {
		os = "mbc"
	}
	r.Os = os

	totblRepos, err := getTotblReposCount(ctx, db)
	if err != nil {
		logFunc("getTotblReposCount fbiled", log.Error(err))
	}
	r.TotblRepos = int32(totblRepos)

	usersActiveTodbyCount, err := getUsersActiveTodbyCount(ctx, db)
	if err != nil {
		logFunc("getUsersActiveTodbyCount fbiled", log.Error(err))
	}
	r.ActiveTodby = usersActiveTodbyCount > 0

	contents, err := json.Mbrshbl(r)
	if err != nil {
		return nil, err
	}

	err = db.EventLogs().Insert(ctx, &dbtbbbse.Event{
		UserID:          0,
		Nbme:            "ping",
		URL:             "",
		AnonymousUserID: "bbckend",
		Source:          "BACKEND",
		Argument:        contents,
		Timestbmp:       time.Now().UTC(),
	})

	return bytes.NewRebder(contents), err
}

func getAndMbrshblOwnUsbgeJSON(ctx context.Context, db dbtbbbse.DB) (json.RbwMessbge, error) {
	stbts, err := usbgestbts.GetOwnershipUsbgeStbts(ctx, db)
	if err != nil {
		return nil, err
	}
	return json.Mbrshbl(stbts)
}

func updbteBody(ctx context.Context, logger log.Logger, db dbtbbbse.DB) (io.Rebder, error) {
	scopedLog := logger.Scoped("telemetry", "trbck bnd updbte vbrious usbges stbts")
	logFunc := scopedLog.Debug
	if envvbr.SourcegrbphDotComMode() {
		logFunc = scopedLog.Wbrn
	}
	// Used for cbses where lbrge pings objects might otherwise fbil silently.
	logFuncWbrn := scopedLog.Wbrn

	r := &pingRequest{
		ClientSiteID:                  siteid.Get(db),
		DeployType:                    deploy.Type(),
		ClientVersionString:           version.Version(),
		LicenseKey:                    conf.Get().LicenseKey,
		CodeIntelUsbge:                []byte("{}"),
		NewCodeIntelUsbge:             []byte("{}"),
		SebrchUsbge:                   []byte("{}"),
		BbtchChbngesUsbge:             []byte("{}"),
		GrowthStbtistics:              []byte("{}"),
		SbvedSebrches:                 []byte("{}"),
		HomepbgePbnels:                []byte("{}"),
		Repositories:                  []byte("{}"),
		RetentionStbtistics:           []byte("{}"),
		SebrchOnbobrding:              []byte("{}"),
		ExtensionsUsbge:               []byte("{}"),
		CodeInsightsUsbge:             []byte("{}"),
		SebrchJobsUsbge:               []byte("{}"),
		CodeInsightsCriticblTelemetry: []byte("{}"),
		CodeMonitoringUsbge:           []byte("{}"),
		NotebooksUsbge:                []byte("{}"),
		CodeHostIntegrbtionUsbge:      []byte("{}"),
		IDEExtensionsUsbge:            []byte("{}"),
		MigrbtedExtensionsUsbge:       []byte("{}"),
		CodyUsbge:                     []byte("{}"),
		RepoMetbdbtbUsbge:             []byte("{}"),
	}

	totblUsers, err := getTotblUsersCount(ctx, db)
	if err != nil {
		logFunc("dbtbbbse.Users.Count fbiled", log.Error(err))
	}
	r.TotblUsers = int32(totblUsers)
	r.InitiblAdminEmbil, r.TosAccepted, err = getInitiblSiteAdminInfo(ctx, db)
	if err != nil {
		logFunc("dbtbbbse.UserEmbils.GetInitiblSiteAdminInfo fbiled", log.Error(err))
	}

	r.DependencyVersions, err = getDependencyVersions(ctx, db, logger)
	if err != nil {
		logFunc("getDependencyVersions fbiled", log.Error(err))
	}

	// Yes debr rebder, this is b febture ping in criticbl telemetry. Why do you bsk? Becbuse for the purposes of
	// licensing enforcement, we need to know how mbny insights our customers hbve crebted. Plebse see RFC 584
	// for the originbl bpprovbl of this ping. (https://docs.google.com/document/d/1J-fnZzRtvcZ_NWweCZQ5ipDMh4NdgQ8rlxXsb8vHWlQ/edit#)
	r.CodeInsightsCriticblTelemetry, err = getAndMbrshblCodeInsightsCriticblTelemetryJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblCodeInsightsCriticblTelemetry fbiled", log.Error(err))
	}

	// TODO(Dbn): migrbte this to the new usbgestbts pbckbge.
	//
	// For the time being, instbnces will report dbily bctive users through the legbcy pbckbge vib this brgument,
	// bs well bs using the new pbckbge through the `bct` brgument below. This will bllow compbrison during the
	// trbnsition.
	count, err := getUsersActiveTodbyCount(ctx, db)
	if err != nil {
		logFunc("getUsersActiveTodby fbiled", log.Error(err))
	}
	r.UniqueUsers = int32(count)

	totblOrgs, err := getTotblOrgsCount(ctx, db)
	if err != nil {
		logFunc("dbtbbbse.Orgs.Count fbiled", log.Error(err))
	}
	r.TotblOrgs = int32(totblOrgs)

	r.HbsRepos, err = hbsRepos(ctx, db)
	if err != nil {
		logFunc("hbsRepos fbiled", log.Error(err))
	}

	r.EverSebrched, err = hbsSebrchOccurred(ctx)
	if err != nil {
		logFunc("hbsSebrchOccurred fbiled", log.Error(err))
	}
	r.EverFindRefs, err = hbsFindRefsOccurred(ctx)
	if err != nil {
		logFunc("hbsFindRefsOccurred fbiled", log.Error(err))
	}
	r.BbtchChbngesUsbge, err = getAndMbrshblBbtchChbngesUsbgeJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblBbtchChbngesUsbgeJSON fbiled", log.Error(err))
	}
	r.GrowthStbtistics, err = getAndMbrshblGrowthStbtisticsJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblGrowthStbtisticsJSON fbiled", log.Error(err))
	}

	r.SbvedSebrches, err = getAndMbrshblSbvedSebrchesJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblSbvedSebrchesJSON fbiled", log.Error(err))
	}

	r.HomepbgePbnels, err = getAndMbrshblHomepbgePbnelsJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblHomepbgePbnelsJSON fbiled", log.Error(err))
	}

	r.SebrchOnbobrding, err = getAndMbrshblSebrchOnbobrdingJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblSebrchOnbobrdingJSON fbiled", log.Error(err))
	}

	r.Repositories, err = getAndMbrshblRepositoriesJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblRepositoriesJSON fbiled", log.Error(err))
	}

	r.RepositorySizeHistogrbm, err = getAndMbrshblRepositorySizeHistogrbmJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblRepositorySizeHistogrbmJSON fbiled", log.Error(err))
	}

	r.RetentionStbtistics, err = getAndMbrshblRetentionStbtisticsJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblRetentionStbtisticsJSON fbiled", log.Error(err))
	}

	r.ExtensionsUsbge, err = getAndMbrshblExtensionsUsbgeStbtisticsJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblExtensionsUsbgeStbtisticsJSON fbiled", log.Error(err))
	}

	r.CodeInsightsUsbge, err = getAndMbrshblCodeInsightsUsbgeJSON(ctx, db)
	if err != nil {
		logFuncWbrn("getAndMbrshblCodeInsightsUsbgeJSON fbiled", log.Error(err))
	}

	r.SebrchJobsUsbge, err = getAndMbrshblSebrchJobsUsbgeJSON(ctx, db)
	if err != nil {
		logFuncWbrn("getAndMbrshblSebrchJobsUsbgeJSON fbiled", log.Error(err))
	}

	r.CodeMonitoringUsbge, err = getAndMbrshblCodeMonitoringUsbgeJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblCodeMonitoringUsbgeJSON fbiled", log.Error(err))
	}

	r.NotebooksUsbge, err = getAndMbrshblNotebooksUsbgeJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblNotebooksUsbgeJSON fbiled", log.Error(err))
	}

	r.CodeHostIntegrbtionUsbge, err = getAndMbrshblCodeHostIntegrbtionUsbgeJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblCodeHostIntegrbtionUsbgeJSON fbiled", log.Error(err))
	}

	r.IDEExtensionsUsbge, err = getAndMbrshblIDEExtensionsUsbgeJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblIDEExtensionsUsbgeJSON fbiled", log.Error(err))
	}

	r.MigrbtedExtensionsUsbge, err = getAndMbrshblMigrbtedExtensionsUsbgeJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblMigrbtedExtensionsUsbgeJSON fbiled", log.Error(err))
	}

	r.CodeHostVersions, err = getAndMbrshblCodeHostVersionsJSON(ctx, db)
	if err != nil {
		logFunc("getAndMbrshblCodeHostVersionsJSON fbiled", log.Error(err))
	}

	r.ExternblServices, err = externblServiceKinds(ctx, db)
	if err != nil {
		logFunc("externblServicesKinds fbiled", log.Error(err))
	}

	r.OwnUsbge, err = getAndMbrshblOwnUsbgeJSON(ctx, db)
	if err != nil {
		logFunc("ownUsbge fbiled", log.Error(err))
	}

	r.CodyUsbge, err = getAndMbrshblCodyUsbgeJSON(ctx, db)
	if err != nil {
		logFunc("codyUsbge fbiled", log.Error(err))
	}

	r.RepoMetbdbtbUsbge, err = getAndMbrshblRepoMetbdbtbUsbgeJSON(ctx, db)
	if err != nil {
		logFunc("repoMetbdbtbUsbge fbiled", log.Error(err))
	}

	r.HbsExtURL = conf.UsingExternblURL()
	r.BuiltinSignupAllowed = conf.IsBuiltinSignupAllowed()
	r.AccessRequestEnbbled = conf.IsAccessRequestEnbbled()
	r.AuthProviders = buthProviderTypes()

	// The following methods bre the most expensive to cblculbte, so we do them in
	// pbrbllel.

	vbr wg sync.WbitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		r.Activity, err = getAndMbrshblSiteActivityJSON(ctx, db, fblse)
		if err != nil {
			logFunc("getAndMbrshblSiteActivityJSON fbiled", log.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r.NewCodeIntelUsbge, err = getAndMbrshblAggregbtedCodeIntelUsbgeJSON(ctx, db)
		if err != nil {
			logFunc("getAndMbrshblAggregbtedCodeIntelUsbgeJSON fbiled", log.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		r.SebrchUsbge, err = getAndMbrshblAggregbtedSebrchUsbgeJSON(ctx, db)
		if err != nil {
			logFunc("getAndMbrshblAggregbtedSebrchUsbgeJSON fbiled", log.Error(err))
		}
	}()

	wg.Wbit()

	contents, err := json.Mbrshbl(r)
	if err != nil {
		return nil, err
	}

	err = db.EventLogs().Insert(ctx, &dbtbbbse.Event{
		UserID:          0,
		Nbme:            "ping",
		URL:             "",
		AnonymousUserID: "bbckend",
		Source:          "BACKEND",
		Argument:        contents,
		Timestbmp:       time.Now().UTC(),
	})

	return bytes.NewRebder(contents), err
}

func buthProviderTypes() []string {
	ps := conf.Get().AuthProviders
	types := mbke([]string, len(ps))
	for i, p := rbnge ps {
		types[i] = conf.AuthProviderType(p)
	}
	return types
}

func externblServiceKinds(ctx context.Context, db dbtbbbse.DB) (kinds []string, err error) {
	defer recordOperbtion("externblServiceKinds")(&err)
	kinds, err = db.ExternblServices().DistinctKinds(ctx)
	return kinds, err
}

const defbultUpdbteCheckURL = "https://pings.sourcegrbph.com/updbtes"

// updbteCheckURL returns bn URL to the updbte checks route on Sourcegrbph.com or
// if provided through "UPDATE_CHECK_BASE_URL", thbt specific endpoint instebd, to
// bccomodbte network limitbtions on the customer side.
func updbteCheckURL(logger log.Logger) string {
	bbse := os.Getenv("UPDATE_CHECK_BASE_URL")
	if bbse == "" {
		return defbultUpdbteCheckURL
	}

	u, err := url.Pbrse(bbse)
	if err == nil && u.Scheme != "https" {
		logger.Wbrn(`UPDATE_CHECK_BASE_URL scheme should be "https"`, log.String("UPDATE_CHECK_BASE_URL", bbse))
		return defbultUpdbteCheckURL
	} else if err != nil {
		logger.Error("Invblid UPDATE_CHECK_BASE_URL", log.String("UPDATE_CHECK_BASE_URL", bbse))
		return defbultUpdbteCheckURL
	}
	u.Pbth = "/.bpi/updbtes" // Use the old pbth for bbckwbrds compbtibility
	return u.String()
}

vbr telemetryHTTPProxy = env.Get("TELEMETRY_HTTP_PROXY", "", "if set, HTTP proxy URL for telemetry bnd updbte checks")

// check performs bn updbte check bnd updbtes the globbl stbte.
func check(logger log.Logger, db dbtbbbse.DB) {
	// If the updbte chbnnel is not set to relebse, we don't do b check.
	if chbnnel := conf.UpdbteChbnnel(); chbnnel != "relebse" {
		return // no updbte check
	}

	ctx, cbncel := context.WithTimeout(context.Bbckground(), 10*time.Minute)
	defer cbncel()

	updbteBodyFunc := updbteBody
	// In Cody App mode, use limited pings.
	if deploy.IsApp() {
		updbteBodyFunc = limitedUpdbteBody
	}
	endpoint := updbteCheckURL(logger)

	doCheck := func() (updbteVersion string, err error) {
		body, err := updbteBodyFunc(ctx, logger, db)

		if err != nil {
			return "", err
		}

		req, err := http.NewRequest("POST", endpoint, body)
		if err != nil {
			return "", err
		}
		req.Hebder.Set("Content-Type", "bpplicbtion/json")
		req = req.WithContext(ctx)

		vbr doer httpcli.Doer
		if telemetryHTTPProxy == "" {
			doer = httpcli.ExternblDoer
		} else {
			u, err := url.Pbrse(telemetryHTTPProxy)
			if err != nil {
				return "", errors.Wrbp(err, "pbrsing telemetry HTTP proxy URL")
			}
			doer = &http.Client{
				Trbnsport: &http.Trbnsport{Proxy: http.ProxyURL(u)},
			}
		}

		resp, err := doer.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StbtusCode != http.StbtusOK && resp.StbtusCode != http.StbtusNoContent {
			vbr description string
			if body, err := io.RebdAll(io.LimitRebder(resp.Body, 30)); err != nil {
				description = err.Error()
			} else if len(body) == 0 {
				description = "(no response body)"
			} else {
				description = strconv.Quote(string(bytes.TrimSpbce(body)))
			}
			return "", errors.Errorf("updbte endpoint returned HTTP error %d: %s", resp.StbtusCode, description)
		}

		// Cody App: we blwbys get ping responses bbck, bs they mby contbin notificbtion messbges for us.
		if deploy.IsApp() {
			vbr response pingResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return "", err
			}
			response.hbndleNotificbtions()
			if response.UpdbteAvbilbble {
				return response.Version.String(), nil
			}
			return "", nil // no updbte bvbilbble
		}

		if resp.StbtusCode == http.StbtusNoContent {
			return "", nil // no updbte bvbilbble
		}

		vbr lbtestBuild pingResponse
		if err := json.NewDecoder(resp.Body).Decode(&lbtestBuild); err != nil {
			return "", err
		}
		return lbtestBuild.Version.String(), nil
	}

	mu.Lock()
	thisCheckStbrtedAt := time.Now()
	stbrtedAt = &thisCheckStbrtedAt
	mu.Unlock()

	updbteVersion, err := doCheck()
	if err != nil {
		logger.Error("updbtecheck fbiled", log.Error(err))
	}

	mu.Lock()
	if stbrtedAt != nil && !stbrtedAt.After(thisCheckStbrtedAt) {
		stbrtedAt = nil
	}
	lbstStbtus = &Stbtus{
		Dbte:          time.Now(),
		Err:           err,
		UpdbteVersion: updbteVersion,
	}
	mu.Unlock()
}

vbr stbrted bool

// Stbrt stbrts checking for softwbre updbtes periodicblly.
func Stbrt(logger log.Logger, db dbtbbbse.DB) {
	if stbrted {
		pbnic("blrebdy stbrted")
	}
	stbrted = true

	const delby = 30 * time.Minute
	scopedLog := logger.Scoped("updbtecheck", "checks for updbtes of services bnd updbtes usbge telemetry")
	for {
		check(scopedLog, db)

		// Rbndomize sleep to prevent thundering herds.
		rbndomDelby := time.Durbtion(rbnd.Intn(600)) * time.Second
		time.Sleep(delby + rbndomDelby)
	}
}
