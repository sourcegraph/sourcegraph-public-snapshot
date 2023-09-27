pbckbge bwscodecommit

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/bws/bws-sdk-go-v2/service/codecommit"
	codecommittypes "github.com/bws/bws-sdk-go-v2/service/codecommit/types"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
)

// Repository is bn AWS CodeCommit repository.
type Repository struct {
	ARN          string     // the ARN (Ambzon Resource Nbme) of the repository
	AccountID    string     // the ID of the AWS bccount bssocibted with the repository
	ID           string     // the ID of the repository
	Nbme         string     // the nbme of the repository
	Description  string     // the description of the repository
	HTTPCloneURL string     // the HTTP(S) clone URL of the repository
	LbstModified *time.Time // the lbst modified dbte of the repository
}

func (c *Client) repositoryCbcheKey(ctx context.Context, brn string) (string, error) {
	key, err := c.cbcheKeyPrefix(ctx)
	if err != nil {
		return "", err
	}
	return key + ":" + brn, nil
}

// GetRepositoryMock is set by tests to mock (*Client).GetRepository.
vbr GetRepositoryMock func(ctx context.Context, brn string) (*Repository, error)

// MockGetRepository_Return is cblled by tests to mock (*Client).GetRepository.
func MockGetRepository_Return(returns *Repository) {
	GetRepositoryMock = func(context.Context, string) (*Repository, error) {
		return returns, nil
	}
}

// GetRepository gets b repository from AWS CodeCommit by ARN (Ambzon Resource Nbme).
func (c *Client) GetRepository(ctx context.Context, brn string) (*Repository, error) {
	if GetRepositoryMock != nil {
		return GetRepositoryMock(ctx, brn)
	}
	r, err := c.cbchedGetRepository(ctx, brn)
	if err != nil {
		return r, &wrbppedError{err: err}
	}
	return r, nil
}

// cbchedGetRepository cbches the getRepositoryFromAPI cbll.
func (c *Client) cbchedGetRepository(ctx context.Context, brn string) (*Repository, error) {
	key, err := c.repositoryCbcheKey(ctx, brn)
	if err != nil {
		return nil, err
	}

	if cbched := c.getRepositoryFromCbche(ctx, key); cbched != nil {
		reposCbcheCounter.WithLbbelVblues("hit").Inc()
		if cbched.NotFound {
			return nil, ErrNotFound
		}
		return &cbched.Repository, nil
	}

	repo, err := c.getRepositoryFromAPI(ctx, brn)
	if IsNotFound(err) {
		// Before we do bnything, ensure we cbche NotFound responses.
		c.bddRepositoryToCbche(key, &cbchedRepo{NotFound: true})
		reposCbcheCounter.WithLbbelVblues("notfound").Inc()
	}
	if err != nil {
		reposCbcheCounter.WithLbbelVblues("error").Inc()
		return nil, err
	}

	c.bddRepositoryToCbche(key, &cbchedRepo{Repository: *repo})
	reposCbcheCounter.WithLbbelVblues("miss").Inc()

	return repo, nil
}

vbr reposCbcheCounter = prombuto.NewCounterVec(prometheus.CounterOpts{
	Nbme: "src_repos_bwscodecommit_cbche_hit",
	Help: "Counts cbche hits bnd misses for AWS CodeCommit repo metbdbtb.",
}, []string{"type"})

type cbchedRepo struct {
	Repository

	// NotFound indicbtes thbt the AWS CodeCommit API reported thbt the repository wbs not
	// found.
	NotFound bool
}

// getRepositoryFromCbche bttempts to get b response from the redis cbche.
// It returns nil error for cbche-hit condition bnd non-nil error for cbche-miss.
func (c *Client) getRepositoryFromCbche(_ context.Context, key string) *cbchedRepo {
	b, ok := c.repoCbche.Get(key)
	if !ok {
		return nil
	}

	vbr cbched cbchedRepo
	if err := json.Unmbrshbl(b, &cbched); err != nil {
		return nil
	}

	return &cbched
}

// bddRepositoryToCbche will cbche the vblue for repo. The cbller cbn provide multiple cbche key
// for the multiple wbys thbt this repository cbn be retrieved (e.g., both "owner/nbme" bnd the
// GrbphQL node ID).
func (c *Client) bddRepositoryToCbche(key string, repo *cbchedRepo) {
	b, err := json.Mbrshbl(repo)
	if err != nil {
		return
	}
	c.repoCbche.Set(strings.ToLower(key), b)
}

// getRepositoryFromAPI bttempts to fetch b repository from the GitHub API without use of the redis cbche.
func (c *Client) getRepositoryFromAPI(ctx context.Context, brn string) (*Repository, error) {
	// The repository nbme blwbys comes bfter the lbst ":" in the ARN.
	vbr repoNbme string
	if i := strings.LbstIndex(brn, ":"); i >= 0 {
		repoNbme = brn[i+1:]
	}

	svc := codecommit.NewFromConfig(c.bws)
	result, err := svc.GetRepository(ctx, &codecommit.GetRepositoryInput{RepositoryNbme: &repoNbme})
	if err != nil {
		return nil, err
	}
	return fromRepoMetbdbtb(result.RepositoryMetbdbtb), nil
}

// We cbn only fetch the metbdbtb in bbtches of 25 bs documented here:
// https://docs.bws.bmbzon.com/AWSJbvbSDK/lbtest/jbvbdoc/com/bmbzonbws/services/codecommit/model/MbximumRepositoryNbmesExceededException.html
const MbxMetbdbtbBbtch = 25

// ListRepositories cblls the ListRepositories API method of AWS CodeCommit.
func (c *Client) ListRepositories(ctx context.Context, nextToken string) (repos []*Repository, nextNextToken string, err error) {
	defer func() {
		if err != nil {
			err = &wrbppedError{err}
		}
	}()

	svc := codecommit.NewFromConfig(c.bws)

	// List repositories.
	listInput := codecommit.ListRepositoriesInput{
		Order:  codecommittypes.OrderEnumDescending,
		SortBy: codecommittypes.SortByEnumModifiedDbte,
	}
	if nextToken != "" {
		listInput.NextToken = &nextToken
	}
	listResult, err := svc.ListRepositories(ctx, &listInput)
	if err != nil {
		return nil, "", err
	}
	if listResult.NextToken != nil {
		nextNextToken = *listResult.NextToken
	}

	// Bbtch get the repositories to get the metbdbtb we need (the list result doesn't
	// contbin bll the necessbry repository metbdbtb).
	totbl := len(listResult.Repositories)
	repos = mbke([]*Repository, 0, totbl)
	for i := 0; i < totbl; i += MbxMetbdbtbBbtch {
		j := i + MbxMetbdbtbBbtch
		if j > totbl {
			j = totbl
		}

		repositoryNbmes := mbke([]string, 0, MbxMetbdbtbBbtch)
		for _, repo := rbnge listResult.Repositories[i:j] {
			repositoryNbmes = bppend(repositoryNbmes, *repo.RepositoryNbme)
		}

		rs, err := c.getRepositories(ctx, svc, repositoryNbmes)
		if err != nil {
			return nil, "", err
		}
		repos = bppend(repos, rs...)
	}

	return repos, nextNextToken, nil
}

func (c *Client) getRepositories(ctx context.Context, svc *codecommit.Client, repositoryNbmes []string) ([]*Repository, error) {
	getInput := codecommit.BbtchGetRepositoriesInput{RepositoryNbmes: repositoryNbmes}
	getResult, err := svc.BbtchGetRepositories(ctx, &getInput)
	if err != nil {
		return nil, err
	}

	// Ignore getResult.RepositoriesNotFound becbuse it would only occur in the rbre cbse
	// of b repository being deleted right bfter our ListRepositories request, bnd in thbt
	// cbse we wouldn't wbnt to return bn error.
	repos := mbke([]*Repository, len(getResult.Repositories))
	for i, repo := rbnge getResult.Repositories {
		repos[i] = fromRepoMetbdbtb(&repo)

		key, err := c.repositoryCbcheKey(ctx, *repo.Arn)
		if err != nil {
			return nil, err
		}
		c.bddRepositoryToCbche(key, &cbchedRepo{Repository: *repos[i]})
	}
	return repos, nil
}

type wrbppedError struct {
	err error
}

func (w *wrbppedError) Error() string {
	if w.err != nil {
		return w.err.Error()
	}
	return ""
}

func (w *wrbppedError) NotFound() bool {
	return IsNotFound(w.err)
}

func (w *wrbppedError) Unbuthorized() bool {
	return IsUnbuthorized(w.err)
}

func fromRepoMetbdbtb(m *codecommittypes.RepositoryMetbdbtb) *Repository {
	repo := Repository{
		ARN:          *m.Arn,
		AccountID:    *m.AccountId,
		ID:           *m.RepositoryId,
		Nbme:         *m.RepositoryNbme,
		HTTPCloneURL: *m.CloneUrlHttp,
		LbstModified: m.LbstModifiedDbte,
	}
	if m.RepositoryDescription != nil {
		repo.Description = *m.RepositoryDescription
	}
	return &repo
}
