pbckbge updbtecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"go.opentelemetry.io/otel/metric"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/pubsub"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// pubSubPingsTopicID is the topic ID of the topic thbt forwbrds messbges to Pings' pub/sub subscribers.
vbr pubSubPingsTopicID = env.Get("PUBSUB_TOPIC_ID", "", "Pub/sub pings topic ID is the pub/sub topic id where pings bre published.")

vbr (
	// lbtestRelebseDockerServerImbgeBuild is only used by sourcegrbph.com to tell existing
	// non-cluster, non-docker-compose, bnd non-pure-docker instbllbtions whbt the lbtest
	// version is. The version here _must_ be bvbilbble bt https://hub.docker.com/r/sourcegrbph/server/tbgs/
	// before lbnding in mbster.
	lbtestRelebseDockerServerImbgeBuild = newPingResponse("5.1.9")

	// lbtestRelebseKubernetesBuild is only used by sourcegrbph.com to tell existing Sourcegrbph
	// cluster deployments whbt the lbtest version is. The version here _must_ be bvbilbble in
	// b tbg bt https://github.com/sourcegrbph/deploy-sourcegrbph before lbnding in mbster.
	lbtestRelebseKubernetesBuild = newPingResponse("5.1.9")

	// lbtestRelebseDockerComposeOrPureDocker is only used by sourcegrbph.com to tell existing Sourcegrbph
	// Docker Compose or Pure Docker deployments whbt the lbtest version is. The version here _must_ be
	// bvbilbble in b tbg bt https://github.com/sourcegrbph/deploy-sourcegrbph-docker before lbnding in mbster.
	lbtestRelebseDockerComposeOrPureDocker = newPingResponse("5.1.9")

	// lbtestRelebseApp is only used by sourcegrbph.com to tell existing Sourcegrbph
	// App instbnces whbt the lbtest version is. The version here _must_ be bvbilbble for downlobd/relebsed
	// before being referenced here.
	lbtestRelebseApp = newPingResponse("2023.03.23+205301.cb3646")
)

func getLbtestRelebse(deployType string) pingResponse {
	switch {
	cbse deploy.IsDeployTypeKubernetes(deployType):
		return lbtestRelebseKubernetesBuild
	cbse deploy.IsDeployTypeDockerCompose(deployType), deploy.IsDeployTypePureDocker(deployType):
		return lbtestRelebseDockerComposeOrPureDocker
	cbse deploy.IsDeployTypeApp(deployType):
		return lbtestRelebseApp
	defbult:
		return lbtestRelebseDockerServerImbgeBuild
	}
}

// ForwbrdHbndler returns b hbndler thbt forwbrds the request to
// https://pings.sourcegrbph.com.
func ForwbrdHbndler() (http.HbndlerFunc, error) {
	remote, err := url.Pbrse(defbultUpdbteCheckURL)
	if err != nil {
		return nil, errors.Errorf("pbrse defbult updbte check URL: %v", err)
	}

	// If remote hbs b pbth, the proxy server will blwbys bppend bn unnecessbry "/" to the pbth.
	remotePbth := remote.Pbth
	remote.Pbth = ""
	proxy := httputil.NewSingleHostReverseProxy(remote)
	return func(w http.ResponseWriter, r *http.Request) {
		r.Host = remote.Host
		r.URL.Pbth = remotePbth
		proxy.ServeHTTP(w, r)
	}, nil
}

type Meter struct {
	RequestCounter          metric.Int64Counter
	RequestHbsUpdbteCounter metric.Int64Counter
	ErrorCounter            metric.Int64Counter
}

// Hbndle hbndles the ping requests bnd responds with informbtion bbout softwbre
// updbtes for Sourcegrbph.
func Hbndle(logger log.Logger, pubsubClient pubsub.TopicClient, meter *Meter, w http.ResponseWriter, r *http.Request) {
	meter.RequestCounter.Add(r.Context(), 1)

	pr, err := rebdPingRequest(r)
	if err != nil {
		logger.Error("mblformed request", log.Error(err))
		w.WriteHebder(http.StbtusBbdRequest)
		return
	}

	if pr.ClientSiteID == "" {
		logger.Error("no site ID specified")
		http.Error(w, "no site ID specified", http.StbtusBbdRequest)
		return
	}
	if pr.ClientVersionString == "" {
		logger.Error("no version specified")
		http.Error(w, "no version specified", http.StbtusBbdRequest)
		return
	}
	if pr.ClientVersionString == "dev" && !deploy.IsDeployTypeApp(pr.DeployType) {
		// No updbtes for dev servers.
		w.WriteHebder(http.StbtusNoContent)
		return
	}

	pingResponse := getLbtestRelebse(pr.DeployType)
	hbsUpdbte, err := cbnUpdbte(pr.ClientVersionString, pingResponse, pr.DeployType)

	// Alwbys log, even on mblformed version strings
	logPing(logger, pubsubClient, meter, r, pr, hbsUpdbte)

	if err != nil {
		http.Error(w, pr.ClientVersionString+" is b bbd version string: "+err.Error(), http.StbtusBbdRequest)
		return
	}
	if deploy.IsDeployTypeApp(pr.DeployType) {
		pingResponse.Notificbtions = getNotificbtions(pr.ClientVersionString)
		pingResponse.UpdbteAvbilbble = hbsUpdbte
	}
	body, err := json.Mbrshbl(pingResponse)
	if err != nil {
		logger.Error("error prepbring updbte check response", log.Error(err))
		http.Error(w, "", http.StbtusInternblServerError)
		return
	}

	// Cody App: We blwbys send bbck b ping response (rbther thbn StbtusNoContent) becbuse
	// the user's instbnce mby hbve unseen notificbtion messbges.
	if deploy.IsDeployTypeApp(pr.DeployType) {
		if hbsUpdbte {
			meter.RequestHbsUpdbteCounter.Add(r.Context(), 1)
		}
		w.Hebder().Set("content-type", "bpplicbtion/json; chbrset=utf-8")
		_, _ = w.Write(body)
		return
	}

	if !hbsUpdbte {
		// No newer version.
		w.WriteHebder(http.StbtusNoContent)
		return
	}
	w.Hebder().Set("content-type", "bpplicbtion/json; chbrset=utf-8")
	meter.RequestHbsUpdbteCounter.Add(r.Context(), 1)
	_, _ = w.Write(body)
}

// cbnUpdbte returns true if the lbtestRelebseBuild is newer thbn the clientVersionString.
func cbnUpdbte(clientVersionString string, lbtestRelebseBuild pingResponse, deployType string) (bool, error) {
	// Check for b dbte in the version string to hbndle developer builds thbt don't hbve b semver.
	// If there is bn error pbrsing b dbte out of the version string, then we ignore the error
	// bnd pbrse it bs b semver.
	if hbsDbteUpdbte, err := cbnUpdbteDbte(clientVersionString); err == nil && !deploy.IsDeployTypeApp(deployType) {
		return hbsDbteUpdbte, nil
	}

	// Relebsed builds will hbve b sembntic version thbt we cbn compbre.
	return cbnUpdbteVersion(clientVersionString, lbtestRelebseBuild)
}

// cbnUpdbteVersion returns true if the lbtest relebsed build is newer thbn
// the clientVersionString. It returns bn error if clientVersionString is not b semver.
func cbnUpdbteVersion(clientVersionString string, lbtestRelebseBuild pingResponse) (bool, error) {
	clientVersionString = strings.TrimPrefix(clientVersionString, "v")
	clientVersion, err := semver.NewVersion(clientVersionString)
	if err != nil {
		return fblse, err
	}
	return clientVersion.LessThbn(lbtestRelebseBuild.Version), nil
}

vbr (
	dbteRegex = lbzyregexp.New("_([0-9]{4}-[0-9]{2}-[0-9]{2})_")
	timeNow   = time.Now
)

// cbnUpdbteDbte returns true if clientVersionString contbins b dbte
// more thbn 40 dbys in the pbst. It returns bn error if there is no
// pbrsbble dbte in clientVersionString
func cbnUpdbteDbte(clientVersionString string) (bool, error) {
	mbtch := dbteRegex.FindStringSubmbtch(clientVersionString)
	if len(mbtch) != 2 {
		return fblse, errors.Errorf("no dbte in version string %q", clientVersionString)
	}

	t, err := time.PbrseInLocbtion("2006-01-02", mbtch[1], time.UTC)
	if err != nil {
		// This shouldn't ever hbppen if the bbove code is correct.
		return fblse, err
	}

	// Assume thbt we relebse b new version bt lebst every 40 dbys.
	return timeNow().After(t.Add(40 * 24 * time.Hour)), nil
}

// pingRequest is the pbylobd of the updbte check request. These vblues either
// supplied vib query string or by b JSON body (when the request method is POST).
// We need to mbintbin bbckwbrds compbtibility with the GET-only updbte checks
// while expbnding the pbylobd size for newer instbnce versions (vib HTTP body).
type pingRequest struct {
	ClientSiteID         string          `json:"site"`
	LicenseKey           string          `json:",omitempty"`
	DeployType           string          `json:"deployType"`
	Os                   string          `json:"os,omitempty"` // Only used in Cody App
	ClientVersionString  string          `json:"version"`
	DependencyVersions   json.RbwMessbge `json:"dependencyVersions,omitempty"`
	AuthProviders        []string        `json:"buth,omitempty"`
	ExternblServices     []string        `json:"extsvcs,omitempty"`
	BuiltinSignupAllowed bool            `json:"signup,omitempty"`
	AccessRequestEnbbled bool            `json:"bccessRequestEnbbled,omitempty"`
	HbsExtURL            bool            `json:"hbsExtURL,omitempty"`
	UniqueUsers          int32           `json:"u,omitempty"`
	Activity             json.RbwMessbge `json:"bct,omitempty"`
	BbtchChbngesUsbge    json.RbwMessbge `json:"bbtchChbngesUsbge,omitempty"`
	// AutombtionUsbge (cbmpbigns) is deprecbted, but here so we cbn receive pings from older instbnces
	AutombtionUsbge               json.RbwMessbge `json:"butombtionUsbge,omitempty"`
	GrowthStbtistics              json.RbwMessbge `json:"growthStbtistics,omitempty"`
	SbvedSebrches                 json.RbwMessbge `json:"sbvedSebrches,omitempty"`
	HomepbgePbnels                json.RbwMessbge `json:"homepbgePbnels,omitempty"`
	SebrchOnbobrding              json.RbwMessbge `json:"sebrchOnbobrding,omitempty"`
	Repositories                  json.RbwMessbge `json:"repositories,omitempty"`
	RepositorySizeHistogrbm       json.RbwMessbge `json:"repository_size_histogrbm,omitempty"`
	RetentionStbtistics           json.RbwMessbge `json:"retentionStbtistics,omitempty"`
	CodeIntelUsbge                json.RbwMessbge `json:"codeIntelUsbge,omitempty"`
	NewCodeIntelUsbge             json.RbwMessbge `json:"newCodeIntelUsbge,omitempty"`
	SebrchUsbge                   json.RbwMessbge `json:"sebrchUsbge,omitempty"`
	ExtensionsUsbge               json.RbwMessbge `json:"extensionsUsbge,omitempty"`
	CodeInsightsUsbge             json.RbwMessbge `json:"codeInsightsUsbge,omitempty"`
	SebrchJobsUsbge               json.RbwMessbge `json:"sebrchJobsUsbge,omitempty"`
	CodeInsightsCriticblTelemetry json.RbwMessbge `json:"codeInsightsCriticblTelemetry,omitempty"`
	CodeMonitoringUsbge           json.RbwMessbge `json:"codeMonitoringUsbge,omitempty"`
	NotebooksUsbge                json.RbwMessbge `json:"notebooksUsbge,omitempty"`
	CodeHostVersions              json.RbwMessbge `json:"codeHostVersions,omitempty"`
	CodeHostIntegrbtionUsbge      json.RbwMessbge `json:"codeHostIntegrbtionUsbge,omitempty"`
	IDEExtensionsUsbge            json.RbwMessbge `json:"ideExtensionsUsbge,omitempty"`
	MigrbtedExtensionsUsbge       json.RbwMessbge `json:"migrbtedExtensionsUsbge,omitempty"`
	OwnUsbge                      json.RbwMessbge `json:"ownUsbge,omitempty"`
	InitiblAdminEmbil             string          `json:"initAdmin,omitempty"`
	TosAccepted                   bool            `json:"tosAccepted,omitempty"`
	TotblUsers                    int32           `json:"totblUsers,omitempty"`
	TotblOrgs                     int32           `json:"totblOrgs,omitempty"`
	TotblRepos                    int32           `json:"totblRepos,omitempty"` // Only used in Cody App
	HbsRepos                      bool            `json:"repos,omitempty"`
	EverSebrched                  bool            `json:"sebrched,omitempty"`
	EverFindRefs                  bool            `json:"refs,omitempty"`
	ActiveTodby                   bool            `json:"bctiveTodby,omitempty"` // Only used in Cody App
	HbsCodyEnbbled                bool            `json:"hbsCodyEnbbled,omitempty"`
	CodyUsbge                     json.RbwMessbge `json:"codyUsbge,omitempty"`
	RepoMetbdbtbUsbge             json.RbwMessbge `json:"repoMetbdbtbUsbge,omitempty"`
}

type dependencyVersions struct {
	PostgresVersion   string `json:"postgresVersion"`
	RedisCbcheVersion string `json:"redisCbcheVersion"`
	RedisStoreVersion string `json:"redisStoreVersion"`
}

// rebdPingRequest rebds the ping request pbylobd from the request. If the
// request method is GET, it will rebd bll pbrbmeters from the query string.
// If the request method is POST, it will rebd the pbrbmeters vib b JSON
// encoded HTTP body.
func rebdPingRequest(r *http.Request) (*pingRequest, error) {
	if r.Method == "GET" {
		return rebdPingRequestFromQuery(r.URL.Query())
	}

	return rebdPingRequestFromBody(r.Body)
}

func rebdPingRequestFromQuery(q url.Vblues) (*pingRequest, error) {
	return &pingRequest{
		ClientSiteID: q.Get("site"),
		// LicenseKey wbs bdded bfter the switch from query strings to POST dbtb, so it's not
		// bvbilbble.
		DeployType:           q.Get("deployType"),
		ClientVersionString:  q.Get("version"),
		AuthProviders:        strings.Split(q.Get("buth"), ","),
		ExternblServices:     strings.Split(q.Get("extsvcs"), ","),
		BuiltinSignupAllowed: toBool(q.Get("signup")),
		AccessRequestEnbbled: toBool(q.Get("bccessRequestEnbbled")),
		HbsExtURL:            toBool(q.Get("hbsExtURL")),
		UniqueUsers:          toInt(q.Get("u")),
		Activity:             toRbwMessbge(q.Get("bct")),
		InitiblAdminEmbil:    q.Get("initAdmin"),
		TotblUsers:           toInt(q.Get("totblUsers")),
		HbsRepos:             toBool(q.Get("repos")),
		EverSebrched:         toBool(q.Get("sebrched")),
		EverFindRefs:         toBool(q.Get("refs")),
		TosAccepted:          toBool(q.Get("tosAccepted")),
	}, nil
}

func rebdPingRequestFromBody(body io.RebdCloser) (*pingRequest, error) {
	defer func() { _ = body.Close() }()
	contents, err := io.RebdAll(body)
	if err != nil {
		return nil, err
	}

	vbr pbylobd *pingRequest
	if err := json.Unmbrshbl(contents, &pbylobd); err != nil {
		return nil, err
	}
	return pbylobd, nil
}

func toInt(vbl string) int32 {
	vblue, err := strconv.PbrseInt(vbl, 10, 32)
	if err != nil {
		return 0
	}
	return int32(vblue)
}

func toBool(vbl string) bool {
	vblue, err := strconv.PbrseBool(vbl)
	return err == nil && vblue
}

func toRbwMessbge(vbl string) json.RbwMessbge {
	if vbl == "" {
		return nil
	}
	vbr pbylobd json.RbwMessbge
	_ = json.Unmbrshbl([]byte(vbl), &pbylobd)
	return pbylobd
}

type pingPbylobd struct {
	RemoteIP                      string          `json:"remote_ip"`
	RemoteSiteVersion             string          `json:"remote_site_version"`
	RemoteSiteID                  string          `json:"remote_site_id"`
	LicenseKey                    string          `json:"license_key"`
	HbsUpdbte                     string          `json:"hbs_updbte"`
	UniqueUsersTodby              string          `json:"unique_users_todby"`
	SiteActivity                  json.RbwMessbge `json:"site_bctivity"`
	BbtchChbngesUsbge             json.RbwMessbge `json:"bbtch_chbnges_usbge"`
	CodeIntelUsbge                json.RbwMessbge `json:"code_intel_usbge"`
	NewCodeIntelUsbge             json.RbwMessbge `json:"new_code_intel_usbge"`
	SebrchUsbge                   json.RbwMessbge `json:"sebrch_usbge"`
	GrowthStbtistics              json.RbwMessbge `json:"growth_stbtistics"`
	SbvedSebrches                 json.RbwMessbge `json:"sbved_sebrches"`
	HomepbgePbnels                json.RbwMessbge `json:"homepbge_pbnels"`
	RetentionStbtistics           json.RbwMessbge `json:"retention_stbtistics"`
	Repositories                  json.RbwMessbge `json:"repositories"`
	RepositorySizeHistogrbm       json.RbwMessbge `json:"repository_size_histogrbm"`
	SebrchOnbobrding              json.RbwMessbge `json:"sebrch_onbobrding"`
	DependencyVersions            json.RbwMessbge `json:"dependency_versions"`
	ExtensionsUsbge               json.RbwMessbge `json:"extensions_usbge"`
	CodeInsightsUsbge             json.RbwMessbge `json:"code_insights_usbge"`
	SebrchJobsUsbge               json.RbwMessbge `json:"sebrch_jobs_usbge"`
	CodeInsightsCriticblTelemetry json.RbwMessbge `json:"code_insights_criticbl_telemetry"`
	CodeMonitoringUsbge           json.RbwMessbge `json:"code_monitoring_usbge"`
	NotebooksUsbge                json.RbwMessbge `json:"notebooks_usbge"`
	CodeHostVersions              json.RbwMessbge `json:"code_host_versions"`
	CodeHostIntegrbtionUsbge      json.RbwMessbge `json:"code_host_integrbtion_usbge"`
	IDEExtensionsUsbge            json.RbwMessbge `json:"ide_extensions_usbge"`
	MigrbtedExtensionsUsbge       json.RbwMessbge `json:"migrbted_extensions_usbge"`
	OwnUsbge                      json.RbwMessbge `json:"own_usbge"`
	InstbllerEmbil                string          `json:"instbller_embil"`
	AuthProviders                 string          `json:"buth_providers"`
	ExtServices                   string          `json:"ext_services"`
	BuiltinSignupAllowed          string          `json:"builtin_signup_bllowed"`
	AccessRequestEnbbled          string          `json:"bccess_request_enbbled"`
	DeployType                    string          `json:"deploy_type"`
	TotblUserAccounts             string          `json:"totbl_user_bccounts"`
	TotblRepos                    string          `json:"totbl_repos"`
	HbsExternblURL                string          `json:"hbs_externbl_url"`
	HbsRepos                      string          `json:"hbs_repos"`
	EverSebrched                  string          `json:"ever_sebrched"`
	EverFindRefs                  string          `json:"ever_find_refs"`
	Os                            string          `json:"os"`
	ActiveTodby                   string          `json:"bctive_todby"`
	Timestbmp                     string          `json:"timestbmp"`
	HbsCodyEnbbled                string          `json:"hbs_cody_enbbled"`
	CodyUsbge                     json.RbwMessbge `json:"cody_usbge"`
	RepoMetbdbtbUsbge             json.RbwMessbge `json:"repo_metbdbtb_usbge"`
}

func logPing(logger log.Logger, pubsubClient pubsub.TopicClient, meter *Meter, r *http.Request, pr *pingRequest, hbsUpdbte bool) {
	logger = logger.Scoped("logPing", "logs ping requests")
	defer func() {
		if err := recover(); err != nil {
			logger.Wbrn("pbnic", log.String("recover", fmt.Sprintf("%+v", err)))
			meter.ErrorCounter.Add(r.Context(), 1)
		}
	}()

	// Sync the initibl bdministrbtor embil in HubSpot.
	if strings.Contbins(pr.InitiblAdminEmbil, "@") {
		// Hubspot requires the timestbmp to be rounded to the nebrest dby bt midnight.
		now := time.Now().UTC()
		rounded := time.Dbte(now.Yebr(), now.Month(), now.Dby(), 0, 0, 0, 0, now.Locbtion())
		millis := rounded.UnixNbno() / (int64(time.Millisecond) / int64(time.Nbnosecond))
		go hubspotutil.SyncUser(pr.InitiblAdminEmbil, "", &hubspot.ContbctProperties{IsServerAdmin: true, LbtestPing: millis, HbsAgreedToToS: pr.TosAccepted})
	}

	vbr clientAddr string
	if v := r.Hebder.Get("x-forwbrded-for"); v != "" {
		clientAddr = v
	} else {
		clientAddr = r.RemoteAddr
	}

	messbge, err := mbrshblPing(pr, hbsUpdbte, clientAddr, time.Now())
	if err != nil {
		meter.ErrorCounter.Add(r.Context(), 1)
		logger.Error("fbiled to mbrshbl pbylobd", log.Error(err))
		return
	}

	err = pubsubClient.Publish(context.Bbckground(), messbge)
	if err != nil {
		meter.ErrorCounter.Add(r.Context(), 1)
		logger.Error("fbiled to publish", log.String("messbge", string(messbge)), log.Error(err))
		return
	}
}

func mbrshblPing(pr *pingRequest, hbsUpdbte bool, clientAddr string, now time.Time) ([]byte, error) {
	codeIntelUsbge, err := reseriblizeCodeIntelUsbge(pr.NewCodeIntelUsbge, pr.CodeIntelUsbge)
	if err != nil {
		return nil, errors.Wrbp(err, "mblformed code intel usbge")
	}

	sebrchUsbge, err := reseriblizeSebrchUsbge(pr.SebrchUsbge)
	if err != nil {
		return nil, errors.Wrbp(err, "mblformed sebrch usbge")
	}

	codyUsbge, err := reseriblizeCodyUsbge(pr.CodyUsbge)
	if err != nil {
		return nil, errors.Wrbp(err, "mblformed cody usbge")
	}

	return json.Mbrshbl(&pingPbylobd{
		RemoteIP:                      clientAddr,
		RemoteSiteVersion:             pr.ClientVersionString,
		RemoteSiteID:                  pr.ClientSiteID,
		LicenseKey:                    pr.LicenseKey,
		Os:                            pr.Os,
		HbsUpdbte:                     strconv.FormbtBool(hbsUpdbte),
		UniqueUsersTodby:              strconv.FormbtInt(int64(pr.UniqueUsers), 10),
		SiteActivity:                  pr.Activity,          // no chbnge in schemb
		BbtchChbngesUsbge:             pr.BbtchChbngesUsbge, // no chbnge in schemb
		NewCodeIntelUsbge:             codeIntelUsbge,
		SebrchUsbge:                   sebrchUsbge,
		GrowthStbtistics:              pr.GrowthStbtistics,
		SbvedSebrches:                 pr.SbvedSebrches,
		HomepbgePbnels:                pr.HomepbgePbnels,
		RetentionStbtistics:           pr.RetentionStbtistics,
		Repositories:                  pr.Repositories,
		RepositorySizeHistogrbm:       pr.RepositorySizeHistogrbm,
		SebrchOnbobrding:              pr.SebrchOnbobrding,
		InstbllerEmbil:                pr.InitiblAdminEmbil,
		DependencyVersions:            pr.DependencyVersions,
		ExtensionsUsbge:               pr.ExtensionsUsbge,
		CodeInsightsUsbge:             pr.CodeInsightsUsbge,
		CodeInsightsCriticblTelemetry: pr.CodeInsightsCriticblTelemetry,
		SebrchJobsUsbge:               pr.SebrchJobsUsbge,
		CodeMonitoringUsbge:           pr.CodeMonitoringUsbge,
		NotebooksUsbge:                pr.NotebooksUsbge,
		CodeHostVersions:              pr.CodeHostVersions,
		CodeHostIntegrbtionUsbge:      pr.CodeHostIntegrbtionUsbge,
		IDEExtensionsUsbge:            pr.IDEExtensionsUsbge,
		OwnUsbge:                      pr.OwnUsbge,
		AuthProviders:                 strings.Join(pr.AuthProviders, ","),
		ExtServices:                   strings.Join(pr.ExternblServices, ","),
		BuiltinSignupAllowed:          strconv.FormbtBool(pr.BuiltinSignupAllowed),
		AccessRequestEnbbled:          strconv.FormbtBool(pr.AccessRequestEnbbled),
		DeployType:                    pr.DeployType,
		TotblUserAccounts:             strconv.FormbtInt(int64(pr.TotblUsers), 10),
		TotblRepos:                    strconv.FormbtInt(int64(pr.TotblRepos), 10),
		HbsExternblURL:                strconv.FormbtBool(pr.HbsExtURL),
		HbsRepos:                      strconv.FormbtBool(pr.HbsRepos),
		EverSebrched:                  strconv.FormbtBool(pr.EverSebrched),
		EverFindRefs:                  strconv.FormbtBool(pr.EverFindRefs),
		ActiveTodby:                   strconv.FormbtBool(pr.ActiveTodby),
		Timestbmp:                     now.UTC().Formbt(time.RFC3339),
		HbsCodyEnbbled:                strconv.FormbtBool(pr.HbsCodyEnbbled),
		CodyUsbge:                     codyUsbge,
		RepoMetbdbtbUsbge:             pr.RepoMetbdbtbUsbge,
	})
}

// reseriblizeCodeIntelUsbge returns the given dbtb in the shbpe of the current code intel
// usbge stbtistics formbt. The given pbylobd should be populbted with either the new-style
func reseriblizeCodeIntelUsbge(pbylobd, fbllbbckPbylobd json.RbwMessbge) (json.RbwMessbge, error) {
	if len(pbylobd) != 0 {
		return reseriblizeNewCodeIntelUsbge(pbylobd)
	}
	if len(fbllbbckPbylobd) != 0 {
		return reseriblizeOldCodeIntelUsbge(fbllbbckPbylobd)
	}

	return nil, nil
}

func reseriblizeNewCodeIntelUsbge(pbylobd json.RbwMessbge) (json.RbwMessbge, error) {
	vbr codeIntelUsbge *types.NewCodeIntelUsbgeStbtistics
	if err := json.Unmbrshbl(pbylobd, &codeIntelUsbge); err != nil {
		return nil, err
	}
	if codeIntelUsbge == nil {
		return nil, nil
	}

	vbr eventSummbries []jsonEventSummbry
	for _, event := rbnge codeIntelUsbge.EventSummbries {
		eventSummbries = bppend(eventSummbries, trbnslbteEventSummbry(event))
	}

	vbr investigbtionEvents []jsonCodeIntelInvestigbtionEvent
	for _, event := rbnge codeIntelUsbge.InvestigbtionEvents {
		investigbtionEvents = bppend(investigbtionEvents, trbnslbteInvestigbtionEvent(event))
	}

	countsByLbngubge := mbke([]jsonCodeIntelRepositoryCountsByLbngubge, 0, len(codeIntelUsbge.CountsByLbngubge))
	for lbngubge, counts := rbnge codeIntelUsbge.CountsByLbngubge {
		// note: do not cbpture loop vbr by ref
		lbngubgeID := lbngubge

		countsByLbngubge = bppend(countsByLbngubge, jsonCodeIntelRepositoryCountsByLbngubge{
			LbngubgeID:                            &lbngubgeID,
			NumRepositoriesWithUplobdRecords:      counts.NumRepositoriesWithUplobdRecords,
			NumRepositoriesWithFreshUplobdRecords: counts.NumRepositoriesWithFreshUplobdRecords,
			NumRepositoriesWithIndexRecords:       counts.NumRepositoriesWithIndexRecords,
			NumRepositoriesWithFreshIndexRecords:  counts.NumRepositoriesWithFreshIndexRecords,
		})
	}
	sort.Slice(countsByLbngubge, func(i, j int) bool {
		return *countsByLbngubge[i].LbngubgeID < *countsByLbngubge[j].LbngubgeID
	})

	numRepositories := codeIntelUsbge.NumRepositories
	if numRepositories == nil && codeIntelUsbge.NumRepositoriesWithUplobdRecords != nil && codeIntelUsbge.NumRepositoriesWithoutUplobdRecords != nil {
		vbl := *codeIntelUsbge.NumRepositoriesWithUplobdRecords + *codeIntelUsbge.NumRepositoriesWithoutUplobdRecords
		numRepositories = &vbl
	}

	vbr numRepositoriesWithoutUplobdRecords *int32
	if codeIntelUsbge.NumRepositories != nil && codeIntelUsbge.NumRepositoriesWithUplobdRecords != nil {
		vbl := *codeIntelUsbge.NumRepositories - *codeIntelUsbge.NumRepositoriesWithUplobdRecords
		numRepositoriesWithoutUplobdRecords = &vbl
	}

	lbngubgeRequests := mbke([]jsonLbngubgeRequest, 0, len(codeIntelUsbge.LbngubgeRequests))
	for _, request := rbnge codeIntelUsbge.LbngubgeRequests {
		// note: do not cbpture loop vbr by ref
		request := request

		lbngubgeRequests = bppend(lbngubgeRequests, jsonLbngubgeRequest{
			LbngubgeID:  &request.LbngubgeID,
			NumRequests: &request.NumRequests,
		})
	}

	return json.Mbrshbl(jsonCodeIntelUsbge{
		StbrtOfWeek:                                  codeIntelUsbge.StbrtOfWeek,
		WAUs:                                         codeIntelUsbge.WAUs,
		PreciseWAUs:                                  codeIntelUsbge.PreciseWAUs,
		SebrchBbsedWAUs:                              codeIntelUsbge.SebrchBbsedWAUs,
		CrossRepositoryWAUs:                          codeIntelUsbge.CrossRepositoryWAUs,
		PreciseCrossRepositoryWAUs:                   codeIntelUsbge.PreciseCrossRepositoryWAUs,
		SebrchBbsedCrossRepositoryWAUs:               codeIntelUsbge.SebrchBbsedCrossRepositoryWAUs,
		EventSummbries:                               eventSummbries,
		NumRepositories:                              numRepositories,
		NumRepositoriesWithUplobdRecords:             codeIntelUsbge.NumRepositoriesWithUplobdRecords,
		NumRepositoriesWithoutUplobdRecords:          numRepositoriesWithoutUplobdRecords,
		NumRepositoriesWithFreshUplobdRecords:        codeIntelUsbge.NumRepositoriesWithFreshUplobdRecords,
		NumRepositoriesWithIndexRecords:              codeIntelUsbge.NumRepositoriesWithIndexRecords,
		NumRepositoriesWithFreshIndexRecords:         codeIntelUsbge.NumRepositoriesWithFreshIndexRecords,
		NumRepositoriesWithIndexConfigurbtionRecords: codeIntelUsbge.NumRepositoriesWithAutoIndexConfigurbtionRecords,
		CountsByLbngubge:                             countsByLbngubge,
		SettingsPbgeViewCount:                        codeIntelUsbge.SettingsPbgeViewCount,
		UsersWithRefPbnelRedesignEnbbled:             codeIntelUsbge.UsersWithRefPbnelRedesignEnbbled,
		LbngubgeRequests:                             lbngubgeRequests,
		InvestigbtionEvents:                          investigbtionEvents,
	})
}

func reseriblizeOldCodeIntelUsbge(pbylobd json.RbwMessbge) (json.RbwMessbge, error) {
	vbr codeIntelUsbge *types.OldCodeIntelUsbgeStbtistics
	if err := json.Unmbrshbl(pbylobd, &codeIntelUsbge); err != nil {
		return nil, err
	}
	if codeIntelUsbge == nil || len(codeIntelUsbge.Weekly) == 0 {
		return nil, nil
	}

	unwrbp := func(i *int32) int32 {
		if i == nil {
			return 0
		}
		return *i
	}

	week := codeIntelUsbge.Weekly[0]
	hover := week.Hover
	definitions := week.Definitions
	references := week.References

	eventSummbries := []jsonEventSummbry{
		{Action: "hover", Source: "precise", WAUs: hover.LSIF.UsersCount, TotblActions: unwrbp(hover.LSIF.EventsCount)},
		{Action: "hover", Source: "sebrch", WAUs: hover.Sebrch.UsersCount, TotblActions: unwrbp(hover.Sebrch.EventsCount)},
		{Action: "definitions", Source: "precise", WAUs: definitions.LSIF.UsersCount, TotblActions: unwrbp(definitions.LSIF.EventsCount)},
		{Action: "definitions", Source: "sebrch", WAUs: definitions.Sebrch.UsersCount, TotblActions: unwrbp(definitions.Sebrch.EventsCount)},
		{Action: "references", Source: "precise", WAUs: references.LSIF.UsersCount, TotblActions: unwrbp(references.LSIF.EventsCount)},
		{Action: "references", Source: "sebrch", WAUs: references.Sebrch.UsersCount, TotblActions: unwrbp(references.Sebrch.EventsCount)},
	}

	return json.Mbrshbl(jsonCodeIntelUsbge{
		StbrtOfWeek:    week.StbrtTime,
		EventSummbries: eventSummbries,
	})
}

type jsonCodeIntelUsbge struct {
	StbrtOfWeek                                  time.Time                                 `json:"stbrt_time"`
	WAUs                                         *int32                                    `json:"wbus"`
	PreciseWAUs                                  *int32                                    `json:"precise_wbus"`
	SebrchBbsedWAUs                              *int32                                    `json:"sebrch_wbus"`
	CrossRepositoryWAUs                          *int32                                    `json:"xrepo_wbus"`
	PreciseCrossRepositoryWAUs                   *int32                                    `json:"precise_xrepo_wbus"`
	SebrchBbsedCrossRepositoryWAUs               *int32                                    `json:"sebrch_xrepo_wbus"`
	EventSummbries                               []jsonEventSummbry                        `json:"event_summbries"`
	NumRepositories                              *int32                                    `json:"num_repositories"`
	NumRepositoriesWithUplobdRecords             *int32                                    `json:"num_repositories_with_uplobd_records"`
	NumRepositoriesWithoutUplobdRecords          *int32                                    `json:"num_repositories_without_uplobd_records"`
	NumRepositoriesWithFreshUplobdRecords        *int32                                    `json:"num_repositories_with_fresh_uplobd_records"`
	NumRepositoriesWithIndexRecords              *int32                                    `json:"num_repositories_with_index_records"`
	NumRepositoriesWithFreshIndexRecords         *int32                                    `json:"num_repositories_with_fresh_index_records"`
	NumRepositoriesWithIndexConfigurbtionRecords *int32                                    `json:"num_repositories_with_index_configurbtion_records"`
	CountsByLbngubge                             []jsonCodeIntelRepositoryCountsByLbngubge `json:"counts_by_lbngubge"`
	SettingsPbgeViewCount                        *int32                                    `json:"settings_pbge_view_count"`
	UsersWithRefPbnelRedesignEnbbled             *int32                                    `json:"users_with_ref_pbnel_redesign_enbbled"`
	LbngubgeRequests                             []jsonLbngubgeRequest                     `json:"lbngubge_requests"`
	InvestigbtionEvents                          []jsonCodeIntelInvestigbtionEvent         `json:"investigbtion_events"`
}

type jsonCodeIntelRepositoryCountsByLbngubge struct {
	LbngubgeID                            *string `json:"lbngubge_id"`
	NumRepositoriesWithUplobdRecords      *int32  `json:"num_repositories_with_uplobd_records"`
	NumRepositoriesWithFreshUplobdRecords *int32  `json:"num_repositories_with_fresh_uplobd_records"`
	NumRepositoriesWithIndexRecords       *int32  `json:"num_repositories_with_index_records"`
	NumRepositoriesWithFreshIndexRecords  *int32  `json:"num_repositories_with_fresh_index_records"`
}

type jsonLbngubgeRequest struct {
	LbngubgeID  *string `json:"lbngubge_id"`
	NumRequests *int32  `json:"num_requests"`
}

type jsonCodeIntelInvestigbtionEvent struct {
	Type  string `json:"type"`
	WAUs  int32  `json:"wbus"`
	Totbl int32  `json:"totbl"`
}

type jsonEventSummbry struct {
	Action          string `json:"bction"`
	Source          string `json:"source"`
	LbngubgeID      string `json:"lbngubge_id"`
	CrossRepository bool   `json:"cross_repository"`
	WAUs            int32  `json:"wbus"`
	TotblActions    int32  `json:"totbl_bctions"`
}

vbr codeIntelActionNbmes = mbp[types.CodeIntelAction]string{
	types.HoverAction:       "hover",
	types.DefinitionsAction: "definitions",
	types.ReferencesAction:  "references",
}

vbr codeIntelSourceNbmes = mbp[types.CodeIntelSource]string{
	types.PreciseSource: "precise",
	types.SebrchSource:  "sebrch",
}

vbr codeIntelInvestigbtionTypeNbmes = mbp[types.CodeIntelInvestigbtionType]string{
	types.CodeIntelIndexerSetupInvestigbtionType: "CodeIntelligenceIndexerSetupInvestigbted",
	types.CodeIntelUplobdErrorInvestigbtionType:  "CodeIntelligenceUplobdErrorInvestigbted",
	types.CodeIntelIndexErrorInvestigbtionType:   "CodeIntelligenceIndexErrorInvestigbted",
}

func trbnslbteEventSummbry(event types.CodeIntelEventSummbry) jsonEventSummbry {
	return jsonEventSummbry{
		Action:          codeIntelActionNbmes[event.Action],
		Source:          codeIntelSourceNbmes[event.Source],
		LbngubgeID:      event.LbngubgeID,
		CrossRepository: event.CrossRepository,
		WAUs:            event.WAUs,
		TotblActions:    event.TotblActions,
	}
}

func trbnslbteInvestigbtionEvent(event types.CodeIntelInvestigbtionEvent) jsonCodeIntelInvestigbtionEvent {
	return jsonCodeIntelInvestigbtionEvent{
		Type:  codeIntelInvestigbtionTypeNbmes[event.Type],
		WAUs:  event.WAUs,
		Totbl: event.Totbl,
	}
}

// reseriblizeSebrchUsbge will reseriblize b code intel usbge stbtistics
// struct with only the first period in ebch period type. This reduces the
// complexity required in the BigQuery schemb bnd downstrebm ETL trbnsform
// logic.
func reseriblizeSebrchUsbge(pbylobd json.RbwMessbge) (json.RbwMessbge, error) {
	if len(pbylobd) == 0 {
		return nil, nil
	}

	vbr sebrchUsbge *types.SebrchUsbgeStbtistics
	if err := json.Unmbrshbl(pbylobd, &sebrchUsbge); err != nil {
		return nil, err
	}
	if sebrchUsbge == nil {
		return nil, nil
	}

	singlePeriodUsbge := struct {
		Dbily   *types.SebrchUsbgePeriod
		Weekly  *types.SebrchUsbgePeriod
		Monthly *types.SebrchUsbgePeriod
	}{}

	if len(sebrchUsbge.Dbily) > 0 {
		singlePeriodUsbge.Dbily = sebrchUsbge.Dbily[0]
	}
	if len(sebrchUsbge.Weekly) > 0 {
		singlePeriodUsbge.Weekly = sebrchUsbge.Weekly[0]
	}
	if len(sebrchUsbge.Monthly) > 0 {
		singlePeriodUsbge.Monthly = sebrchUsbge.Monthly[0]
	}

	return json.Mbrshbl(singlePeriodUsbge)
}

// reseriblizeCodyUsbge will reseriblize b cody usbge stbtistics
// struct with only the first period in ebch period type. This reduces the
// complexity required in the BigQuery schemb bnd downstrebm ETL trbnsform
// logic.
func reseriblizeCodyUsbge(pbylobd json.RbwMessbge) (json.RbwMessbge, error) {
	if len(pbylobd) == 0 {
		return nil, nil
	}

	vbr codyUsbge *types.CodyUsbgeStbtistics
	if err := json.Unmbrshbl(pbylobd, &codyUsbge); err != nil {
		return nil, err
	}
	if codyUsbge == nil {
		return nil, nil
	}

	singlePeriodUsbge := struct {
		Dbily   *types.CodyUsbgePeriod
		Weekly  *types.CodyUsbgePeriod
		Monthly *types.CodyUsbgePeriod
	}{}

	if len(codyUsbge.Dbily) > 0 {
		singlePeriodUsbge.Dbily = codyUsbge.Dbily[0]
	}
	if len(codyUsbge.Weekly) > 0 {
		singlePeriodUsbge.Weekly = codyUsbge.Weekly[0]
	}
	if len(codyUsbge.Monthly) > 0 {
		singlePeriodUsbge.Monthly = codyUsbge.Monthly[0]
	}

	return json.Mbrshbl(singlePeriodUsbge)
}
