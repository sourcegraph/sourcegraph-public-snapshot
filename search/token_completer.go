package search

import (
	"fmt"
	"log"
	"path/filepath"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sort"
	"strings"
	"sync"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/svc"
)

// A TokenCompletionConfig configures a CompleteToken call.
type TokenCompletionConfig struct {
	// DontResolveDefs determines whether defs are resolved. If
	// DontResolveDefs is true, then sourcegraph.Term tokens are not
	// attempted to be resolved to defs.
	DontResolveDefs bool

	DontResolveUnits bool

	// MaxPerType is the maxber of completions to return for each type
	// of completion (e.g., if set to 5, then Complete will return up
	// to 5 repo completions, up to 5 user completions, etc.). This is
	// a hint and the actual number could be more or less (if the
	// completer thinks it's likely you'd want more or less of what it
	// returns).
	//
	// If not set, MaxPerType defaults to 5.
	MaxPerType int

	// TODO(sqs): add Max to allow short-circuiting when we have 1
	// result (but only if it's a repo, for example).
}

// CompleteToken generates suggested completions for a token.
func CompleteToken(ctx context.Context, partial sourcegraph.Token, scope []sourcegraph.Token, conf TokenCompletionConfig) ([]sourcegraph.Token, error) {
	var comps []sourcegraph.Token
	defer func() {
		sort.Sort(repoOrgUserTermOrder(comps))
	}()
	c := NewTokenCompleter(partial, scope, conf)
	compc, errc := c.Complete(ctx)
loop:
	for {
		select {
		case comp := <-compc:
			comps = append(comps, comp)
		case err, ok := <-errc:
			if !ok {
				break loop
			}
			return comps, err
		}
	}
	for {
		select {
		case comp := <-compc:
			comps = append(comps, comp)
		default:
			return comps, nil
		}
	}
}

type repoOrgUserTermOrder []sourcegraph.Token

func (v repoOrgUserTermOrder) Len() int      { return len(v) }
func (v repoOrgUserTermOrder) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v repoOrgUserTermOrder) Less(i, j int) bool {
	ai, bi := v.sortKey(i)
	aj, bj := v.sortKey(j)
	return ai < aj || (ai == aj && bi > bj)
}
func (v repoOrgUserTermOrder) sortKey(i int) (int, int) {
	switch tok := v[i].(type) {
	case sourcegraph.RepoToken:
		var stars int
		if tok.Repo.GitHub != nil {
			stars = int(tok.Repo.GitHub.Stars)
		}
		return 0, stars
	case sourcegraph.UserToken:
		if tok.User != nil && tok.User.IsOrganization {
			return 1, 0
		}
		return 2, 0
	case sourcegraph.RevToken:
		return 3, 0
	case sourcegraph.UnitToken:
		return 4, 0
	case sourcegraph.FileToken:
		return 5, 0
	case sourcegraph.Term, sourcegraph.AnyToken:
		return 6, 0
	}
	panic(fmt.Sprintf("unknown token type %T", v[i]))
}

// A TokenCompleter generates suggested completions for a
// token. Usually you should just call the CompleteToken func instead
// of using a TokenCompleter (and its Complete method). Only use a
// TokenCompleter if you need the completions asynchronously.
type TokenCompleter struct {
	conf TokenCompletionConfig

	partial sourcegraph.Token   // the token to complete
	scope   []sourcegraph.Token // tokens in the query that are already resolved

	listOpts sourcegraph.ListOptions // with PerPage == x.MaxPerType

	compc chan sourcegraph.Token
	errc  chan error
}

func NewTokenCompleter(partial sourcegraph.Token, scope []sourcegraph.Token, conf TokenCompletionConfig) *TokenCompleter {
	if conf.MaxPerType <= 0 {
		conf.MaxPerType = 5 // default
	}
	return &TokenCompleter{
		conf: conf,

		partial: partial,
		scope:   scope,

		compc: make(chan sourcegraph.Token, 10),
		errc:  make(chan error),
	}
}

func (c *TokenCompleter) Complete(ctx context.Context) (<-chan sourcegraph.Token, <-chan error) {
	funcs := []func(context.Context) error{
		c.prefetchRepos,
		c.prefetchUsers,
		c.prefetchOrgs,
		c.completeRepo,
		c.completeRepoAny,
		c.completeRev,
		c.completeUnit,
		c.completeUnitAny,
		c.completeFile,
		c.completeUser,
		c.completeUserAny,
		c.completeDefNameAny,
	}

	// When all completion funcs return, close the errc. That's the
	// signal that completion has finished.
	done := make(chan struct{})
	go func() {
		for i := 0; i < len(funcs); i++ {
			<-done
		}
		close(c.errc)
	}()

	// TODO(perf): use context.WithCancel/context.WithDeadline
	// to provide better performance here.

	for _, f := range funcs {
		go func(f func(ctx context.Context) error) {
			if err := f(ctx); err != nil {
				c.errc <- err
			}
			done <- struct{}{}
		}(f)
	}

	return c.compc, c.errc
}

func (c *TokenCompleter) prefetchOnly() bool {
	return len(c.scope) == 0 && (c.partial == nil || c.partial.Token() == "")
}

// prefetchRepos fetches repos that the user is likely to want to use
// in the query.
func (c *TokenCompleter) prefetchRepos(ctx context.Context) error {
	if !c.prefetchOnly() {
		return nil
	}

	// TODO(sqs): fetch the user's repos plus the repos of the orgs
	// they're in, instead of just a random collection.
	repos, err := svc.Repos(ctx).List(ctx, &sourcegraph.RepoListOptions{
		BuiltOnly:   true,
		ListOptions: c.listOpts,
	})

	if err != nil {
		return err
	}
	for _, repo := range repos.Repos {
		c.compc <- sourcegraph.RepoToken{URI: repo.URI, Repo: repo}
	}
	return nil
}

// prefetchUsers fetches users that the user is likely to want to use
// in the query.
func (c *TokenCompleter) prefetchUsers(ctx context.Context) error {
	if !c.prefetchOnly() {
		return nil
	}

	addedSelf := false

	// TODO(sqs): fetch the user's co-collaborators plus the users of
	// the orgs they're in, instead of just a random collection.
	users, err := svc.Users(ctx).List(ctx, &sourcegraph.UsersListOptions{
		ListOptions: c.listOpts,
	})

	currentUserLogin, _ := auth.LoginFromContext(ctx)

	if err != nil {
		return err
	}
	for _, user := range users.Users {
		if user.Login == currentUserLogin {
			addedSelf = true
		}
		c.compc <- sourcegraph.UserToken{Login: user.Login, User: user}
	}

	if !addedSelf && currentUserLogin != "" {
		c.compc <- sourcegraph.UserToken{
			Login: currentUserLogin,
			User:  &sourcegraph.User{Login: currentUserLogin},
		}
	}

	return nil
}

// prefetchOrgs fetches orgs that the user is likely to want to use
// in the query.
func (c *TokenCompleter) prefetchOrgs(ctx context.Context) error {
	if !c.prefetchOnly() {
		return nil
	}

	login, hasLogin := auth.LoginFromContext(ctx)
	if !hasLogin {
		return nil
	}

	// TODO(sqs): fetch the user's orgs instead of just a random
	// collection.
	orgs, err := svc.Orgs(ctx).List(ctx, &sourcegraph.OrgsListOp{Member: sourcegraph.UserSpec{Login: login}, ListOptions: c.listOpts})

	if err != nil && grpc.Code(err) != codes.Unimplemented {
		return err
	}
	if orgs != nil {
		for _, org := range orgs.Orgs {
			c.compc <- sourcegraph.UserToken{Login: org.Login, User: &org.User}
		}
	}
	return nil
}

func (c *TokenCompleter) completeRepo(ctx context.Context) error {
	if c.prefetchOnly() {
		return nil
	}
	tok, ok := c.partial.(sourcegraph.RepoToken)
	if !ok {
		return nil
	}

	repos, err := svc.Repos(ctx).List(ctx, &sourcegraph.RepoListOptions{
		Query: tok.URI,

		ListOptions: c.listOpts,
	})

	if err != nil {
		return err
	}
	for _, repo := range repos.Repos {
		c.compc <- sourcegraph.RepoToken{URI: repo.URI, Repo: repo}
	}
	return nil
}

func (c *TokenCompleter) completeRepoAny(ctx context.Context) error {
	if c.prefetchOnly() {
		return nil
	}
	if len(c.scope) > 0 {
		return nil
	}
	tok, ok := c.partial.(sourcegraph.AnyToken)
	if !ok {
		return nil
	}

	opt := sourcegraph.RepoListOptions{
		Query: tok.Token(),
		// BuiltOnly:   true,
		ListOptions: c.listOpts,
		Sort:        "updated",
		Direction:   "desc",
	}

	need := c.conf.MaxPerType
	seen := map[string]struct{}{}
	sendRepos := func(repos []*sourcegraph.Repo) {
		for _, repo := range repos {
			if _, seen := seen[repo.URI]; seen {
				continue
			}
			seen[repo.URI] = struct{}{}

			need--
			opt.ListOptions.PerPage--
			c.compc <- sourcegraph.RepoToken{URI: repo.URI, Repo: repo}
		}
	}

	if login, ok := auth.LoginFromContext(ctx); ok {
		// Query the user's repos.
		opt.Owner = login
		opt.NoFork = false
		repos, err := svc.Repos(ctx).List(ctx, &opt)
		if err != nil {
			return err
		}
		sendRepos(repos.Repos)
		if need <= 0 {
			return nil
		}

		// Query the user's orgs' repos.
		orgs, err := svc.Orgs(ctx).List(ctx, &sourcegraph.OrgsListOp{Member: sourcegraph.UserSpec{Login: login}, ListOptions: sourcegraph.ListOptions{PerPage: 100}})

		if err != nil && grpc.Code(err) != codes.Unimplemented {
			return err
		}
		if orgs != nil {
			for _, org := range orgs.Orgs {
				opt.Owner = org.Login
				repos, err := svc.Repos(ctx).List(ctx, &opt)
				if err != nil {
					return err
				}
				sendRepos(repos.Repos)
				if need <= 0 {
					return nil
				}
			}
		}
	}

	// Query all repos.
	opt.Owner = ""
	opt.Type = "public"
	opt.NoFork = true
	repos, err := svc.Repos(ctx).List(ctx, &opt)
	if err != nil {
		return err
	}
	sendRepos(repos.Repos)
	if need <= 0 {
		return nil
	}

	return nil
}

func (c *TokenCompleter) completeRev(ctx context.Context) error {
	if c.prefetchOnly() {
		return nil
	}
	tok, ok := c.partial.(sourcegraph.RevToken)
	if !ok {
		return nil
	}

	repoSpec := c.getRepoSpec()
	if repoSpec == nil {
		return nil
	}

	repoRevSpec := sourcegraph.RepoRevSpec{
		RepoSpec: *repoSpec,
		Rev:      tok.Rev,
	}

	need := c.conf.MaxPerType
	seen := map[string]struct{}{}
	sendRev := func(wg *sync.WaitGroup, revToken sourcegraph.RevToken, fetchCommit bool) {
		if wg != nil {
			defer wg.Done()
		}

		if need <= 0 {
			return
		}

		if _, seen := seen[revToken.Rev]; seen {
			return
		}
		seen[revToken.Rev] = struct{}{}

		need--

		if fetchCommit {
			repoRevSpec := sourcegraph.RepoRevSpec{
				RepoSpec: *repoSpec,
				Rev:      string(revToken.Commit.ID),
				CommitID: string(revToken.Commit.ID),
			}
			commit, err := svc.Repos(ctx).GetCommit(ctx, &repoRevSpec)
			if err != nil {
				log.Println("GetCommit:", err)
			}
			if commit != nil {
				revToken.Commit = commit
			}
		}

		c.compc <- revToken
	}

	commit, err := svc.Repos(ctx).GetCommit(ctx, &repoRevSpec)
	if err != nil {
		log.Println("GetCommit:", err)
	}
	if commit != nil {
		sendRev(nil, sourcegraph.RevToken{Rev: tok.Rev, Commit: commit}, false)
	}

	q := strings.ToLower(tok.Rev)

	if need <= 0 {
		return nil
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		branches, err := svc.Repos(ctx).ListBranches(ctx, &sourcegraph.ReposListBranchesOp{Repo: repoRevSpec.RepoSpec, Opt: nil})
		if err != nil {
			log.Println("ListBranches", err)
		}
		for _, branch := range branches.Branches {
			if strings.Contains(strings.ToLower(branch.Name), q) {
				wg.Add(1)
				sendRev(&wg, sourcegraph.RevToken{Rev: branch.Name, Commit: &vcs.Commit{ID: branch.Head}}, true)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		tags, err := svc.Repos(ctx).ListTags(ctx, &sourcegraph.ReposListTagsOp{Repo: repoRevSpec.RepoSpec, Opt: nil})
		if err != nil {
			log.Println("ListTags", err)
		}
		for _, tag := range tags.Tags {
			if strings.Contains(strings.ToLower(tag.Name), q) {
				wg.Add(1)
				sendRev(&wg, sourcegraph.RevToken{Rev: tag.Name, Commit: &vcs.Commit{ID: tag.CommitID}}, true)
			}
		}
	}()

	wg.Wait()
	return nil
}

func (c *TokenCompleter) completeUnit(ctx context.Context) error {
	tok, ok := c.partial.(sourcegraph.UnitToken)
	if !ok {
		return nil
	}
	return c.doCompleteUnit(ctx, tok.Name)

}

func (c *TokenCompleter) completeUnitAny(ctx context.Context) error {
	// Don't complete if there's already a UnitToken.
	for _, tok := range c.scope {
		if _, ok := tok.(sourcegraph.UnitToken); ok {
			return nil
		}
	}

	tok, ok := c.partial.(sourcegraph.AnyToken)
	if !ok {
		return nil
	}
	if len(tok.Token()) < 3 {
		return nil
	}
	return c.doCompleteUnit(ctx, tok.Token())
}

func (c *TokenCompleter) doCompleteUnit(ctx context.Context, nameQuery string) error {
	if c.prefetchOnly() {
		return nil
	}
	if c.conf.DontResolveUnits {
		return nil
	}

	repoRevSpec, err := c.getRepoBuildRevSpec(ctx)
	if repoRevSpec == nil || err != nil {
		return err
	}

	units, err := svc.Units(ctx).List(ctx, &sourcegraph.UnitListOptions{
		RepoRevs:    []string{repoRevSpec.URI + "@" + repoRevSpec.CommitID},
		NameQuery:   nameQuery,
		ListOptions: c.listOpts,
	})

	if err != nil {
		return err
	}

	for _, unit := range units.Units {
		c.compc <- sourcegraph.UnitToken{UnitType: unit.UnitType, Name: unit.Unit, Unit: unit}
	}

	return nil
}

func (c *TokenCompleter) completeFile(ctx context.Context) error {
	if c.prefetchOnly() {
		return nil
	}
	tok, ok := c.partial.(sourcegraph.FileToken)
	if !ok {
		return nil
	}

	// Actually resolve since RepoTree.getFromVCS requires us to do so.
	repoRevSpec, err := c.getRepoRevSpecWithCommitID(ctx)
	if repoRevSpec == nil || err != nil {
		return nil
	}

	entrySpec := sourcegraph.TreeEntrySpec{
		RepoRev: *repoRevSpec,
		Path:    tok.Path,
	}

	need := c.conf.MaxPerType
	seen := map[string]struct{}{}
	sendFile := func(entry *sourcegraph.TreeEntry) {
		if _, seen := seen[entry.Name]; seen {
			return
		}
		seen[entry.Name] = struct{}{}

		need--

		entry.Contents = nil
		c.compc <- sourcegraph.FileToken{Path: entry.Name, Entry: entry.TreeEntry}

		for _, subentry := range entry.Entries {
			c.compc <- sourcegraph.FileToken{Path: subentry.Name, Entry: subentry}
		}
	}

	entry, err := svc.RepoTree(ctx).Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: entrySpec})
	if err != nil {
		log.Println("RepoTree.Get:", err)
	}
	if entry != nil {
		sendFile(entry)
	}

	if need <= 0 {
		return nil
	}

	parentDir := filepath.Dir(entrySpec.Path)
	if parentDir != filepath.Clean(entrySpec.Path) {
		entry, err := svc.RepoTree(ctx).Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: entrySpec})
		if err != nil {
			log.Println("RepoTree.Get:", err)
		}
		if entry != nil {
			sendFile(entry)
		}
	}

	return nil
}

type orgs []*sourcegraph.Org

func (v orgs) Len() int           { return len(v) }
func (v orgs) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v orgs) Less(i, j int) bool { return v[i].UID < v[j].UID }

func (c *TokenCompleter) completeUser(ctx context.Context) error {
	if c.prefetchOnly() {
		return nil
	}
	tok, ok := c.partial.(sourcegraph.UserToken)
	if !ok {
		return nil
	}

	users, err := svc.Users(ctx).List(ctx, &sourcegraph.UsersListOptions{
		Query:       tok.Login,
		ListOptions: c.listOpts,
	})

	if err != nil {
		return err
	}
	for _, user := range users.Users {
		c.compc <- sourcegraph.UserToken{Login: user.Login, User: user}
	}
	return nil
}

func (c *TokenCompleter) completeUserAny(ctx context.Context) error {
	if c.prefetchOnly() {
		return nil
	}
	if len(c.scope) > 0 {
		return nil
	}
	tok, ok := c.partial.(sourcegraph.AnyToken)
	if !ok {
		return nil
	}

	users, err := svc.Users(ctx).List(ctx, &sourcegraph.UsersListOptions{
		Query:       tok.Token(),
		ListOptions: c.listOpts,
	})

	if err != nil {
		return err
	}
	for _, user := range users.Users {
		c.compc <- sourcegraph.UserToken{Login: user.Login, User: user}
	}
	return nil
}

func (c *TokenCompleter) completeDefNameAny(ctx context.Context) error {
	if c.conf.DontResolveDefs {
		return nil
	}

	if c.prefetchOnly() {
		return nil
	}
	_, ok := c.partial.(sourcegraph.AnyToken)
	if !ok {
		_, ok = c.partial.(sourcegraph.Term)
	}
	if !ok {
		return nil
	}

	plan, err := NewPlan(ctx, c.scope)
	if err != nil {
		return err
	}

	if plan.Defs == nil {
		return nil
	}

	plan.Defs.Query = c.partial.Token()
	plan.Defs.ListOptions = c.listOpts

	defs, err := svc.Defs(ctx).List(ctx, plan.Defs)
	if err != nil {
		return err
	}

	seen := map[string]struct{}{}
	for _, def := range defs.Defs {
		nameLower := strings.ToLower(def.Name)
		if _, seen := seen[nameLower]; seen {
			continue
		}
		seen[nameLower] = struct{}{}

		c.compc <- sourcegraph.Term(def.Name)
	}
	return nil
}

// getRepoSpec generates the effective RepoSpec if there is a
// RepoToken.
func (c *TokenCompleter) getRepoSpec() *sourcegraph.RepoSpec {
	for _, tok := range c.scope {
		if tok, ok := tok.(sourcegraph.RepoToken); ok {
			return &sourcegraph.RepoSpec{URI: tok.URI}
		}
	}
	return nil
}

func (c *TokenCompleter) getRepoRevSpec(ctx context.Context) (*sourcegraph.RepoRevSpec, error) {
	repoSpec := c.getRepoSpec()
	if repoSpec == nil {
		return nil, nil
	}

	repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: *repoSpec}
	for _, tok := range c.scope {
		if tok, ok := tok.(sourcegraph.RevToken); ok {
			repoRevSpec.Rev = tok.Rev
			if tok.Commit != nil {
				repoRevSpec.CommitID = string(tok.Commit.ID)
			}
			break
		}
	}

	if repoRevSpec.Rev == "" {
		// Look up repo's default branch.
		repo, err := svc.Repos(ctx).Get(ctx, repoSpec)
		if err != nil {
			return nil, err
		}
		repoRevSpec.Rev = repo.DefaultBranch
	}
	return &repoRevSpec, nil
}

func (c *TokenCompleter) getRepoRevSpecWithCommitID(ctx context.Context) (*sourcegraph.RepoRevSpec, error) {
	repoRevSpec, err := c.getRepoRevSpec(ctx)
	if repoRevSpec == nil || err != nil {
		return repoRevSpec, err
	}
	if repoRevSpec.CommitID == "" {
		commit, err := svc.Repos(ctx).GetCommit(ctx, &*repoRevSpec)
		if err != nil {
			return nil, err
		}
		repoRevSpec.CommitID = string(commit.ID)
	}
	return repoRevSpec, nil
}

func (c *TokenCompleter) getRepoBuildRevSpec(ctx context.Context) (*sourcegraph.RepoRevSpec, error) {
	repoRevSpec, err := c.getRepoRevSpec(ctx)
	if repoRevSpec == nil || err != nil {
		return repoRevSpec, err
	}
	buildInfo, err := svc.Builds(ctx).GetRepoBuildInfo(ctx, &sourcegraph.BuildsGetRepoBuildInfoOp{Repo: *repoRevSpec})
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	if buildInfo.LastSuccessful == nil {
		return nil, nil
	}
	repoRevSpec.CommitID = buildInfo.LastSuccessful.CommitID
	return repoRevSpec, nil
}
