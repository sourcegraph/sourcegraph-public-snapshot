package app

import (
	"bytes"
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"strings"

	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/appconf"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveSearchResults(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.SearchOptions
	err := schemautil.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	// We use a simple element-wise space check on the string to
	// tell if it's a query for "some/repo someDef" or not.
	inRepo := len(strings.Split(strings.TrimSpace(opt.Query), " ")) > 1

	// If searching in repo or if global search is enabled then it's
	// okay to search for everything -- otherwise we must disable them
	// all (defs, people, and tree search are too slow).
	if inRepo || !appconf.Flags.DisableGlobalSearch {
		opt.Defs = true
		opt.Repos = true
		opt.People = true

		opt.Tree = !appconf.Flags.DisableRepoTreeSearch
	}

	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	if opt.PerPage == 0 {
		opt.PerPage = 15
	}
	if opt.Query == "" {
		return serveSearchForm(w, r)
	}

	results, err := apiclient.Search.Search(ctx, &opt)
	if err != nil {
		return err
	}
	addPopoversToTextSearchResults(results.Tree)

	return tmpl.Exec(r, w, "search/results.html", http.StatusOK, nil, &struct {
		SearchOptions *sourcegraph.SearchOptions
		SearchResults *sourcegraph.SearchResults

		tmpl.Common
	}{
		SearchOptions: &opt,
		SearchResults: results,
	})
}

func addPopoversToTextSearchResults(res []*sourcegraph.RepoTreeSearchResult) {
	// HACK
	for _, res := range res {
		res.Match = bytes.Replace(res.Match, []byte("class=\""), []byte("class=\"defn-popover "), -1)
	}
}

func serveSearchForm(w http.ResponseWriter, r *http.Request) error {
	return tmpl.Exec(r, w, "search/form.html", http.StatusOK, nil, &struct{ tmpl.Common }{})
}

// searchForm holds info for the top navbar search form. It represents
// the global search form or a per-repo search form.
type searchForm struct {
	// ActionURL is the <form action=""> value for the search form.
	ActionURL *url.URL

	// ActionImplicitQuery is the query prefix that the form action
	// (i.e., the ActionURL endpoint) prepends to any query it is
	// passed. The ?q= URL querystring parameter should not contain
	// this value, but the search field should display it.
	//
	// E.g., the repo search URL should be
	// /example.com/myrepo/.search?q=foo, NOT
	// /example.com/myrepo/.search?q=example.com%2Fmyrepo%20foo (which
	// is redundant and confusing).
	ActionImplicitQueryPrefix sourcegraph.Tokens

	// InputValue is the <input name="q" value=""> value for the
	// search field.
	InputValue string

	// ResolvedTokens are the tokens that are parsed and resolved from
	// the query string in InputValue (so that the search field can
	// show rich tokens).
	ResolvedTokens sourcegraph.Tokens

	// PJAXContainer is the "#id" of the PJAX container element that
	// should be replaced with the search results.
	PJAXContainer string

	// ResolveErrors are token resolution errors.
	ResolveErrors []sourcegraph.TokenError
}

// searchFormActionURL takes the global template data object (i.e.,
// ".") and returns information necessary to generate the top navbar
// search form.
//
// There are 3 possible kinds of search forms:
//
// 1. The page is for a specific repo revision (i.e., the template
// data contains a RepoBuildInfo). If there is a successful build
// associated with the repo revision (using Builds.GetRepoBuildInfo, so not
// always an exact match), the search form will search within that
// commit.
//
// 2. The page is for a specific repo but no revision is specified
// (i.e., the template contains a Repo but not RepoBuildInfo). The
// search form will search within the repo's default branch.
//
// 3. The page is not a repo page, or any page underneath a repo
// page. The global search endpoint is used.
func searchFormInfo(tmplData interface{}) (*searchForm, error) {
	var err error
	sf := &searchForm{}

	var repo *sourcegraph.Repo
	if repoField, ok := getStructField(tmplData, "Repo"); ok {
		repo, ok = repoField.(*sourcegraph.Repo)
	}

	// Treat a non-enabled repo as though it didn't exist.
	if settingsField, ok := getStructField(tmplData, "RepoConfig"); ok {
		if settings, ok := settingsField.(*sourcegraph.RepoConfig); ok {
			if settings != nil && !settings.Enabled {
				repo = nil
			}
		}
	}

	if repo != nil {
		// Repo search.
		sf.PJAXContainer = "#repo-pjax-container"

		// Now see if we're viewing a specific commit.
		sf.ActionImplicitQueryPrefix = append(sf.ActionImplicitQueryPrefix, sourcegraph.RepoToken{URI: repo.URI, Repo: repo})

		var revToken sourcegraph.RevToken

		if routeVars, ok := getStructField(tmplData, "CurrentRouteVars"); ok {
			if routeVars, ok := routeVars.(map[string]string); ok {
				repoRevSpec, err := sourcegraph.UnmarshalRepoRevSpec(routeVars)
				if err != nil {
					return nil, err
				}
				revToken.Rev = repoRevSpec.Rev
			}
		}

		if buildInfoField, ok := getStructField(tmplData, "RepoBuildInfo"); ok {
			if repoBuildInfo, ok := buildInfoField.(*sourcegraph.RepoBuildInfo); ok {
				if repoBuildInfo != nil && repoBuildInfo.LastSuccessful != nil {
					// Search a specific commit.
					revToken.Commit = repoBuildInfo.LastSuccessfulCommit
				}
				// NOTE: When we add more things that can be searched other
				// than just defs, we may want to relax this so that even if
				// there's no successful build, users can still search other
				// objects that exist at that commit (e.g., files).
			}
		}

		if revToken.Rev != "" {
			sf.ActionImplicitQueryPrefix = append(sf.ActionImplicitQueryPrefix, revToken)
		}

		sf.ActionURL, err = router.Rel.URLToRepoSubrouteRev(router.RepoSearch, repo.URI, revToken.Rev)
		if err != nil {
			return nil, err
		}
	} else {
		if appconf.Flags.DisableGlobalSearch {
			return nil, nil
		}

		// Global search.
		sf.ActionURL = router.Rel.URLTo(router.SearchResults)
		sf.PJAXContainer = "#pjax-container"
	}

	if len(sf.ActionImplicitQueryPrefix) > 0 {
		sf.InputValue = sourcegraph.Join(sf.ActionImplicitQueryPrefix).Text
		sf.ResolvedTokens = append(sf.ResolvedTokens, sf.ActionImplicitQueryPrefix...)
	}

	if searchOpt, ok := getStructField(tmplData, "SearchOptions"); ok {
		if searchOpt, ok := searchOpt.(*sourcegraph.SearchOptions); ok && searchOpt != nil {
			sf.InputValue += " " + searchOpt.Query
		}
	}
	sf.InputValue = strings.TrimSpace(sf.InputValue)

	if res, ok := getStructField(tmplData, "SearchResults"); ok {
		if res, ok := res.(*sourcegraph.SearchResults); ok && res != nil {
			// Overwrite the previous value of sf.ResolvedTokens set
			// above because ResolvedTokens should always be resolved
			// from the full query, not just the part after the
			// implicit query prefix.
			sf.ResolvedTokens = sourcegraph.PBTokens(res.ResolvedTokens)
			sf.ResolveErrors = res.ResolveErrors
		}
	}
	sf.InputValue = strings.TrimSpace(sf.InputValue)

	return sf, nil
}

func showSearchForm(ctx context.Context, query url.Values) bool {
	if _, ok := query["EnableSearch"]; ok {
		return true
	}

	if appconf.Flags.DisableSearch {
		return false
	}

	if appconf.Flags.CustomNavLayout != "" {
		return false
	}

	cl := sourcegraph.NewClientFromContext(ctx)
	config, err := cl.Meta.Config(ctx, &pbtypes.Void{})
	if err != nil {
		return true
	}

	if config.AllowAnonymousReaders {
		return true
	}

	u := handlerutil.UserFromContext(ctx)
	if u != nil && u.UID != 0 {
		return true
	}

	return false
}
