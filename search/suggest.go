package search

import (
	"log"
	"math/rand"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/svc"
)

type suggestions []*sourcegraph.Suggestion

func (v suggestions) Len() int      { return len(v) }
func (v suggestions) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v suggestions) Less(i, j int) bool {
	return sourcegraph.PBTokens(v[i].Query).RawQueryString() < sourcegraph.PBTokens(v[j].Query).RawQueryString()
}

// SuggestionConfig specifies configuration to use when
// suggesting queries. Pass it to Suggest to suggest queries.
type SuggestionConfig struct {
	// MaxPerType is the number of suggestions to return for each type
	// of suggestion (e.g., if set to 5, then Suggest will return up
	// to 5 repo suggestions, up to 5 repo owner suggestions,
	// etc.). This is a hint and the actual number could be more or
	// less (if the suggester thinks it's likely you'd want more or
	// less of what it returns).
	//
	// If not set, MaxPerType defaults to 5.
	MaxPerType int
}

// Suggest generates suggested query suggestions from a tokenized
// query.
func Suggest(ctx context.Context, resolved []sourcegraph.Token, conf SuggestionConfig) ([]*sourcegraph.Suggestion, error) {
	var suggs []*sourcegraph.Suggestion
	c := NewSuggester(resolved, conf)
	compc, errc := c.Suggest(ctx)
loop:
	for {
		select {
		case comp := <-compc:
			suggs = append(suggs, comp)
		case err, ok := <-errc:
			if !ok {
				break loop
			}
			return suggs, err
		}
	}
	for {
		select {
		case comp := <-compc:
			suggs = append(suggs, comp)
		default:
			return suggs, nil
		}
	}
}

// A Suggester generates suggested query suggestions from a tokenized
// query. Usually you should just call the Suggest func instead of
// using a Suggester (and its Suggest method). Only use a Suggester
// if you need the suggestions asynchronously.
type Suggester struct {
	conf SuggestionConfig

	orig []sourcegraph.Token

	compc chan *sourcegraph.Suggestion
	errc  chan error
	wg    *sync.WaitGroup

	listOpts sourcegraph.ListOptions // with PerPage == x.MaxPerType
}

func NewSuggester(resolved []sourcegraph.Token, conf SuggestionConfig) *Suggester {
	if conf.MaxPerType <= 0 {
		conf.MaxPerType = 5 // default
	}
	return &Suggester{
		conf:     conf,
		orig:     resolved,
		listOpts: sourcegraph.ListOptions{PerPage: int32(conf.MaxPerType)},
	}
}

func (c *Suggester) Suggest(ctx context.Context) (<-chan *sourcegraph.Suggestion, <-chan error) {
	c.compc = make(chan *sourcegraph.Suggestion, 10)
	c.errc = make(chan error)
	c.wg = new(sync.WaitGroup)

	if len(c.orig) == 0 {
		if login, ok := auth.LoginFromContext(ctx); ok {
			c.do(ctx, func(ctx context.Context) error {
				prefix := []sourcegraph.Token{
					sourcegraph.UserToken{Login: login, User: &sourcegraph.User{Login: login}},
				}
				return c.suggestDefSearchInCodeOfOwner(ctx, prefix, login)
			})
		}
		c.do(ctx, c.suggestDefSearchInOrgs)
	} else {
		for _, tok := range c.orig {
			switch tok := tok.(type) {
			case sourcegraph.RepoToken:
				c.do(ctx, func(ctx context.Context) error {
					return c.suggestDefSearchInRepo(ctx, []sourcegraph.Tokens{{tok}}, tok.Repo)
				})
			}
		}
	}

	// When all suggestion funcs return, close the errc. That's the
	// signal that suggestion has finished.
	go func() {
		c.wg.Wait()
		close(c.errc)
	}()

	return c.compc, c.errc
}

func (c *Suggester) do(ctx context.Context, f func(ctx context.Context) error) {
	// TODO(sqs): can probably simplify this after it was refactored to use context.Context
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		if err := f(ctx); err != nil {
			c.errc <- err
		}
	}()
}

func (c *Suggester) suggestDefSearchInCodeOfOwner(ctx context.Context, prefix []sourcegraph.Token, ownerLogin string) error {
	repos, err := svc.Repos(ctx).List(ctx, &sourcegraph.RepoListOptions{
		Sort:        "updated_at",
		Direction:   "desc",
		Owner:       ownerLogin,
		BuiltOnly:   true,
		NoFork:      true,
		ListOptions: c.listOpts,
	})

	if err != nil {
		return err
	}

	for _, repo_ := range repos.Repos {
		repo := repo_
		prefixes := []sourcegraph.Tokens{prefix, []sourcegraph.Token{sourcegraph.RepoToken{URI: repo.URI, Repo: repo}}}
		c.do(ctx, func(ctx context.Context) error { return c.suggestDefSearchInRepo(ctx, prefixes, repo) })
	}

	return nil
}

func (c *Suggester) suggestDefSearchInRepo(ctx context.Context, prefixes []sourcegraph.Tokens, repo *sourcegraph.Repo) error {
	repoRevSpec := sourcegraph.RepoRevSpec{
		RepoSpec: repo.RepoSpec(),
		Rev:      repo.DefaultBranch,
	}
	buildInfo, err := svc.Builds(ctx).GetRepoBuildInfo(ctx, &sourcegraph.BuildsGetRepoBuildInfoOp{Repo: repoRevSpec})
	if err != nil || buildInfo.LastSuccessful == nil {
		// Skip.
		log.Printf("GetBuild(%v): error: %s", repoRevSpec, err)
		return nil
	}

	commitID := buildInfo.LastSuccessful.CommitID
	defs, err := svc.Defs(ctx).List(ctx, &sourcegraph.DefListOptions{
		RepoRevs:    []string{repo.URI + "@" + commitID},
		Exported:    true,
		Nonlocal:    true,
		ListOptions: c.listOpts,
	})

	if err != nil {
		return err
	}

	if len(defs.Defs) > c.listOpts.PerPageOrDefault() {
		defs.Defs = defs.Defs[:c.listOpts.PerPageOrDefault()]
	}

	for i, def := range defs.Defs {
		prefix := prefixes[i%len(prefixes)]
		var q []sourcegraph.Token
		q = append(q, prefix...)

		// Add source unit token to some.
		if rand.Int()%2 == 0 && def.UnitType != "" && def.Unit != "" && def.Unit != "." {
			repoPrefixOnly := true
			for _, tok := range q {
				if _, ok := tok.(sourcegraph.RepoToken); !ok {
					repoPrefixOnly = false
					break
				}
			}
			if repoPrefixOnly {
				q = append(q, sourcegraph.UnitToken{Name: def.Unit})
			}
		}

		q = append(q, sourcegraph.Term(def.Name))

		c.compc <- &sourcegraph.Suggestion{
			Query:       sourcegraph.PBTokensWrap(q),
			Description: Describe(q),
		}
	}
	return nil
}

var builtinOrgs = []*sourcegraph.Org{
	&sourcegraph.Org{User: sourcegraph.User{Login: "docker", IsOrganization: true}},
	&sourcegraph.Org{User: sourcegraph.User{Login: "aws", IsOrganization: true}},
	&sourcegraph.Org{User: sourcegraph.User{Login: "JodaOrg", IsOrganization: true}},
}

func (c *Suggester) suggestDefSearchInOrgs(ctx context.Context) error {
	var orgs []*sourcegraph.Org
	if login, ok := auth.LoginFromContext(ctx); ok {
		orgs2, err := svc.Orgs(ctx).List(ctx, &sourcegraph.OrgsListOp{Member: sourcegraph.UserSpec{Login: login}, ListOptions: c.listOpts})
		if err != nil && grpc.Code(err) != codes.Unimplemented {
			return err
		}
		if orgs2 != nil {
			orgs = append(orgs, orgs2.Orgs...)
		}
	}

	orgs = append(orgs, builtinOrgs...)

	for _, org_ := range orgs {
		org := org_
		prefix := []sourcegraph.Token{
			sourcegraph.UserToken{Login: org.Login, User: &org.User},
		}
		c.do(ctx, func(ctx context.Context) error { return c.suggestDefSearchInCodeOfOwner(ctx, prefix, org.Login) })
	}

	return nil
}

// queryWithPrepended creates a new query by prepending toks to a copy
// of the existing query.
func (c *Suggester) queryWithPrepended(toks ...sourcegraph.Token) []sourcegraph.Token {
	qtoks := append([]sourcegraph.Token{}, toks...)
	qtoks = append(qtoks, c.orig...)
	return qtoks
}

// queryByReplacing creates a new query by replacing the old token
// with the new token in a copy of the existing query.
func (c *Suggester) queryByReplacing(old, new sourcegraph.Token) []sourcegraph.Token {
	return c.queryWithReplacements(map[sourcegraph.Token]sourcegraph.Token{old: new})
}

// queryWithReplacements creates a new query by replacing the tokens in the map keys
// with their corresponding map values.
func (c *Suggester) queryWithReplacements(replace map[sourcegraph.Token]sourcegraph.Token) []sourcegraph.Token {
	qtoks := append([]sourcegraph.Token{}, c.orig...)
	for i, tok := range qtoks {
		if new, present := replace[tok]; present {
			qtoks[i] = new
		}
	}
	return qtoks
}
