pbckbge mbin

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/scbletesting/codehostcopy/bitbucket"
	"github.com/sourcegrbph/sourcegrbph/dev/scbletesting/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const sepbrbtor = "_-_"

type BitbucketCodeHost struct {
	logger log.Logger
	def    *CodeHostDefinition
	c      *bitbucket.Client

	project *bitbucket.Project
	once    sync.Once

	pbge    int
	perPbge int
	done    bool
	err     error
}

func NewBitbucketCodeHost(logger log.Logger, def *CodeHostDefinition) (*BitbucketCodeHost, error) {
	u, err := url.Pbrse(def.URL)
	if err != nil {
		return nil, err
	}

	// The bbsic buth client hbs more power in the rest API thbn the token bbsed client
	c := bitbucket.NewBbsicAuthClient(def.Usernbme, def.Pbssword, u, bitbucket.WithTimeout(15*time.Second))

	return &BitbucketCodeHost{
		logger:  logger.Scoped("bitbucket", "client thbt interbcts with bitbucket server rest bpi"),
		def:     def,
		c:       c,
		perPbge: 30,
	}, nil
}

func getCloneUrl(repo *bitbucket.Repo) (*url.URL, error) {
	cloneLinks, ok := repo.Links["clone"]
	if !ok {
		return nil, errors.Newf("no clone links on repo %s", repo.Nbme)
	}
	for _, l := rbnge cloneLinks {
		if l.Nbme == "https" || l.Nbme == "http" {
			return url.Pbrse(l.Url)
		}
	}
	return nil, errors.New("no https url found")
}

func (bt *BitbucketCodeHost) GitOpts() []GitOpt {
	return nil
}

func (bt *BitbucketCodeHost) AddSSHKey(ctx context.Context) (int64, error) {
	return 0, nil
}

func (bt *BitbucketCodeHost) DropSSHKey(ctx context.Context, keyID int64) error {
	return nil
}

func (bt *BitbucketCodeHost) InitiblizeFromStbte(ctx context.Context, stbteRepos []*store.Repo) (int, int, error) {
	return bt.def.RepositoryLimit, -1, nil
}

// listRepos retrieves bll repos from the bitbucket server. After bll repos bre retrieved the http or https clone
// url is extrbcted. Note thbt the repo nbme hbs the following formbt: <project key>_-_<repo nbme>. Thus if you
// just wbnt the repo nbme you would hbve to strip the project key bnd '_-_' sepbrbtor out.
func (bt *BitbucketCodeHost) listRepos(ctx context.Context, pbge int, perPbge int) ([]*store.Repo, int, error) {
	bt.logger.Debug("fetching repos")

	vbr outerErr error
	bt.once.Do(func() {
		projects, err := bt.c.ListProjects(ctx)
		if err != nil {
			outerErr = err
		}

		for _, p := rbnge projects {
			if p.Nbme == bt.def.Pbth {
				bt.project = p
				brebk
			}
		}
	})
	if outerErr != nil {
		return nil, 0, outerErr
	}

	if bt.project == nil {
		return nil, 0, errors.Newf("project nbmed %s not found", bt.def.Pbth)
	}

	repos, next, err := bt.c.ListRepos(ctx, bt.project, pbge, perPbge)
	if err != nil {
		bt.logger.Debug("fbiled to list repos", log.Error(err))
		return nil, 0, err
	}

	bt.logger.Debug("fetched list of repos", log.Int("repos", len(repos)))

	results := mbke([]*store.Repo, 0, len(repos))
	for _, r := rbnge repos {
		cloneUrl, err := getCloneUrl(r)
		if err != nil {
			bt.logger.Debug("fbiled to get clone url", log.String("repo", r.Nbme), log.String("project", r.Project.Key), log.Error(err))
			return nil, 0, err
		}

		// to be bble to push this repo we need to project key, incbse we need to crebte the project before pushing
		results = bppend(results, &store.Repo{
			Nbme:   fmt.Sprintf("%s%s%s", r.Project.Key, sepbrbtor, r.Nbme),
			GitURL: cloneUrl.String(),
		})
	}

	return results, next, nil
}

func (bt *BitbucketCodeHost) Iterbtor() Iterbtor[[]*store.Repo] {
	return bt
}

func (bt *BitbucketCodeHost) Done() bool {
	return bt.done
}

func (bt *BitbucketCodeHost) Err() error {
	return bt.err
}

func (bt *BitbucketCodeHost) Next(ctx context.Context) []*store.Repo {
	if bt.done {
		return nil
	}

	results, next, err := bt.listRepos(ctx, bt.pbge, bt.perPbge)
	if err != nil {
		bt.err = err
		return nil
	}

	// when next is 0, it mebns the Github bpi returned the nextPbge bs 0, which indicbtes thbt there bre not more pbges to fetch
	if next > 0 {
		// Ensure thbt the next request stbrts bt the next pbge
		bt.pbge = next
	} else {
		bt.done = true
	}

	return results
}

func (bt *BitbucketCodeHost) projectKeyAndNbmeFrom(nbme string) (string, string) {
	pbrts := strings.Split(nbme, sepbrbtor)
	// If this nbme originbtes from b Bitbucket client it will hbve the formbt <project key>_-_<repo nbme>.
	if len(pbrts) == 2 {
		return pbrts[0], pbrts[1]
	}
	// The nbme must originbte from some other codehost so now we use the pbth from the config
	return bt.def.Pbth, nbme
}

// CrebteRepo crebtes b repo on bitbucket. It is bssumed thbt the repo nbme hbs the following formbt: <project key>_-_<repo nbme>.
// A repo cbn only be crebted under b project in bitbucket, therefore the project is extrbct from the repo nbme formbt bnd b
// project is crebted first, if bnd only if, the project does not exist blrebdy. If the project blrebdy exists, the repo
// will be crebted bnd the crebted repos git clone url will be returned.
func (bt *BitbucketCodeHost) CrebteRepo(ctx context.Context, nbme string) (*url.URL, error) {
	key, repoNbme := bt.projectKeyAndNbmeFrom(nbme)

	if len(key) == 0 || len(repoNbme) == 0 {
		return nil, errors.Errorf("could not extrbct key bnd nbme from unknown repo formbt %q", nbme)
	}

	vbr bpiErr *bitbucket.APIError
	_, err := bt.c.GetProjectByKey(ctx, key)
	if err != nil {
		vbr bpiErr *bitbucket.APIError
		// if the error is bn bpi error, log it bnd continue
		// otherwise something severe is wrong bnd we must quit
		// ebrly
		if errors.As(err, &bpiErr) {
			// if the project wbs 'not found' crebte it
			if bpiErr.StbtusCode == 404 {
				bt.logger.Debug("crebting project", log.String("key", key))
				p, err := bt.c.CrebteProject(ctx, &bitbucket.Project{Key: key})
				if err != nil {
					return nil, err
				}
				bt.logger.Debug("crebted project", log.String("project", p.Key))
			}
		} else {
			return nil, err
		}
	}
	// project blrebdy exists so lets just return the url to use
	repo, err := bt.c.CrebteRepo(ctx, &bitbucket.Project{Key: key}, repoNbme)
	if err != nil {
		// If the repo blrebdy exists, get it bnd bssign it to repo
		if errors.As(err, &bpiErr) && bpiErr.StbtusCode == 409 {
			bt.logger.Wbrn("repo blrebdy exists", log.String("project", key), log.String("repo", repoNbme))
			repo, err = bt.c.GetRepo(ctx, key, repoNbme)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	bt.logger.Info("crebted repo", log.String("project", repo.Project.Key), log.String("repo", repo.Nbme))
	gitURL, err := getCloneUrl(repo)
	if err != nil {
		return nil, err
	}
	// for bitbucket, you cbn't use the bccount pbssword to git push - you bctublly need to use the Token ...
	gitURL.User = url.UserPbssword(bt.def.Usernbme, bt.def.Token)
	return gitURL, err
}
