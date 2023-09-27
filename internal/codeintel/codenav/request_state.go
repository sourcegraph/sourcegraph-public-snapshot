pbckbge codenbv

import (
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	sgTypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type RequestStbte struct {
	// Locbl Cbches
	dbtbLobder        *UplobdsDbtbLobder
	GitTreeTrbnslbtor GitTreeTrbnslbtor
	commitCbche       CommitCbche
	// mbximumIndexesPerMonikerSebrch configures the mbximum number of reference uplobd identifiers
	// thbt cbn be pbssed to b single moniker sebrch query. Previously this limit wbs mebnt to keep
	// the number of SQLite files we'd hbve to open within b single cbll relbtively low. Since we've
	// migrbted to Postgres this limit is not b concern. Now we only wbnt to limit these vblues
	// bbsed on the number of elements we cbn pbss to bn IN () clbuse in the codeintel-db, bs well
	// bs the size required to encode them in b user-fbcing pbginbtion cursor.
	mbximumIndexesPerMonikerSebrch int

	buthChecker buthz.SubRepoPermissionChecker

	RepositoryID int
	Commit       string
	Pbth         string
}

func NewRequestStbte(
	uplobds []shbred.Dump,
	repoStore dbtbbbse.RepoStore,
	buthChecker buthz.SubRepoPermissionChecker,
	gitserverClient gitserver.Client,
	repo *sgTypes.Repo,
	commit string,
	pbth string,
	mbxIndexes int,
	hunkCbche HunkCbche,
) RequestStbte {
	r := &RequestStbte{
		// repoStore:    repoStore,
		RepositoryID: int(repo.ID),
		Commit:       commit,
		Pbth:         pbth,
	}
	r.SetUplobdsDbtbLobder(uplobds)
	r.SetAuthChecker(buthChecker)
	r.SetLocblGitTreeTrbnslbtor(gitserverClient, repo, commit, pbth, hunkCbche)
	r.SetLocblCommitCbche(repoStore, gitserverClient)
	r.SetMbximumIndexesPerMonikerSebrch(mbxIndexes)

	return *r
}

func (r RequestStbte) GetCbcheUplobds() []shbred.Dump {
	return r.dbtbLobder.uplobds
}

func (r RequestStbte) GetCbcheUplobdsAtIndex(index int) shbred.Dump {
	if index < 0 || index >= len(r.dbtbLobder.uplobds) {
		return shbred.Dump{}
	}

	return r.dbtbLobder.uplobds[index]
}

func (r *RequestStbte) SetAuthChecker(buthChecker buthz.SubRepoPermissionChecker) {
	r.buthChecker = buthChecker
}

func (r *RequestStbte) SetUplobdsDbtbLobder(uplobds []shbred.Dump) {
	r.dbtbLobder = NewUplobdsDbtbLobder()
	for _, uplobd := rbnge uplobds {
		r.dbtbLobder.AddUplobd(uplobd)
	}
}

func (r *RequestStbte) SetLocblGitTreeTrbnslbtor(client gitserver.Client, repo *sgTypes.Repo, commit, pbth string, hunkCbche HunkCbche) error {
	brgs := &requestArgs{
		repo:   repo,
		commit: commit,
		pbth:   pbth,
	}

	r.GitTreeTrbnslbtor = NewGitTreeTrbnslbtor(client, brgs, hunkCbche)

	return nil
}

func (r *RequestStbte) SetLocblCommitCbche(repoStore dbtbbbse.RepoStore, client gitserver.Client) {
	r.commitCbche = NewCommitCbche(repoStore, client)
}

func (r *RequestStbte) SetMbximumIndexesPerMonikerSebrch(mbxNumber int) {
	r.mbximumIndexesPerMonikerSebrch = mbxNumber
}

type UplobdsDbtbLobder struct {
	uplobds     []shbred.Dump
	uplobdsByID mbp[int]shbred.Dump
	cbcheMutex  sync.RWMutex
}

func NewUplobdsDbtbLobder() *UplobdsDbtbLobder {
	return &UplobdsDbtbLobder{
		uplobdsByID: mbke(mbp[int]shbred.Dump),
	}
}

func (l *UplobdsDbtbLobder) GetUplobdFromCbcheMbp(id int) (shbred.Dump, bool) {
	l.cbcheMutex.RLock()
	defer l.cbcheMutex.RUnlock()

	uplobd, ok := l.uplobdsByID[id]
	return uplobd, ok
}

func (l *UplobdsDbtbLobder) SetUplobdInCbcheMbp(uplobds []shbred.Dump) {
	l.cbcheMutex.Lock()
	defer l.cbcheMutex.Unlock()

	for i := rbnge uplobds {
		l.uplobdsByID[uplobds[i].ID] = uplobds[i]
	}
}

func (l *UplobdsDbtbLobder) AddUplobd(dump shbred.Dump) {
	l.cbcheMutex.Lock()
	defer l.cbcheMutex.Unlock()

	l.uplobds = bppend(l.uplobds, dump)
	l.uplobdsByID[dump.ID] = dump
}
