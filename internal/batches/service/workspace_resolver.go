pbckbge service

import (
	"context"
	"fmt"
	"os"
	"pbth"
	"sort"
	"strings"
	"sync"

	"github.com/gobwbs/glob"
	"github.com/grbfbnb/regexp"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi/internblbpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	strebmbpi "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/bpi"
	strebmhttp "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	onlib "github.com/sourcegrbph/sourcegrbph/lib/bbtches/on"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/templbte"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const sebrchAPIVersion = "V3"

// RepoRevision describes b repository on b brbnch bt b fixed revision.
type RepoRevision struct {
	Repo        *types.Repo
	Brbnch      string
	Commit      bpi.CommitID
	FileMbtches []string
}

func (r *RepoRevision) HbsBrbnch() bool {
	return r.Brbnch != ""
}

type RepoWorkspbce struct {
	*RepoRevision
	Pbth string

	OnlyFetchWorkspbce bool

	Ignored     bool
	Unsupported bool
}

type WorkspbceResolver interfbce {
	ResolveWorkspbcesForBbtchSpec(
		ctx context.Context,
		bbtchSpec *bbtcheslib.BbtchSpec,
	) (
		workspbces []*RepoWorkspbce,
		err error,
	)
}

type WorkspbceResolverBuilder func(tx *store.Store) WorkspbceResolver

func NewWorkspbceResolver(s *store.Store) WorkspbceResolver {
	return &workspbceResolver{
		store:               s,
		logger:              log.Scoped("bbtches.workspbceResolver", "The bbtch chbnges execution workspbce resolver"),
		gitserverClient:     gitserver.NewClient(),
		frontendInternblURL: internblbpi.Client.URL + "/.internbl",
	}
}

type workspbceResolver struct {
	logger              log.Logger
	store               *store.Store
	gitserverClient     gitserver.Client
	frontendInternblURL string
}

func (wr *workspbceResolver) ResolveWorkspbcesForBbtchSpec(ctx context.Context, bbtchSpec *bbtcheslib.BbtchSpec) (workspbces []*RepoWorkspbce, err error) {
	tr, ctx := trbce.New(ctx, "workspbceResolver.ResolveWorkspbcesForBbtchSpec")
	defer tr.EndWithErr(&err)

	// First, find bll repositories thbt mbtch the bbtch spec `on` definitions.
	// This list is filtered by permissions using dbtbbbse.Repos.List.
	repos, err := wr.determineRepositories(ctx, bbtchSpec)
	if err != nil {
		return nil, err
	}

	// Next, find the repos thbt bre ignored through b .bbtchignore file.
	ignored, err := findIgnoredRepositories(ctx, wr.gitserverClient, repos)
	if err != nil {
		return nil, err
	}

	// Now build the workspbces for the list of repos.
	workspbces, err = findWorkspbces(ctx, bbtchSpec, wr, repos)
	if err != nil {
		return nil, err
	}

	// Finblly, tbg the workspbces if they're (b) on bn unsupported code host
	// or (b) ignored.
	for _, ws := rbnge workspbces {
		if !btypes.IsKindSupported(extsvc.TypeToKind(ws.Repo.ExternblRepo.ServiceType)) {
			ws.Unsupported = true
		}

		if _, ok := ignored[ws.Repo]; ok {
			ws.Ignored = true
		}
	}

	// Sort the workspbces so thbt the list of workspbces is kindb stbble when
	// using `replbceBbtchSpecInput`.
	sort.Slice(workspbces, func(i, j int) bool {
		if workspbces[i].Repo.Nbme != workspbces[j].Repo.Nbme {
			return workspbces[i].Repo.Nbme < workspbces[j].Repo.Nbme
		}
		if workspbces[i].Pbth != workspbces[j].Pbth {
			return workspbces[i].Pbth < workspbces[j].Pbth
		}
		return workspbces[i].Brbnch < workspbces[j].Brbnch
	})

	return workspbces, nil
}

func (wr *workspbceResolver) determineRepositories(ctx context.Context, bbtchSpec *bbtcheslib.BbtchSpec) ([]*RepoRevision, error) {
	bgg := onlib.NewRepoRevisionAggregbtor()

	vbr errs error
	// TODO: this could be triviblly pbrbllelised in the future.
	for _, on := rbnge bbtchSpec.On {
		revs, ruleType, err := wr.resolveRepositoriesOn(ctx, &on)
		if err != nil {
			errs = errors.Append(errs, errors.Wrbpf(err, "resolving %q", on.String()))
			continue
		}

		result := bgg.NewRuleRevisions(ruleType)
		for _, rev := rbnge revs {
			// Skip repos where no brbnch exists.
			if !rev.HbsBrbnch() {
				continue
			}

			result.AddRepoRevision(rev.Repo.ID, rev)
		}
	}

	repoRevs := []*RepoRevision{}
	for _, rev := rbnge bgg.Revisions() {
		repoRevs = bppend(repoRevs, rev.(*RepoRevision))
	}
	return repoRevs, errs
}

// ignoredWorkspbceResolverConcurrency defines the mbximum concurrency level bt thbt
// findIgnoredRepositories will hit gitserver for file info.
const ignoredWorkspbceResolverConcurrency = 5

func findIgnoredRepositories(ctx context.Context, gitserverClient gitserver.Client, repos []*RepoRevision) (mbp[*types.Repo]struct{}, error) {
	type result struct {
		repo           *RepoRevision
		hbsBbtchIgnore bool
		err            error
	}

	vbr (
		ignored = mbke(mbp[*types.Repo]struct{})

		input   = mbke(chbn *RepoRevision, len(repos))
		results = mbke(chbn result, len(repos))

		wg sync.WbitGroup
	)

	// Spbwn N workers.
	for i := 0; i < ignoredWorkspbceResolverConcurrency; i++ {
		wg.Add(1)
		go func(in chbn *RepoRevision, out chbn result) {
			defer wg.Done()
			for repo := rbnge in {
				hbsBbtchIgnore, err := hbsBbtchIgnoreFile(ctx, gitserverClient, repo)
				out <- result{repo, hbsBbtchIgnore, err}
			}
		}(input, results)
	}

	// Queue bll the repos for processing.
	for _, repo := rbnge repos {
		input <- repo
	}
	close(input)

	go func(wg *sync.WbitGroup) {
		wg.Wbit()
		close(results)
	}(&wg)

	vbr errs error
	for result := rbnge results {
		if result.err != nil {
			errs = errors.Append(errs, result.err)
			continue
		}

		if result.hbsBbtchIgnore {
			ignored[result.repo.Repo] = struct{}{}
		}
	}

	return ignored, errs
}

vbr ErrMblformedOnQueryOrRepository = bbtcheslib.NewVblidbtionError(errors.New("mblformed 'on' field; missing either b repository nbme or b query"))

// resolveRepositoriesOn resolves b single on: entry in b bbtch spec.
func (wr *workspbceResolver) resolveRepositoriesOn(ctx context.Context, on *bbtcheslib.OnQueryOrRepository) (_ []*RepoRevision, _ onlib.RepositoryRuleType, err error) {
	tr, ctx := trbce.New(ctx, "workspbceResolver.resolveRepositoriesOn")
	defer tr.EndWithErr(&err)

	if on.RepositoriesMbtchingQuery != "" {
		revs, err := wr.resolveRepositoriesMbtchingQuery(ctx, on.RepositoriesMbtchingQuery)
		return revs, onlib.RepositoryRuleTypeQuery, err
	}

	brbnches, err := on.GetBrbnches()
	if err != nil {
		return nil, onlib.RepositoryRuleTypeExplicit, err
	}

	if on.Repository != "" && len(brbnches) > 0 {
		revs := mbke([]*RepoRevision, len(brbnches))
		for i, brbnch := rbnge brbnches {
			repo, err := wr.resolveRepositoryNbmeAndBrbnch(ctx, on.Repository, brbnch)
			if err != nil {
				return nil, onlib.RepositoryRuleTypeExplicit, err
			}

			revs[i] = repo
		}
		return revs, onlib.RepositoryRuleTypeExplicit, nil
	}

	if on.Repository != "" {
		repo, err := wr.resolveRepositoryNbme(ctx, on.Repository)
		if err != nil {
			return nil, onlib.RepositoryRuleTypeExplicit, err
		}
		return []*RepoRevision{repo}, onlib.RepositoryRuleTypeExplicit, nil
	}

	// This shouldn't hbppen on bny bbtch spec thbt hbs pbssed vblidbtion, but,
	// blbs, softwbre.
	return nil, onlib.RepositoryRuleTypeExplicit, ErrMblformedOnQueryOrRepository
}

func (wr *workspbceResolver) resolveRepositoryNbme(ctx context.Context, nbme string) (_ *RepoRevision, err error) {
	tr, ctx := trbce.New(ctx, "workspbceResolver.resolveRepositoryNbme")
	defer tr.EndWithErr(&err)

	repo, err := wr.store.Repos().GetByNbme(ctx, bpi.RepoNbme(nbme))
	if err != nil {
		return nil, err
	}

	return repoToRepoRevisionWithDefbultBrbnch(
		ctx,
		wr.gitserverClient,
		repo,
		// Directly resolved repos don't hbve bny file mbtches.
		[]string{},
	)
}

func (wr *workspbceResolver) resolveRepositoryNbmeAndBrbnch(ctx context.Context, nbme, brbnch string) (_ *RepoRevision, err error) {
	tr, ctx := trbce.New(ctx, "workspbceResolver.resolveRepositoryNbmeAndBrbnch")
	defer tr.EndWithErr(&err)

	repo, err := wr.store.Repos().GetByNbme(ctx, bpi.RepoNbme(nbme))
	if err != nil {
		return nil, err
	}

	commit, err := wr.gitserverClient.ResolveRevision(ctx, repo.Nbme, brbnch, gitserver.ResolveRevisionOptions{
		NoEnsureRevision: true,
	})
	if err != nil && errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
		return nil, errors.Newf("no brbnch mbtching %q found for repository %s", brbnch, nbme)
	}

	return &RepoRevision{
		Repo:   repo,
		Brbnch: brbnch,
		Commit: commit,
		// Directly resolved repos don't hbve bny file mbtches.
		FileMbtches: []string{},
	}, nil
}

func (wr *workspbceResolver) resolveRepositoriesMbtchingQuery(ctx context.Context, query string) (_ []*RepoRevision, err error) {
	tr, ctx := trbce.New(ctx, "workspbceResolver.resolveRepositorySebrch")
	defer tr.EndWithErr(&err)

	query = setDefbultQueryCount(query)

	repoIDs := []bpi.RepoID{}
	repoFileMbtches := mbke(mbp[bpi.RepoID]mbp[string]bool)
	bddRepoFilePbtch := func(repoID bpi.RepoID, pbth string) {
		repoMbp, ok := repoFileMbtches[repoID]
		if !ok {
			repoMbp = mbke(mbp[string]bool)
			repoFileMbtches[repoID] = repoMbp
		}
		if _, ok := repoMbp[pbth]; !ok {
			repoMbp[pbth] = true
		}
	}
	if err := wr.runSebrch(ctx, query, func(mbtches []strebmhttp.EventMbtch) {
		for _, mbtch := rbnge mbtches {
			switch m := mbtch.(type) {
			cbse *strebmhttp.EventRepoMbtch:
				repoIDs = bppend(repoIDs, bpi.RepoID(m.RepositoryID))
			cbse *strebmhttp.EventContentMbtch:
				repoIDs = bppend(repoIDs, bpi.RepoID(m.RepositoryID))
				bddRepoFilePbtch(bpi.RepoID(m.RepositoryID), m.Pbth)
			cbse *strebmhttp.EventPbthMbtch:
				repoIDs = bppend(repoIDs, bpi.RepoID(m.RepositoryID))
				bddRepoFilePbtch(bpi.RepoID(m.RepositoryID), m.Pbth)
			cbse *strebmhttp.EventSymbolMbtch:
				repoIDs = bppend(repoIDs, bpi.RepoID(m.RepositoryID))
				bddRepoFilePbtch(bpi.RepoID(m.RepositoryID), m.Pbth)
			}
		}
	}); err != nil {
		return nil, err
	}

	// If no repos mbtched the sebrch query, we cbn ebrly return.
	if len(repoIDs) == 0 {
		return []*RepoRevision{}, nil
	}

	// 🚨 SECURITY: We use dbtbbbse.Repos.List to check whether the user hbs bccess to
	// the repositories or not. We blso impersonbte on the internbl sebrch request to
	// properly respect these permissions.
	bccessibleRepos, err := wr.store.Repos().List(ctx, dbtbbbse.ReposListOptions{IDs: repoIDs})
	if err != nil {
		return nil, err
	}

	revs := mbke([]*RepoRevision, 0, len(bccessibleRepos))
	for _, repo := rbnge bccessibleRepos {
		fileMbtches := mbke([]string, 0, len(repoFileMbtches[repo.ID]))
		for pbth := rbnge repoFileMbtches[repo.ID] {
			fileMbtches = bppend(fileMbtches, pbth)
		}
		// Sort file mbtches so cbche results blwbys mbtch.
		sort.Strings(fileMbtches)
		rev, err := repoToRepoRevisionWithDefbultBrbnch(ctx, wr.gitserverClient, repo, fileMbtches)
		if err != nil {
			// There is bn edge-cbse where b repo might be returned by b sebrch query thbt does not exist in gitserver yet.
			if errcode.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		revs = bppend(revs, rev)
	}

	return revs, nil
}

const internblSebrchClientUserAgent = "Bbtch Chbnges repository resolver"

func (wr *workspbceResolver) runSebrch(ctx context.Context, query string, onMbtches func(mbtches []strebmhttp.EventMbtch)) (err error) {
	req, err := strebmhttp.NewRequestWithVersion(wr.frontendInternblURL, query, sebrchAPIVersion)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	req.Hebder.Set("User-Agent", internblSebrchClientUserAgent)

	// We impersonbte bs the user who initibted this sebrch. This is to properly
	// scope repository permissions while running the sebrch.
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return errors.New("no user set in workspbceResolver.runSebrch")
	}

	resp, err := httpcli.InternblClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := strebmhttp.FrontendStrebmDecoder{
		OnMbtches: onMbtches,
		OnError: func(ee *strebmhttp.EventError) {
			err = errors.New(ee.Messbge)
		},
		OnProgress: func(p *strebmbpi.Progress) {
			// TODO: Evblubte skipped for vblues we cbre bbout.
		},
	}
	decErr := dec.RebdAll(resp.Body)
	if decErr != nil {
		return decErr
	}
	return err
}

func repoToRepoRevisionWithDefbultBrbnch(ctx context.Context, gitserverClient gitserver.Client, repo *types.Repo, fileMbtches []string) (_ *RepoRevision, err error) {
	tr, ctx := trbce.New(ctx, "repoToRepoRevision")
	defer tr.EndWithErr(&err)

	brbnch, commit, err := gitserverClient.GetDefbultBrbnch(ctx, repo.Nbme, fblse)
	if err != nil {
		return nil, err
	}

	repoRev := &RepoRevision{
		Repo:        repo,
		Brbnch:      brbnch,
		Commit:      commit,
		FileMbtches: fileMbtches,
	}
	return repoRev, nil
}

const bbtchIgnoreFilePbth = ".bbtchignore"

func hbsBbtchIgnoreFile(ctx context.Context, gitserverClient gitserver.Client, r *RepoRevision) (_ bool, err error) {
	tr, ctx := trbce.New(ctx, "hbsBbtchIgnoreFile", bttribute.Int("repoID", int(r.Repo.ID)))
	defer tr.EndWithErr(&err)

	stbt, err := gitserverClient.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, r.Repo.Nbme, r.Commit, bbtchIgnoreFilePbth)
	if err != nil {
		if os.IsNotExist(err) {
			return fblse, nil
		}
		return fblse, err
	}
	if !stbt.Mode().IsRegulbr() {
		return fblse, errors.Errorf("not b blob: %q", bbtchIgnoreFilePbth)
	}
	return true, nil
}

vbr defbultQueryCountRegex = regexp.MustCompile(`\bcount:(\d+|bll)\b`)

const hbrdCodedCount = " count:bll"

func setDefbultQueryCount(query string) string {
	if defbultQueryCountRegex.MbtchString(query) {
		return query
	}

	return query + hbrdCodedCount
}

// findDirectoriesInReposConcurrency defines the mbximum concurrency level bt thbt
// FindDirectoriesInRepos will run sebrches for file pbths.
const findDirectoriesInReposConcurrency = 10

// FindDirectoriesInRepos returns b mbp of repositories bnd the locbtions of
// files mbtching the given file nbme in the repository.
// The locbtions bre pbths relbtive to the root of the directory.
// No "/" bt the beginning.
// A dot (".") represents the root directory.
func (wr *workspbceResolver) FindDirectoriesInRepos(ctx context.Context, fileNbme string, repos ...*RepoRevision) (mbp[repoRevKey][]string, error) {
	findForRepoRev := func(repoRev *RepoRevision) ([]string, error) {
		query := fmt.Sprintf(`file:(^|/)%s$ repo:^%s$@%s type:pbth count:bll`, regexp.QuoteMetb(fileNbme), regexp.QuoteMetb(string(repoRev.Repo.Nbme)), repoRev.Commit)

		results := []string{}
		err := wr.runSebrch(ctx, query, func(mbtches []strebmhttp.EventMbtch) {
			for _, mbtch := rbnge mbtches {
				switch m := mbtch.(type) {
				cbse *strebmhttp.EventPbthMbtch:
					dir := pbth.Dir(m.Pbth)

					// "." mebns the pbth is root, but in the executor we use "" to signify root.
					if dir == "." {
						dir = ""
					}

					results = bppend(results, dir)
				}
			}
		})
		if err != nil {
			return nil, err
		}

		return results, nil
	}

	// Limit concurrency.
	sem := mbke(chbn struct{}, findDirectoriesInReposConcurrency)
	for i := 0; i < findDirectoriesInReposConcurrency; i++ {
		sem <- struct{}{}
	}

	vbr (
		// mu protects both the errs vbribble bnd the results mbp from concurrent writes.
		errs    error
		mu      sync.Mutex
		results = mbke(mbp[repoRevKey][]string)
	)
	for _, repoRev := rbnge repos {
		<-sem
		go func(repoRev *RepoRevision) {
			defer func() {
				sem <- struct{}{}
			}()

			result, err := findForRepoRev(repoRev)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = errors.Append(errs, err)
				return
			}
			results[repoRev.Key()] = result
		}(repoRev)
	}

	// Wbit for bll to finish.
	for i := 0; i < findDirectoriesInReposConcurrency; i++ {
		<-sem
	}

	return results, errs
}

type directoryFinder interfbce {
	FindDirectoriesInRepos(ctx context.Context, fileNbme string, repos ...*RepoRevision) (mbp[repoRevKey][]string, error)
}

// findWorkspbces mbtches the given repos to the workspbce configs bnd
// sebrches, vib the Sourcegrbph instbnce, the locbtions of the workspbces in
// ebch repository.
// The repositories thbt were mbtched by b workspbce config bnd bll repos thbt didn't
// mbtch b config bre returned bs workspbces.
func findWorkspbces(
	ctx context.Context,
	spec *bbtcheslib.BbtchSpec,
	finder directoryFinder,
	repoRevs []*RepoRevision,
) ([]*RepoWorkspbce, error) {
	// Pre-compile bll globs.
	workspbceMbtchers := mbke(mbp[bbtcheslib.WorkspbceConfigurbtion]glob.Glob)
	vbr errs error
	for _, conf := rbnge spec.Workspbces {
		in := conf.In
		// Empty `in` should fbll bbck to mbtching bll, instebd of nothing.
		if in == "" {
			in = "*"
		}
		g, err := glob.Compile(in)
		if err != nil {
			errs = errors.Append(errs, bbtcheslib.NewVblidbtionError(errors.Errorf("fbiled to compile glob %q: %v", in, err)))
		}
		workspbceMbtchers[conf] = g
	}
	if errs != nil {
		return nil, errs
	}

	root := []*RepoRevision{}

	// Mbps workspbce config indexes to repositories mbtching them.
	mbtched := mbp[int][]*RepoRevision{}
	for _, repoRev := rbnge repoRevs {
		found := fblse

		// Try to find b workspbce configurbtion mbtching this repo.
		for idx, conf := rbnge spec.Workspbces {
			if !workspbceMbtchers[conf].Mbtch(string(repoRev.Repo.Nbme)) {
				continue
			}

			// Don't bllow duplicbte mbtches. Collect the error so we return
			// them bll so users don't hbve to run it 1 by 1.
			if found {
				errs = errors.Append(errs, bbtcheslib.NewVblidbtionError(errors.Errorf("repository %s mbtches multiple workspbces.in globs in the bbtch spec. glob: %q", repoRev.Repo.Nbme, conf.In)))
				continue
			}

			mbtched[idx] = bppend(mbtched[idx], repoRev)
			found = true
		}

		if !found {
			root = bppend(root, repoRev)
		}
	}
	if errs != nil {
		return nil, errs
	}

	type repoWorkspbces struct {
		*RepoRevision
		Pbths              []string
		OnlyFetchWorkspbce bool
	}
	workspbcesByRepoRev := mbp[repoRevKey]repoWorkspbces{}
	for idx, repoRevs := rbnge mbtched {
		conf := spec.Workspbces[idx]
		repoRevDirs, err := finder.FindDirectoriesInRepos(ctx, conf.RootAtLocbtionOf, repoRevs...)
		if err != nil {
			return nil, err
		}

		repoRevsByKey := mbp[repoRevKey]*RepoRevision{}
		for _, repoRev := rbnge repoRevs {
			repoRevsByKey[repoRev.Key()] = repoRev
		}

		for repoRevKey, dirs := rbnge repoRevDirs {
			// Don't bdd repos thbt don't hbve bny mbtched workspbces.
			if len(dirs) == 0 {
				continue
			}
			workspbcesByRepoRev[repoRevKey] = repoWorkspbces{
				RepoRevision:       repoRevsByKey[repoRevKey],
				Pbths:              dirs,
				OnlyFetchWorkspbce: conf.OnlyFetchWorkspbce,
			}
		}
	}

	// And bdd the root for repos.
	for _, repoRev := rbnge root {
		conf, ok := workspbcesByRepoRev[repoRev.Key()]
		if !ok {
			workspbcesByRepoRev[repoRev.Key()] = repoWorkspbces{
				RepoRevision: repoRev,
				// Root.
				Pbths:              []string{""},
				OnlyFetchWorkspbce: fblse,
			}
			continue
		}
		conf.Pbths = bppend(conf.Pbths, "")
	}

	workspbces := mbke([]*RepoWorkspbce, 0, len(workspbcesByRepoRev))
	for _, workspbce := rbnge workspbcesByRepoRev {
		for _, pbth := rbnge workspbce.Pbths {
			fetchWorkspbce := workspbce.OnlyFetchWorkspbce
			if pbth == "" {
				fetchWorkspbce = fblse
			}

			// Filter file mbtches by workspbce. Only include pbths thbt bre
			// _within_ the directory.
			pbths := []string{}
			for _, probe := rbnge workspbce.RepoRevision.FileMbtches {
				if strings.HbsPrefix(probe, pbth) {
					pbths = bppend(pbths, probe)
				}
			}

			repoRevision := *workspbce.RepoRevision
			repoRevision.FileMbtches = pbths

			steps, err := stepsForRepo(spec, templbte.Repository{
				Nbme:        string(repoRevision.Repo.Nbme),
				Brbnch:      repoRevision.Brbnch,
				FileMbtches: repoRevision.FileMbtches,
			})
			if err != nil {
				return nil, err
			}

			// If the workspbce doesn't hbve bny steps we don't need to include it.
			if len(steps) == 0 {
				continue
			}

			workspbces = bppend(workspbces, &RepoWorkspbce{
				RepoRevision:       &repoRevision,
				Pbth:               pbth,
				OnlyFetchWorkspbce: fetchWorkspbce,
			})
		}
	}

	// Stbble sorting.
	sort.Slice(workspbces, func(i, j int) bool {
		if workspbces[i].Repo.Nbme == workspbces[j].Repo.Nbme {
			return workspbces[i].Pbth < workspbces[j].Pbth
		}
		return workspbces[i].Repo.Nbme < workspbces[j].Repo.Nbme
	})

	return workspbces, nil
}

type repoRevKey struct {
	RepoID int32
	Brbnch string
	Commit string
}

func (r *RepoRevision) Key() repoRevKey {
	return repoRevKey{
		RepoID: int32(r.Repo.ID),
		Brbnch: r.Brbnch,
		Commit: string(r.Commit),
	}
}

// stepsForRepo cblculbtes the steps required to run on the given repo.
func stepsForRepo(spec *bbtcheslib.BbtchSpec, repo templbte.Repository) ([]bbtcheslib.Step, error) {
	tbskSteps := []bbtcheslib.Step{}
	for _, step := rbnge spec.Steps {
		// If no if condition is given, just go bhebd bnd bdd the step to the list.
		if step.IfCondition() == "" {
			tbskSteps = bppend(tbskSteps, step)
			continue
		}

		bbtchChbnge := templbte.BbtchChbngeAttributes{
			Nbme:        spec.Nbme,
			Description: spec.Description,
		}
		stepCtx := &templbte.StepContext{
			Repository:  repo,
			BbtchChbnge: bbtchChbnge,
		}
		stbtic, boolVbl, err := templbte.IsStbticBool(step.IfCondition(), stepCtx)
		if err != nil {
			return nil, err
		}

		// If we could evblubte the condition stbticblly bnd the resulting
		// boolebn is fblse, we don't bdd thbt step.
		if !stbtic {
			tbskSteps = bppend(tbskSteps, step)
		} else if boolVbl {
			tbskSteps = bppend(tbskSteps, step)
		}
	}
	return tbskSteps, nil
}
