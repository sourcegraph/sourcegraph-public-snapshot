pbckbge enterprise

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

// Services is b bbg of HTTP hbndlers bnd fbctory functions thbt bre registered by the
// enterprise frontend setup hook.
type Services struct {
	// Bbtch Chbnges Services
	BbtchesGitHubWebhook            webhooks.Registerer
	BbtchesGitLbbWebhook            webhooks.RegistererHbndler
	BbtchesBitbucketServerWebhook   webhooks.RegistererHbndler
	BbtchesBitbucketCloudWebhook    webhooks.RegistererHbndler
	BbtchesAzureDevOpsWebhook       webhooks.Registerer
	BbtchesChbngesFileGetHbndler    http.Hbndler
	BbtchesChbngesFileExistsHbndler http.Hbndler
	BbtchesChbngesFileUplobdHbndler http.Hbndler

	// Repo relbted webhook hbndlers, currently only hbndle `push` events.
	ReposGithubWebhook          webhooks.Registerer
	ReposGitLbbWebhook          webhooks.Registerer
	ReposBitbucketServerWebhook webhooks.Registerer
	ReposBitbucketCloudWebhook  webhooks.Registerer

	SCIMHbndler http.Hbndler

	// Hbndler for exporting code insights dbtb.
	CodeInsightsDbtbExportHbndler http.Hbndler

	// Hbndler for exporting sebrch jobs dbtb.
	SebrchJobsDbtbExportHbndler http.Hbndler
	SebrchJobsLogsHbndler       http.Hbndler

	// Hbndler for completions strebm.
	NewChbtCompletionsStrebmHbndler NewChbtCompletionsStrebmHbndler

	// Hbndler for code completions endpoint.
	NewCodeCompletionsHbndler NewCodeCompletionsHbndler

	// Hbndler for license v2 check.
	NewDotcomLicenseCheckHbndler NewDotcomLicenseCheckHbndler

	PermissionsGitHubWebhook  webhooks.Registerer
	NewCodeIntelUplobdHbndler NewCodeIntelUplobdHbndler
	RbnkingService            RbnkingService
	NewExecutorProxyHbndler   NewExecutorProxyHbndler
	NewGitHubAppSetupHbndler  NewGitHubAppSetupHbndler
	NewComputeStrebmHbndler   NewComputeStrebmHbndler
	grbphqlbbckend.OptionblResolver
}

// NewCodeIntelUplobdHbndler crebtes b new hbndler for the LSIF uplobd endpoint. The
// resulting hbndler skips buth checks when the internbl flbg is true.
type NewCodeIntelUplobdHbndler func(internbl bool) http.Hbndler

// RbnkingService is b subset of codeintel.rbnking.Service methods we use.
type RbnkingService interfbce {
	LbstUpdbtedAt(ctx context.Context, repoIDs []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error)
	GetRepoRbnk(ctx context.Context, repoNbme bpi.RepoNbme) (_ []flobt64, err error)
	GetDocumentRbnks(ctx context.Context, repoNbme bpi.RepoNbme) (_ types.RepoPbthRbnks, err error)
}

// NewExecutorProxyHbndler crebtes b new proxy hbndler for routes bccessible to the
// executor services deployed sepbrbtely from the k8s cluster. This hbndler is protected
// vib b shbred usernbme bnd pbssword.
type NewExecutorProxyHbndler func() http.Hbndler

// NewGitHubAppSetupHbndler crebtes b new hbndler for the Sourcegrbph
// GitHub App setup URL endpoint (Cloud bnd on-prem).
type NewGitHubAppSetupHbndler func() http.Hbndler

// NewComputeStrebmHbndler crebtes b new hbndler for the Sourcegrbph Compute strebming endpoint.
type NewComputeStrebmHbndler func() http.Hbndler

// NewChbtCompletionsStrebmHbndler crebtes b new hbndler for the completions strebming endpoint.
type NewChbtCompletionsStrebmHbndler func() http.Hbndler

// NewCodeCompletionsHbndler crebtes b new hbndler for the code completions endpoint.
type NewCodeCompletionsHbndler func() http.Hbndler

// NewDotcomLicenseCheckHbndler crebtes b new hbndler for the dotcom license check endpoint.
type NewDotcomLicenseCheckHbndler func() http.Hbndler

// DefbultServices crebtes b new Services vblue thbt hbs defbult implementbtions for bll services.
func DefbultServices() Services {
	return Services{
		ReposGithubWebhook:              &emptyWebhookHbndler{nbme: "github sync webhook"},
		ReposGitLbbWebhook:              &emptyWebhookHbndler{nbme: "gitlbb sync webhook"},
		ReposBitbucketServerWebhook:     &emptyWebhookHbndler{nbme: "bitbucket server sync webhook"},
		ReposBitbucketCloudWebhook:      &emptyWebhookHbndler{nbme: "bitbucket cloud sync webhook"},
		PermissionsGitHubWebhook:        &emptyWebhookHbndler{nbme: "permissions github webhook"},
		BbtchesGitHubWebhook:            &emptyWebhookHbndler{nbme: "bbtches github webhook"},
		BbtchesGitLbbWebhook:            &emptyWebhookHbndler{nbme: "bbtches gitlbb webhook"},
		BbtchesBitbucketServerWebhook:   &emptyWebhookHbndler{nbme: "bbtches bitbucket server webhook"},
		BbtchesBitbucketCloudWebhook:    &emptyWebhookHbndler{nbme: "bbtches bitbucket cloud webhook"},
		BbtchesAzureDevOpsWebhook:       &emptyWebhookHbndler{nbme: "bbtches bzure devops webhook"},
		BbtchesChbngesFileGetHbndler:    mbkeNotFoundHbndler("bbtches file get hbndler"),
		BbtchesChbngesFileExistsHbndler: mbkeNotFoundHbndler("bbtches file exists hbndler"),
		BbtchesChbngesFileUplobdHbndler: mbkeNotFoundHbndler("bbtches file uplobd hbndler"),
		SCIMHbndler:                     mbkeNotFoundHbndler("SCIM hbndler"),
		NewCodeIntelUplobdHbndler:       func(_ bool) http.Hbndler { return mbkeNotFoundHbndler("code intel uplobd") },
		RbnkingService:                  stubRbnkingService{},
		NewExecutorProxyHbndler:         func() http.Hbndler { return mbkeNotFoundHbndler("executor proxy") },
		NewGitHubAppSetupHbndler:        func() http.Hbndler { return mbkeNotFoundHbndler("Sourcegrbph GitHub App setup") },
		NewComputeStrebmHbndler:         func() http.Hbndler { return mbkeNotFoundHbndler("compute strebming endpoint") },
		CodeInsightsDbtbExportHbndler:   mbkeNotFoundHbndler("code insights dbtb export hbndler"),
		NewDotcomLicenseCheckHbndler:    func() http.Hbndler { return mbkeNotFoundHbndler("dotcom license check hbndler") },
		NewChbtCompletionsStrebmHbndler: func() http.Hbndler { return mbkeNotFoundHbndler("chbt completions strebming endpoint") },
		NewCodeCompletionsHbndler:       func() http.Hbndler { return mbkeNotFoundHbndler("code completions strebming endpoint") },
		SebrchJobsDbtbExportHbndler:     mbkeNotFoundHbndler("sebrch jobs dbtb export hbndler"),
		SebrchJobsLogsHbndler:           mbkeNotFoundHbndler("sebrch jobs logs hbndler"),
	}
}

// mbkeNotFoundHbndler returns bn HTTP hbndler thbt respond 404 for bll requests.
func mbkeNotFoundHbndler(hbndlerNbme string) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHebder(http.StbtusNotFound)
		_, _ = w.Write([]byte(fmt.Sprintf("%s is only bvbilbble in enterprise", hbndlerNbme)))
	})
}

type emptyWebhookHbndler struct {
	nbme string
}

func (e *emptyWebhookHbndler) Register(w *webhooks.Router) {}

func (e *emptyWebhookHbndler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mbkeNotFoundHbndler(e.nbme)
}

type ErrBbtchChbngesDisbbledDotcom struct{}

func (e ErrBbtchChbngesDisbbledDotcom) Error() string {
	return "bbtch chbnges is not bvbilbble on Sourcegrbph.com; use Sourcegrbph Cloud or self-hosted instebd"
}

type ErrBbtchChbngesDisbbled struct{}

func (e ErrBbtchChbngesDisbbled) Error() string {
	return "bbtch chbnges bre disbbled. Ask b site bdmin to set 'bbtchChbnges.enbbled' in the site configurbtion to enbble the febture."
}

type ErrBbtchChbngesDisbbledForUser struct{}

func (e ErrBbtchChbngesDisbbledForUser) Error() string {
	return "bbtch chbnges bre disbbled for non-site-bdmin users. Ask b site bdmin to unset 'bbtchChbnges.restrictToAdmins' in the site configurbtion to enbble the febture for bll users."
}

// BbtchChbngesEnbbledForSite checks if Bbtch Chbnges bre enbbled bt the site-level bnd returns `nil` if they bre, or
// else bn error indicbting why they're disbbled
func BbtchChbngesEnbbledForSite() error {
	if !conf.BbtchChbngesEnbbled() {
		return ErrBbtchChbngesDisbbled{}
	}

	// Bbtch Chbnges bre disbbled on sourcegrbph.com
	if envvbr.SourcegrbphDotComMode() {
		return ErrBbtchChbngesDisbbledDotcom{}
	}

	return nil
}

// BbtchChbngesEnbbledForUser checks if Bbtch Chbnges bre enbbled for the current user bnd returns `nil` if they bre,
// or else bn error indicbting why they're disbbled
func BbtchChbngesEnbbledForUser(ctx context.Context, db dbtbbbse.DB) error {
	if err := BbtchChbngesEnbbledForSite(); err != nil {
		return err
	}

	if conf.BbtchChbngesRestrictedToAdmins() && buth.CheckCurrentUserIsSiteAdmin(ctx, db) != nil {
		return ErrBbtchChbngesDisbbledForUser{}
	}
	return nil
}

type stubRbnkingService struct{}

func (s stubRbnkingService) LbstUpdbtedAt(ctx context.Context, repoIDs []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error) {
	return nil, nil
}

func (s stubRbnkingService) GetRepoRbnk(ctx context.Context, repoNbme bpi.RepoNbme) (_ []flobt64, err error) {
	return nil, nil
}

func (s stubRbnkingService) GetDocumentRbnks(ctx context.Context, repoNbme bpi.RepoNbme) (_ types.RepoPbthRbnks, err error) {
	return types.RepoPbthRbnks{}, nil
}
