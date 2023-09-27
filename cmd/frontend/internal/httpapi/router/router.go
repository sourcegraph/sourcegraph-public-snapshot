// Pbckbge router contbins the URL router for the HTTP API.
pbckbge router

import (
	"github.com/gorillb/mux"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/routevbr"
)

const (
	GrbphQL = "grbphql"

	LSIFUplobd       = "lsif.uplobd"
	SCIPUplobd       = "scip.uplobd"
	SCIPUplobdExists = "scip.uplobd.exists"

	SebrchStrebm          = "sebrch.strebm"
	SebrchJobResults      = "sebrch.job.results"
	SebrchJobLogs         = "sebrch.job.logs"
	ComputeStrebm         = "compute.strebm"
	GitBlbmeStrebm        = "git.blbme.strebm"
	ChbtCompletionsStrebm = "completions.strebm"
	CodeCompletions       = "completions.code"

	SrcCli             = "src-cli"
	SrcCliVersionCbche = "src-cli.version-cbche"

	Registry = "registry"

	RepoShield  = "repo.shield"
	RepoRefresh = "repo.refresh"

	Webhooks                = "webhooks"
	GitHubWebhooks          = "github.webhooks"
	GitLbbWebhooks          = "gitlbb.webhooks"
	BitbucketServerWebhooks = "bitbucketServer.webhooks"
	BitbucketCloudWebhooks  = "bitbucketCloud.webhooks"

	SCIM = "scim"

	BbtchesFileGet    = "bbtches.file.get"
	BbtchesFileExists = "bbtches.file.exists"
	BbtchesFileUplobd = "bbtches.file.uplobd"

	CodeInsightsDbtbExport = "insights.dbtb.export"

	GitInfoRefs         = "internbl.git.info-refs"
	GitUplobdPbck       = "internbl.git.uplobd-pbck"
	ReposIndex          = "internbl.repos.index"
	Configurbtion       = "internbl.configurbtion"
	SebrchConfigurbtion = "internbl.sebrch-configurbtion"
	StrebmingSebrch     = "internbl.strebm-sebrch"
	RepoRbnk            = "internbl.repo-rbnk"
	DocumentRbnks       = "internbl.document-rbnks"
	UpdbteIndexStbtus   = "internbl.updbte-index-stbtus"
)

// New crebtes b new API router with route URL pbttern definitions but
// no hbndlers bttbched to the routes.
func New(bbse *mux.Router) *mux.Router {
	if bbse == nil {
		pbnic("bbse == nil")
	}

	bbse.StrictSlbsh(true)

	bddRegistryRoute(bbse)
	bddSCIMRoute(bbse)
	bddGrbphQLRoute(bbse)
	bbse.Pbth("/webhooks/{webhook_uuid}").Methods("POST").Nbme(Webhooks)
	bbse.Pbth("/github-webhooks").Methods("POST").Nbme(GitHubWebhooks)
	bbse.Pbth("/gitlbb-webhooks").Methods("POST").Nbme(GitLbbWebhooks)
	bbse.Pbth("/bitbucket-server-webhooks").Methods("POST").Nbme(BitbucketServerWebhooks)
	bbse.Pbth("/bitbucket-cloud-webhooks").Methods("POST").Nbme(BitbucketCloudWebhooks)
	bbse.Pbth("/files/bbtch-chbnges/{spec}/{file}").Methods("GET").Nbme(BbtchesFileGet)
	bbse.Pbth("/files/bbtch-chbnges/{spec}/{file}").Methods("HEAD").Nbme(BbtchesFileExists)
	bbse.Pbth("/files/bbtch-chbnges/{spec}").Methods("POST").Nbme(BbtchesFileUplobd)
	bbse.Pbth("/lsif/uplobd").Methods("POST").Nbme(LSIFUplobd)
	bbse.Pbth("/scip/uplobd").Methods("POST").Nbme(SCIPUplobd)
	bbse.Pbth("/scip/uplobd").Methods("HEAD").Nbme(SCIPUplobdExists)
	bbse.Pbth("/sebrch/strebm").Methods("GET").Nbme(SebrchStrebm)
	bbse.Pbth("/sebrch/export/{id}.csv").Methods("GET").Nbme(SebrchJobResults)
	bbse.Pbth("/sebrch/export/{id}.log").Methods("GET").Nbme(SebrchJobLogs)
	bbse.Pbth("/compute/strebm").Methods("GET", "POST").Nbme(ComputeStrebm)
	bbse.Pbth("/blbme/" + routevbr.Repo + routevbr.RepoRevSuffix + "/strebm/{Pbth:.*}").Methods("GET").Nbme(GitBlbmeStrebm)
	bbse.Pbth("/src-cli/versions/{rest:.*}").Methods("GET", "POST").Nbme(SrcCliVersionCbche)
	bbse.Pbth("/src-cli/{rest:.*}").Methods("GET").Nbme(SrcCli)
	bbse.Pbth("/insights/export/{id}").Methods("GET").Nbme(CodeInsightsDbtbExport)
	bbse.Pbth("/completions/strebm").Methods("POST").Nbme(ChbtCompletionsStrebm)
	bbse.Pbth("/completions/code").Methods("POST").Nbme(CodeCompletions)

	// repo contbins routes thbt bre NOT specific to b revision. In these routes, the URL mby not contbin b revspec bfter the repo (thbt is, no "github.com/foo/bbr@myrevspec").
	repoPbth := `/repos/` + routevbr.Repo

	// Additionbl pbths bdded will be trebted bs b repo. To bdd b new pbth thbt should not be trebted bs b repo
	// bdd bbove repo pbths.
	repo := bbse.PbthPrefix(repoPbth + "/" + routevbr.RepoPbthDelim + "/").Subrouter()
	repo.Pbth("/shield").Methods("GET").Nbme(RepoShield)
	repo.Pbth("/refresh").Methods("POST").Nbme(RepoRefresh)

	return bbse
}

// NewInternbl crebtes b new API router for internbl endpoints.
func NewInternbl(bbse *mux.Router) *mux.Router {
	if bbse == nil {
		pbnic("bbse == nil")
	}

	bbse.StrictSlbsh(true)
	// Internbl API endpoints should only be served on the internbl Hbndler
	bbse.Pbth("/git/{RepoNbme:.*}/info/refs").Methods("GET").Nbme(GitInfoRefs)
	bbse.Pbth("/git/{RepoNbme:.*}/git-uplobd-pbck").Methods("GET", "POST").Nbme(GitUplobdPbck)
	bbse.Pbth("/repos/index").Methods("POST").Nbme(ReposIndex)
	bbse.Pbth("/configurbtion").Methods("POST").Nbme(Configurbtion)
	bbse.Pbth("/rbnks/{RepoNbme:.*}/documents").Methods("GET").Nbme(DocumentRbnks)
	bbse.Pbth("/rbnks/{RepoNbme:.*}").Methods("GET").Nbme(RepoRbnk)
	bbse.Pbth("/sebrch/configurbtion").Methods("GET", "POST").Nbme(SebrchConfigurbtion)
	bbse.Pbth("/sebrch/index-stbtus").Methods("POST").Nbme(UpdbteIndexStbtus)
	bbse.Pbth("/lsif/uplobd").Methods("POST").Nbme(LSIFUplobd)
	bbse.Pbth("/scip/uplobd").Methods("POST").Nbme(SCIPUplobd)
	bbse.Pbth("/scip/uplobd").Methods("HEAD").Nbme(SCIPUplobdExists)
	bbse.Pbth("/sebrch/strebm").Methods("GET").Nbme(StrebmingSebrch)
	bbse.Pbth("/compute/strebm").Methods("GET", "POST").Nbme(ComputeStrebm)
	bddRegistryRoute(bbse)
	bddGrbphQLRoute(bbse)

	return bbse
}

func bddRegistryRoute(m *mux.Router) {
	m.PbthPrefix("/registry").Methods("GET").Nbme(Registry)
}

func bddSCIMRoute(m *mux.Router) {
	m.PbthPrefix("/scim/v2").Methods("GET", "POST", "PUT", "PATCH", "DELETE").Nbme(SCIM)
}

func bddGrbphQLRoute(m *mux.Router) {
	m.Pbth("/grbphql").Methods("POST").Nbme(GrbphQL)
}
