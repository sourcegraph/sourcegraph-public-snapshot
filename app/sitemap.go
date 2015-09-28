package app

import (
	"net/http"
	"time"

	"github.com/sourcegraph/sitemap"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func serveSitemapIndex(w http.ResponseWriter, r *http.Request) error {
	start := time.Now()
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	// get top repos
	var si sitemap.Index
	const (
		maxPages = 50
		maxRepos = 350
	)
	for page := 1; page < maxPages && time.Since(start) < time.Second*20 || len(si.Sitemaps) < maxRepos; page++ {
		repos, err := apiclient.Repos.List(ctx, &sourcegraph.RepoListOptions{
			BuiltOnly: true,
			NoFork:    true,
			Sort:      "updated",
			Type:      "public",
			Direction: "desc",
			ListOptions: sourcegraph.ListOptions{
				Page:    int32(page),
				PerPage: 1000,
			},
		})

		if err != nil {
			return err
		}
		if len(repos.Repos) == 0 {
			break
		}

		// Only take Go, Java, and Python repos for now, since we support those best.
		for _, repo := range repos.Repos {
			if repo.Language == "Java" || repo.Language == "Go" || repo.Language == "Python" {
				var lastMod *time.Time
				if repo.UpdatedAt.Time().IsZero() {
					tmp := repo.UpdatedAt.Time()
					lastMod = &tmp
				}
				si.Sitemaps = append(si.Sitemaps, sitemap.Sitemap{
					Loc:     conf.AppURL(ctx).ResolveReference(router.Rel.URLToRepoSubroute(router.RepoSitemap, repo.URI)).String(),
					LastMod: lastMod,
				})
			}
		}
	}

	// Truncate to sitemaps limit.
	if len(si.Sitemaps) > sitemap.MaxSitemaps {
		si.Sitemaps = si.Sitemaps[:sitemap.MaxSitemaps]
	}

	siXML, err := sitemap.MarshalIndex(&si)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Cache-Control", "max-age=900")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write(siXML)
	return nil
}

func serveRepoSitemap(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	rc, vc, err := handlerutil.GetRepoAndRevCommon(r, nil)
	if err != nil {
		return err
	}

	var sm sitemap.URLSet

	bc, err := handlerutil.GetRepoBuildCommon(r, rc, vc, nil)
	if err != nil {
		return err
	}
	vc.RepoRevSpec = bc.BestRevSpec // Remove after getRepo refactor.

	var lastMod *time.Time
	if bc.RepoBuildInfo != nil && bc.RepoBuildInfo.LastSuccessful != nil {
		tmp := bc.RepoBuildInfo.LastSuccessful.EndedAt.Time()
		lastMod = &tmp
	}

	const (
		repoPriority  = 0.9
		defHiPriority = 0.7
		defLoPriority = 0.3
	)

	// Default change freq is Weekly, but if the repo was rebuilt in
	// the last week, use Daily.
	chgFreq := sitemap.Weekly
	if lastMod != nil && time.Since(*lastMod) < time.Hour*24*7 {
		chgFreq = sitemap.Daily
	}

	// Add repo main page.
	sm.URLs = append(sm.URLs, sitemap.URL{
		Loc:        conf.AppURL(ctx).ResolveReference(router.Rel.URLToRepo(rc.Repo.URI)).String(),
		ChangeFreq: chgFreq,
		Priority:   repoPriority,
	})

	// Add defs.
	seenDefs := map[graph.DefKey]bool{}
	defs, err := apiclient.Defs.List(ctx, &sourcegraph.DefListOptions{
		RepoRevs:    []string{rc.Repo.URI + "@" + vc.RepoRevSpec.CommitID},
		Exported:    true,
		IncludeTest: false,
		Doc:         true,
		ListOptions: sourcegraph.ListOptions{PerPage: sitemap.MaxURLs - 1},
	})
	if err != nil {
		return err
	}

	for _, def := range defs.Defs {
		// TODO(sqs): when can defs be repeated? probably only
		// when srclib store pagination is faulty.
		if seenDefs[def.DefKey] {
			continue
		}
		seenDefs[def.DefKey] = true

		if !defRobotsIndex(rc.Repo, def) {
			continue
		}

		var pri float64
		if len(def.Docs) == 0 {
			pri = defLoPriority
		} else {
			pri = defHiPriority
		}

		url := conf.AppURL(ctx).ResolveReference(router.Rel.URLToDef(def.DefKey)).String()
		if len(url) > 1000 {
			// Google rejects long URLs >2000 chars, but let's limit
			// them to 1000 just to be safe/sane.
			continue
		}

		sm.URLs = append(sm.URLs, sitemap.URL{
			Loc:        url,
			LastMod:    lastMod,
			ChangeFreq: chgFreq,
			Priority:   pri,
		})
	}

	// Truncate to sitemaps limit.
	if len(sm.URLs) > sitemap.MaxURLs {
		sm.URLs = sm.URLs[:sitemap.MaxURLs]
	}

	sitemapXML, err := sitemap.Marshal(&sm)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Cache-Control", "max-age=900")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Write(sitemapXML)
	return nil
}
