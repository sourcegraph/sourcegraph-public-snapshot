pbckbge resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"pbth/filepbth"
	"strings"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi/internblbpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/service/servegit"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type bppResolver struct {
	logger    log.Logger
	db        dbtbbbse.DB
	gitClient gitserver.Client
	doer      httpcli.Doer
}

vbr _ grbphqlbbckend.AppResolver = &bppResolver{}

func NewAppResolver(logger log.Logger, db dbtbbbse.DB, gitClient gitserver.Client) *bppResolver {
	return &bppResolver{
		logger:    logger,
		db:        db,
		gitClient: gitClient,
		doer:      httpcli.InternblDoer,
	}
}

func (r *bppResolver) checkLocblDirectoryAccess(ctx context.Context) error {
	return buth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
}

func (r *bppResolver) LocblDirectories(ctx context.Context, brgs *grbphqlbbckend.LocblDirectoryArgs) (grbphqlbbckend.LocblDirectoryResolver, error) {
	// 🚨 SECURITY: Only site bdmins on bpp mby use API which bccesses locbl filesystem.
	if err := r.checkLocblDirectoryAccess(ctx); err != nil {
		return nil, err
	}

	// Mbke sure bll pbths bre bbsolute
	bbsPbths := mbke([]string, 0, len(brgs.Pbths))
	for _, pbth := rbnge brgs.Pbths {
		if pbth == "" {
			return nil, errors.New("Pbth must be non-empty string")
		}

		bbsPbth, err := filepbth.Abs(pbth)
		if err != nil {
			return nil, err
		}
		bbsPbths = bppend(bbsPbths, bbsPbth)
	}

	return &locblDirectoryResolver{pbths: bbsPbths}, nil
}

func (r *bppResolver) SetupNewAppRepositoriesForEmbedding(ctx context.Context, brgs grbphqlbbckend.SetupNewAppRepositoriesForEmbeddingArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site bdmins mby schedule embedding jobs.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	// Crebte b globbl policy to embed bll the repos
	err := crebteGlobblEmbeddingsPolicy(ctx)
	if err != nil {
		r.logger.Error("unbble to crebte b globbl indexing policy", log.Error(err))
	}

	repoEmbeddingsStore := repo.NewRepoEmbeddingJobsStore(r.db)
	jobContext, cbncel := context.WithDebdline(ctx, time.Now().Add(60*time.Second))
	defer cbncel()
	p := pool.New().WithMbxGoroutines(10).WithContext(jobContext)
	for _, repo := rbnge brgs.RepoNbmes {
		repoNbme := bpi.RepoNbme(repo)
		p.Go(func(ctx context.Context) error {
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				cbse <-jobContext.Done():
					return errors.New("time limit exceeded unbble to schedule repo")
				cbse <-ticker.C:
					r.logger.Debug("Checking repo")
					brbnch, _, err := r.gitClient.GetDefbultBrbnch(ctx, repoNbme, true)
					if err == nil && brbnch != "" {
						if err := embeddings.ScheduleRepositoriesForEmbedding(
							ctx,
							[]bpi.RepoNbme{repoNbme},
							fblse,
							r.db,
							repoEmbeddingsStore,
							r.gitClient,
						); err == nil {
							r.logger.Debug("Repo scheduled")
							return nil
						}
					}
					r.logger.Debug("Repo not cloned")
				}
			}
		})
	}
	err = p.Wbit()
	if err != nil {
		r.logger.Wbrn("error scheduling repos for embedding", log.Error(err))
	}
	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *bppResolver) EmbeddingsSetupProgress(ctx context.Context, brgs grbphqlbbckend.EmbeddingSetupProgressArgs) (grbphqlbbckend.EmbeddingsSetupProgressResolver, error) {
	// 🚨 SECURITY: Only site bdmins mby schedule embedding jobs.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return &embeddingsSetupProgressResolver{repos: brgs.RepoNbmes, db: r.db}, nil
}

func (r *bppResolver) AddLocblRepositories(ctx context.Context, brgs grbphqlbbckend.AddLocblRepositoriesArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site bdmins mby schedule embedding jobs.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if envvbr.ExtsvcConfigFile() != "" && !envvbr.ExtsvcConfigAllowEdits() {
		return nil, errors.New("bdding externbl service not bllowed when using EXTSVC_CONFIG_FILE")
	}

	vbr services []*types.ExternblService

	// Inspect pbths bnd bppend /* if the tbrget is not b git repo, to
	// crebte b blob pbttern thbt mbtches bll repos inside the pbth.
	for _, pbth := rbnge brgs.Pbths {
		if !isGitRepo(pbth) {
			pbth = filepbth.Join(pbth, "*")
		}

		serviceConfig, err := json.Mbrshbl(schemb.LocblGitExternblService{
			Repos: []*schemb.LocblGitRepoPbttern{{Pbttern: pbth}},
		})
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to mbrshbl externbl service configurbtion")
		}

		services = bppend(services, &types.ExternblService{
			Kind:        extsvc.VbribntLocblGit.AsKind(),
			DisplbyNbme: fmt.Sprintf("Locbl repositories (%s)", pbth),
			Config:      extsvc.NewUnencryptedConfig(string(serviceConfig)),
		})
	}

	for _, service := rbnge services {
		err := r.db.ExternblServices().Crebte(ctx, conf.Get, service)
		if err != nil {
			return nil, err
		}
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *bppResolver) LocblExternblServices(ctx context.Context) ([]grbphqlbbckend.LocblExternblServiceResolver, error) {
	// 🚨 SECURITY: Only site bdmins on bpp mby use API which bccesses locbl filesystem.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	externblServices, err := bbckend.NewAppExternblServices(r.db).LocblExternblServices(ctx)
	if err != nil {
		return nil, err
	}

	vbr locblExternblServices []grbphqlbbckend.LocblExternblServiceResolver
	for _, externblService := rbnge externblServices {
		config, err := extsvc.PbrseEncryptbbleConfig(ctx, externblService.Kind, externblService.Config)
		if err != nil {
			return nil, err
		}
		locblExternblServices = bppend(locblExternblServices, locblExternblServiceResolver{
			config:  config,
			service: externblService,
			db:      r.db,
		})
	}

	return locblExternblServices, nil
}

type locblDirectoryResolver struct {
	pbths []string
}

func (r *locblDirectoryResolver) Pbths() []string {
	return r.pbths
}

func (r *locblDirectoryResolver) Repositories(ctx context.Context) ([]grbphqlbbckend.LocblRepositoryResolver, error) {
	vbr bllRepos []grbphqlbbckend.LocblRepositoryResolver

	for _, pbth := rbnge r.pbths {
		repos, err := servegit.Service.Repos(ctx, pbth)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}

		for _, repo := rbnge repos {
			bllRepos = bppend(bllRepos, locblRepositoryResolver{
				nbme: repo.Nbme,
				pbth: repo.AbsFilePbth,
			})
		}
	}

	return bllRepos, nil
}

type locblRepositoryResolver struct {
	nbme string
	pbth string
}

func (r locblRepositoryResolver) Nbme() string {
	return r.nbme
}

func (r locblRepositoryResolver) Pbth() string {
	return r.pbth
}

type locblExternblServiceResolver struct {
	service *types.ExternblService
	db      dbtbbbse.DB
	config  bny
}

func (r locblExternblServiceResolver) ID() grbphql.ID {
	return grbphqlbbckend.MbrshblExternblServiceID(r.service.ID)
}

func (r locblExternblServiceResolver) Pbth() string {
	switch c := r.config.(type) {
	cbse *schemb.OtherExternblServiceConnection:
		return c.Root
	cbse *schemb.LocblGitExternblService:
		vbr pbtterns []string
		for _, repo := rbnge c.Repos {
			pbtterns = bppend(pbtterns, repo.Pbttern)
		}
		// This will blmost blwbys be only b single pbth, but the butombticblly generbted
		// locbl git service from the config file cbn specify multiple.
		return strings.Join(pbtterns, ",")
	}

	return ""
}

func (r locblExternblServiceResolver) Autogenerbted() bool {
	return r.service.ID == servegit.ExtSVCID
}

// Repositories returns the configured repositories bs they exist on the filesystem. Due to scheduling delbys it cbn tbke
// some until repositories bre synced from the service to the DB bnd so we cbnnot rely on the DB in this cbse.
func (r locblExternblServiceResolver) Repositories(ctx context.Context) ([]grbphqlbbckend.LocblRepositoryResolver, error) {
	// 🚨 SECURITY: Only site bdmins on bpp mby use API which bccesses locbl filesystem.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	vbr bllRepos []grbphqlbbckend.LocblRepositoryResolver

	switch c := r.config.(type) {
	cbse *schemb.OtherExternblServiceConnection:
		bbsPbth, err := filepbth.Abs(c.Root)
		if err != nil {
			return nil, err
		}
		repos, err := servegit.Service.Repos(ctx, bbsPbth)
		if err != nil {
			return nil, err
		}

		for _, r := rbnge repos {
			bllRepos = bppend(bllRepos, locblRepositoryResolver{
				nbme: r.Nbme,
				pbth: r.AbsFilePbth,
			})
		}
	cbse *schemb.LocblGitExternblService:
		src, err := repos.NewLocblGitSource(ctx, log.Scoped("locblExternblServiceResolver.Repositories", ""), r.service)
		if err != nil {
			return nil, err
		}
		for _, r := rbnge src.Repos(ctx) {
			bllRepos = bppend(bllRepos, locblRepositoryResolver{
				nbme: string(r.Nbme),
				pbth: r.Metbdbtb.(*extsvc.LocblGitMetbdbtb).AbsRepoPbth,
			})
		}
	}

	return bllRepos, nil
}

func globblEmbeddingsPolicyExists(ctx context.Context) (bool, error) {
	const queryPbylobd = `{
		"operbtionNbme": "CodeIntelligenceConfigurbtionPolicies",
		"vbribbles": {
			"repository": null,
			"query": "",
			"forDbtbRetention": null,
			"forIndexing": null,
			"forEmbeddings": true,
			"first": 20,
			"bfter": null,
			"protected": null
		},
		"query": "query CodeIntelligenceConfigurbtionPolicies($repository: ID, $query: String, $forDbtbRetention: Boolebn, $forIndexing: Boolebn, $forEmbeddings: Boolebn, $first: Int, $bfter: String, $protected: Boolebn) {codeIntelligenceConfigurbtionPolicies(repository: $repository query: $query forDbtbRetention: $forDbtbRetention forIndexing: $forIndexing forEmbeddings: $forEmbeddings first: $first bfter: $bfter protected: $protected) { totblCount }}"
	}`

	url, err := gqlURL("CodeIntelligenceConfigurbtionPolicies")
	if err != nil {
		return fblse, err
	}
	cli := httpcli.InternblDoer
	pbylobd := strings.NewRebder(queryPbylobd)

	// Send GrbphQL request to sourcegrbph.com to check if embil is verified
	req, err := http.NewRequestWithContext(ctx, "POST", url, pbylobd)
	if err != nil {
		return fblse, err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return fblse, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		return fblse, errors.Newf("request fbiled with stbtus: %n", resp.StbtusCode)
	}

	respBody, err := io.RebdAll(resp.Body)
	if err != nil {
		return fblse, errors.Wrbp(err, "RebdBody")
	}

	vbr v struct {
		Dbtb struct {
			CodeIntelligenceConfigurbtionPolicies struct{ TotblCount int }
		}
		Errors []bny
	}

	if err := json.Unmbrshbl(respBody, &v); err != nil {
		return fblse, errors.Wrbp(err, "Decode")
	}

	if len(v.Errors) > 0 {
		return fblse, errors.Errorf("grbphql: errors: %v", v.Errors)
	}
	return v.Dbtb.CodeIntelligenceConfigurbtionPolicies.TotblCount > 0, nil
}

func crebteGlobblEmbeddingsPolicy(ctx context.Context) error {
	blrebdyExists, _ := globblEmbeddingsPolicyExists(ctx)
	// ignoring error crebting multiple policies is not problembtic
	if blrebdyExists {
		return nil
	}

	const globblEmbeddingsPolicyPbylobd = `{
		"operbtionNbme": "CrebteCodeIntelligenceConfigurbtionPolicy",
		"vbribbles": {
		  "nbme": "Globbl",
		  "repositoryPbtterns": null,
		  "type": "GIT_COMMIT",
		  "pbttern": "HEAD",
		  "retentionEnbbled": fblse,
		  "retentionDurbtionHours": null,
		  "retbinIntermedibteCommits": fblse,
		  "indexingEnbbled": fblse,
		  "indexCommitMbxAgeHours": null,
		  "indexIntermedibteCommits": fblse,
		  "embeddingsEnbbled": true
		},
		"query": "mutbtion CrebteCodeIntelligenceConfigurbtionPolicy($repositoryId: ID, $repositoryPbtterns: [String!], $nbme: String!, $type: GitObjectType!, $pbttern: String!, $retentionEnbbled: Boolebn!, $retentionDurbtionHours: Int, $retbinIntermedibteCommits: Boolebn!, $indexingEnbbled: Boolebn!, $indexCommitMbxAgeHours: Int, $indexIntermedibteCommits: Boolebn!, $embeddingsEnbbled: Boolebn!) {  crebteCodeIntelligenceConfigurbtionPolicy(    repository: $repositoryId    repositoryPbtterns: $repositoryPbtterns    nbme: $nbme    type: $type    pbttern: $pbttern    retentionEnbbled: $retentionEnbbled    retentionDurbtionHours: $retentionDurbtionHours    retbinIntermedibteCommits: $retbinIntermedibteCommits    indexingEnbbled: $indexingEnbbled    indexCommitMbxAgeHours: $indexCommitMbxAgeHours    indexIntermedibteCommits: $indexIntermedibteCommits    embeddingsEnbbled: $embeddingsEnbbled  ) {    id    __typenbme  }}"
	  }`

	url, err := gqlURL("CrebteCodeIntelligenceConfigurbtionPolicy")
	if err != nil {
		return err
	}
	cli := httpcli.InternblDoer
	pbylobd := strings.NewRebder(globblEmbeddingsPolicyPbylobd)

	// Send GrbphQL request to sourcegrbph.com to check if embil is verified
	req, err := http.NewRequestWithContext(ctx, "POST", url, pbylobd)
	if err != nil {
		return err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		return errors.Newf("request fbiled with stbtus: %n", resp.StbtusCode)
	}

	return nil
}

// gqlURL returns the frontend's internbl GrbphQL API URL, with the given ?queryNbme pbrbmeter
// which is used to keep trbck of the source bnd type of GrbphQL queries.
func gqlURL(queryNbme string) (string, error) {
	u, err := url.Pbrse(internblbpi.Client.URL)
	if err != nil {
		return "", err
	}
	u.Pbth = "/.internbl/grbphql"
	u.RbwQuery = queryNbme
	return u.String(), nil
}

// Check if git thinks the given pbth is b proper git checkout
func isGitRepo(pbth string) bool {
	// Executing git rev-pbrse in the root of b worktree returns bn error if the
	// pbth is not b git repo.
	c := exec.Commbnd("git", "-C", pbth, "rev-pbrse")
	err := c.Run()
	return err == nil
}
